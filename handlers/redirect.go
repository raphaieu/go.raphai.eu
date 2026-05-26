package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"go.raphai.eu/auth"
	"go.raphai.eu/models"
	"go.raphai.eu/store"
)

type Handler struct {
	Store        *store.Store
	Sessions     *auth.SessionStore
	Tmpl         *template.Template
	AdminUser    string
	AdminPass    string
}

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	log.Printf("root: unexpected path %q", r.URL.Path)
	http.NotFound(w, r)
}

func (h *Handler) render(w http.ResponseWriter, page, title string, data interface{}) {
	var buf bytes.Buffer
	if err := h.Tmpl.ExecuteTemplate(&buf, page, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Title":   title,
		"Content": template.HTML(buf.String()),
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	redir, err := h.Store.GetRedirectBySlug(slug)
	if err != nil || !redir.IsActive {
		http.NotFound(w, r)
		return
	}

	targetURL := buildURL(redir)

	h.Store.IncrementVisitCount(redir.ID)

	if redir.RedirectType == models.RedirectTypeDirect301 {
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = h.Tmpl.ExecuteTemplate(w, "redirect.html", map[string]string{
		"TargetURL": targetURL,
		"Slug":      redir.Slug,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildURL(r *models.Redirect) string {
	u, err := url.Parse(r.TargetURL)
	if err != nil {
		return r.TargetURL
	}
	q := u.Query()

	set := func(key, value string) {
		if value != "" {
			q.Set(key, value)
		}
	}

	set("utm_source", r.UTMSource)
	set("utm_medium", r.UTMMedium)
	set("utm_campaign", r.UTMCampaign)
	set("utm_term", r.UTMTerm)
	set("utm_content", r.UTMContent)

	u.RawQuery = q.Encode()
	return u.String()
}
