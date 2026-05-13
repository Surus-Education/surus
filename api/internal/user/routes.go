package user

import (
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()

	r.Get("/{userId}", h.GetUser)

	r.Group(func(r chi.Router) {
		r.Use(authCfg.RequireAuth)
		r.Patch("/me", h.UpdateMe)
		r.Get("/me/library", h.GetLibrary)
		r.Get("/me/quiz-attempts", h.GetQuizAttempts)
		r.Post("/me/export", h.ExportData)
		r.Delete("/me", h.DeleteMe)
		r.Post("/me/restore", h.RestoreMe)
	})

	return r
}
