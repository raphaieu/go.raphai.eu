package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"go.raphai.eu/auth"
	"go.raphai.eu/handlers"
	"go.raphai.eu/store"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates static
var embedded embed.FS

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/redirects.db"
	}
	os.MkdirAll("./data", 0755)

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer s.Close()

	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		log.Fatal("ADMIN_PASS environment variable is required")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	sessions := auth.NewSessionStore()

	tmplFS, _ := fs.Sub(embedded, "templates")
	tmpl := template.Must(template.New("").ParseFS(tmplFS, "*.html"))

	h := &handlers.Handler{
		Store:     s,
		Sessions:  sessions,
		Tmpl:      tmpl,
		AdminUser: adminUser,
		AdminPass: string(passHash),
	}

	staticFS, _ := fs.Sub(embedded, "static")
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /{slug}", h.Redirect)

	mux.HandleFunc("GET /admin/login", h.AdminLogin)
	mux.HandleFunc("POST /admin/login", h.AdminLogin)
	mux.HandleFunc("POST /admin/logout", h.AdminLogout)
	mux.HandleFunc("GET /admin", h.AdminDashboard)
	mux.HandleFunc("GET /admin/new", h.AdminNew)
	mux.HandleFunc("POST /admin/new", h.AdminNew)
	mux.HandleFunc("GET /admin/edit/{id}", h.AdminEdit)
	mux.HandleFunc("POST /admin/edit/{id}", h.AdminEdit)
	mux.HandleFunc("POST /admin/delete/{id}", h.AdminDelete)
	mux.HandleFunc("POST /admin/toggle/{id}", h.AdminToggle)

	wrappedMux := auth.Middleware(sessions)(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, wrappedMux))
}
