package auth

import (
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()

	r.Get("/google/start", h.GoogleStartRedirect)
	r.Get("/google/callback", h.GoogleCallbackRedirect)
	r.Post("/google/callback", h.GoogleCallback)
	r.Post("/magic-link/request", h.MagicLinkRequest)
	r.Post("/magic-link/verify", h.MagicLinkVerify)
	r.Post("/refresh", h.Refresh)

	r.Group(func(r chi.Router) {
		r.Use(authCfg.OptionalAuth)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.Me)
	})

	return r
}
