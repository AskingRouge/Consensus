package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"consensus/internal/auth"
)

// Routes registers every REST endpoint onto r. WebSocket routes are
// registered separately in main.go once internal/ws lands (Phase 3).
func (h *Handler) Routes(r chi.Router) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", h.Signup)
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(h.AuthStore))
			r.Get("/me", h.Me)
		})
	})

	r.Route("/api/sessions", func(r chi.Router) {
		r.Use(auth.RequireAuth(h.AuthStore))

		r.Post("/", h.CreateSession)
		r.Get("/", h.ListSessions)

		r.Route("/{sessionID}", func(r chi.Router) {
			r.Get("/", h.GetSession)

			r.Get("/items", h.ListItems)
			r.Post("/items/upload", h.UploadItems)
		})
	})
}
