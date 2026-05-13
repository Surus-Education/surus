package course

import (
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()

	r.Use(authCfg.OptionalAuth)
	r.Get("/", h.ListCourses)
	r.Get("/{courseId}", h.GetCourse)

	r.Group(func(r chi.Router) {
		r.Use(authCfg.RequireAuth)
		r.Post("/", h.CreateCourse)
		r.Patch("/{courseId}", h.UpdateCourse)
		r.Delete("/{courseId}", h.DeleteCourse)
		r.Post("/{courseId}/save", h.SaveCourse)
		r.Delete("/{courseId}/save", h.UnsaveCourse)
	})

	return r
}
