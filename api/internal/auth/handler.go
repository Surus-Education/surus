package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/erc-pham/surus/api/internal/middleware"
)

// oauthStateCookieName is shared by every handler that sets or reads the
// CSRF state cookie for the Google OAuth flow.
const oauthStateCookieName = "oauth_state"

// setOAuthStateCookie mirrors the Secure/SameSite attributes the session
// cookies use (Service.CookieAttrs) rather than hardcoding them.
//
// Note the previous SameSite=Lax was NOT the reason OAuth failed: Lax cookies
// are sent on cross-site top-level GET navigations, which is exactly what
// Google's redirect back to the callback is, and a cookie without Secure is
// still sent over HTTPS. The real defect was that state was never verified at
// all. The attributes are shared here so the JSON GoogleStart/GoogleCallback
// pair works too — that flow reaches the callback via fetch rather than a
// top-level navigation, and Lax genuinely would withhold the cookie there.
func (h *Handler) setOAuthStateCookie(w http.ResponseWriter, state string) {
	secure, sameSite := h.service.CookieAttrs()
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   600,
	})
}

func (h *Handler) clearOAuthStateCookie(w http.ResponseWriter) {
	secure, sameSite := h.service.CookieAttrs()
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

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

	h.setOAuthStateCookie(w, state)

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

	h.setOAuthStateCookie(w, state)

	http.Redirect(w, r, h.service.GoogleAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallbackRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// CSRF check: the state returned by Google must match the one we handed
	// out and stored in the oauth_state cookie. Missing cookie or mismatch
	// both fail closed.
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || state == "" ||
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		h.clearOAuthStateCookie(w)
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	user, tokens, err := h.service.ExchangeGoogleCode(r.Context(), code)
	if err != nil {
		h.clearOAuthStateCookie(w)
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// Populate the API-origin cookie jar. This still works because this
	// handler runs on a top-level navigation to the API origin — see
	// PLAN.md decision 2. It does NOT make the session visible to the web
	// app, which reads its own origin's jar; the handoff code below bridges
	// that gap.
	h.service.SetTokenCookies(w, tokens)
	h.clearOAuthStateCookie(w)

	handoffCode, err := h.service.CreateHandoffCode(r.Context(), user.ID)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// Never put the JWT itself in the URL — it would land in browser
	// history, access logs, and Referer headers. The code is an opaque,
	// single-use, 60s-TTL lookup key; POST /v1/auth/handoff exchanges it for
	// the real tokens server-to-server.
	http.Redirect(w, r, h.frontendURL+"/auth/callback?code="+url.QueryEscape(handoffCode), http.StatusTemporaryRedirect)
}

// Handoff exchanges a single-use handoff code (minted by
// GoogleCallbackRedirect) for a fresh token pair. It is meant to be called
// server-to-server by the web app's own route handler, not by a browser
// directly: the response body carries the raw access_token/refresh_token so
// the caller can set its own cookies on the web origin, so this endpoint
// deliberately does NOT set any cookies itself. Single-use enforcement and
// the short TTL on the code are the entire defense for handing out raw
// tokens in a response body — see Service.RedeemHandoffCode.
func (h *Handler) Handoff(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid request body")
		return
	}
	if input.Code == "" {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Code is required")
		return
	}

	_, tokens, err := h.service.RedeemHandoffCode(r.Context(), input.Code)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}
