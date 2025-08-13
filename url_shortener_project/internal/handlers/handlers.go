package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"

	"urlShortener/internal/database"
	"urlShortener/internal/shortid"
)

type Handler struct {
	store *database.Store
	tmpl  *template.Template
}

func New(store *database.Store) *Handler {
	// parse templates once
	t := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/home.html",
		"templates/result.html",
	))
	return &Handler{store: store, tmpl: t}
}

// Home shows the form
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Error": "", "Result": ""}
	_ = h.tmpl.ExecuteTemplate(w, "home.html", data)
}

// Create handles POST /shorten
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}
	longURL := strings.TrimSpace(r.FormValue("long_url"))
	if !isValidURL(longURL) {
		w.WriteHeader(http.StatusBadRequest)
		_ = h.tmpl.ExecuteTemplate(w, "home.html", map[string]any{
			"Error":  "Invalid URL. Please enter http(s)://...",
			"Result": "",
		})
		return
	}

	// generate unique code (retry a few times on collision)
	var code string
	var err error
	for i := 0; i < 5; i++ {
		code, err = shortid.New(7)
		if err != nil {
			http.Error(w, "cannot generate code", http.StatusInternalServerError)
			return
		}
		if err = h.store.Insert(code, longURL); err == nil {
			break
		}
		if !strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		http.Error(w, "failed to create short url", http.StatusInternalServerError)
		return
	}

	short := origin(r) + "/" + code
	_ = h.tmpl.ExecuteTemplate(w, "result.html", map[string]any{
		"ShortURL": short,
	})
}

// Redirect handles GET /{code}
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" || strings.ContainsRune(code, '/') {
		http.NotFound(w, r)
		return
	}
	u, err := h.store.GetByCode(code)
	if err == sql.ErrNoRows {
		http.NotFound(w, r) // <- clean 404 if not found
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	_ = h.store.IncrementClick(u.ID)
	http.Redirect(w, r, u.LongURL, http.StatusFound)
}

func isValidURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func origin(r *http.Request) string {
	scheme := "http"
	// honor reverse proxy headers if present
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
		scheme = xfp
	} else if r.TLS != nil {
		scheme = "https"
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host
}
