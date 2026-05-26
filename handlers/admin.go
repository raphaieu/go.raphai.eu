package handlers

import (
	"net/http"
	"strconv"

	"go.raphai.eu/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.Tmpl.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != h.AdminUser {
		h.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Credenciais inválidas",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(h.AdminPass), []byte(password)); err != nil {
		h.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Credenciais inválidas",
		})
		return
	}

	token := h.Sessions.Create(username)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.Sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	redirects, err := h.Store.ListRedirects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "dashboard.html", "Dashboard - go.raphai.eu", map[string]interface{}{
		"Redirects": redirects,
		"Success":   r.URL.Query().Get("success"),
	})
}

func (h *Handler) AdminNew(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.render(w, "form.html", "Novo Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": &models.Redirect{RedirectType: models.RedirectTypeMetaRefresh, IsActive: true},
			"IsNew":    true,
		})
		return
	}

	redir := parseForm(r)

	if redir.Slug == "" || redir.TargetURL == "" {
		h.render(w, "form.html", "Novo Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": redir,
			"IsNew":    true,
			"Error":    "Slug e URL de destino são obrigatórios",
		})
		return
	}

	exists, _ := h.Store.SlugExists(redir.Slug, 0)
	if exists {
		h.render(w, "form.html", "Novo Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": redir,
			"IsNew":    true,
			"Error":    "Este slug já existe",
		})
		return
	}

	_, err := h.Store.CreateRedirect(redir)
	if err != nil {
		h.render(w, "form.html", "Novo Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": redir,
			"IsNew":    true,
			"Error":    "Erro ao criar: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/admin?success=Redirect+criado+com+sucesso", http.StatusSeeOther)
}

func (h *Handler) AdminEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	redir, err := h.Store.GetRedirectByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == "GET" {
		h.render(w, "form.html", "Editar Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": redir,
			"IsNew":    false,
		})
		return
	}

	updated := parseForm(r)
	updated.ID = id

	if updated.Slug == "" || updated.TargetURL == "" {
		h.render(w, "form.html", "Editar Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": updated,
			"IsNew":    false,
			"Error":    "Slug e URL de destino são obrigatórios",
		})
		return
	}

	exists, _ := h.Store.SlugExists(updated.Slug, id)
	if exists {
		h.render(w, "form.html", "Editar Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": updated,
			"IsNew":    false,
			"Error":    "Este slug já está em uso",
		})
		return
	}

	if err := h.Store.UpdateRedirect(updated); err != nil {
		h.render(w, "form.html", "Editar Redirect - go.raphai.eu", map[string]interface{}{
			"Redirect": updated,
			"IsNew":    false,
			"Error":    "Erro ao atualizar: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/admin?success=Redirect+atualizado+com+sucesso", http.StatusSeeOther)
}

func (h *Handler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	_ = h.Store.DeleteRedirect(id)
	http.Redirect(w, r, "/admin?success=Redirect+excluido", http.StatusSeeOther)
}

func (h *Handler) AdminToggle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	_ = h.Store.ToggleRedirect(id)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func parseForm(r *http.Request) *models.Redirect {
	active := true
	if _, ok := r.Form["is_active"]; ok {
		active = r.FormValue("is_active") == "on"
	}
	return &models.Redirect{
		Slug:         r.FormValue("slug"),
		TargetURL:    r.FormValue("target_url"),
		IsActive:     active,
		RedirectType: r.FormValue("redirect_type"),
		UTMSource:    r.FormValue("utm_source"),
		UTMMedium:    r.FormValue("utm_medium"),
		UTMCampaign:  r.FormValue("utm_campaign"),
		UTMTerm:      r.FormValue("utm_term"),
		UTMContent:   r.FormValue("utm_content"),
	}
}
