package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/erc-pham/surus/api/internal/middleware"
)

type Handler struct {
	service     *Service
	frontendURL string
}

func NewHandler(service *Service, frontendURL string) *Handler {
	return &Handler{service: service, frontendURL: frontendURL}
}

func (h *Handler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   600,
	})

	http.Redirect(w, r, h.service.GoogleAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid request body")
		return
	}

	user, tokens, err := h.service.ExchangeGoogleCode(r.Context(), input.Code)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	h.service.SetTokenCookies(w, tokens)
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) MagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid request body")
		return
	}
	if input.Email == "" {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}

	_, err := h.service.RequestMagicLink(r.Context(), input.Email)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	// In production, send email via Postmark here
	// For dev, token is logged server-side

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid request body")
		return
	}

	user, tokens, err := h.service.VerifyMagicLink(r.Context(), input.Token)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	h.service.SetTokenCookies(w, tokens)
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		middleware.RespondError(w, http.StatusUnauthorized, "unauthenticated", "No refresh token")
		return
	}

	user, tokens, err := h.service.RefreshSession(r.Context(), cookie.Value)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	_ = user
	h.service.SetTokenCookies(w, tokens)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user != nil {
		h.service.Logout(r.Context(), user.ID)
	}
	h.service.ClearTokenCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		middleware.RespondError(w, http.StatusUnauthorized, "unauthenticated", "No valid session")
		return
	}

	dbUser, err := h.service.GetMe(r.Context(), user.ID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"user": dbUser})
}

func (h *Handler) GoogleStartRedirect(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   600,
	})

	http.Redirect(w, r, h.service.GoogleAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallbackRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	_ = r.URL.Query().Get("state")

	if code == "" {
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	user, tokens, err := h.service.ExchangeGoogleCode(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	_ = user
	h.service.SetTokenCookies(w, tokens)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	http.Redirect(w, r, h.frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}
