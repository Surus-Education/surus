package lesson

import (
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()

	r.Use(authCfg.OptionalAuth)
	r.Get("/", h.ListLessons)
	r.Get("/{lessonId}", h.GetLesson)

	r.Group(func(r chi.Router) {
		r.Use(authCfg.RequireAuth)
		r.Post("/", h.CreateLesson)
		r.Patch("/{lessonId}", h.UpdateLesson)
		r.Delete("/{lessonId}", h.DeleteLesson)
		r.Patch("/reorder", h.ReorderLessons)
		r.Post("/{lessonId}/complete", h.MarkComplete)
		r.Delete("/{lessonId}/complete", h.UnmarkComplete)
	})

	return r
}
