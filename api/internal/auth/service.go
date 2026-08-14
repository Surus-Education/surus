package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/erc-pham/surus/api/db"
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool              *pgxpool.Pool
	queries           *db.Queries
	jwtSecret         []byte
	googleClientID    string
	googleSecret      string
	googleRedirectURL string
	magicLinkBaseURL  string
	appEnv            string
}

type Config struct {
	Pool              *pgxpool.Pool
	JWTSecret         []byte
	GoogleClientID    string
	GoogleSecret      string
	GoogleRedirectURL string
	MagicLinkBaseURL  string
	AppEnv            string
}

func NewService(cfg Config) *Service {
	return &Service{
		pool:              cfg.Pool,
		queries:           db.New(cfg.Pool),
		jwtSecret:         cfg.JWTSecret,
		googleClientID:    cfg.GoogleClientID,
		googleSecret:      cfg.GoogleSecret,
		googleRedirectURL: cfg.GoogleRedirectURL,
		magicLinkBaseURL:  cfg.MagicLinkBaseURL,
		appEnv:            cfg.AppEnv,
	}
}

type GoogleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *Service) ExchangeGoogleCode(ctx context.Context, code string) (*db.User, *TokenPair, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleClientID)
	data.Set("client_secret", s.googleSecret)
	data.Set("redirect_uri", s.googleRedirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to exchange code", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, middleware.NewServiceError("validation_error", fmt.Sprintf("Google token exchange failed: %s", string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to decode token response", err)
	}

	userInfo, err := s.fetchGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.upsertGoogleUser(ctx, userInfo)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) fetchGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to fetch user info", err)
	}
	defer resp.Body.Close()

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to decode user info", err)
	}
	return &info, nil
}

func (s *Service) upsertGoogleUser(ctx context.Context, info *GoogleUserInfo) (*db.User, error) {
	existing, err := s.queries.GetOAuthAccount(ctx, db.GetOAuthAccountParams{
		Provider:   "google",
		ProviderID: info.Sub,
	})
	if err == nil {
		user, err := s.queries.GetUserByID(ctx, existing.UserID)
		if err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to fetch user", err)
		}
		return &user, nil
	}

	user, err := s.queries.GetUserByEmail(ctx, info.Email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, middleware.WrapServiceError("internal_error", "Failed to check user", err)
	}

	if err == pgx.ErrNoRows {
		displayName := info.Name
		if displayName == "" {
			displayName = strings.Split(info.Email, "@")[0]
		}
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:       info.Email,
			DisplayName: displayName,
		})
		if err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to create user", err)
		}
	}

	_, err = s.queries.CreateOAuthAccount(ctx, db.CreateOAuthAccountParams{
		UserID:     user.ID,
		Provider:   "google",
		ProviderID: info.Sub,
	})
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to link OAuth", err)
	}

	return &user, nil
}

func (s *Service) RequestMagicLink(ctx context.Context, email string) (string, error) {
	raw, hash, err := GenerateMagicLinkToken()
	if err != nil {
		return "", middleware.WrapServiceError("internal_error", "Failed to generate token", err)
	}

	_, err = s.queries.CreateMagicLinkToken(ctx, db.CreateMagicLinkTokenParams{
		Email:     email,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", middleware.WrapServiceError("internal_error", "Failed to store token", err)
	}

	return raw, nil
}

func (s *Service) VerifyMagicLink(ctx context.Context, rawToken string) (*db.User, *TokenPair, error) {
	hash := HashMagicLinkToken(rawToken)

	// GetMagicLinkTokenByHash already filters to used_at IS NULL AND
	// expires_at > now(), so a missing row covers "not found", "expired", and
	// "already used" alike.
	rows, err := s.queries.GetMagicLinkTokenByHash(ctx, hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, middleware.NewServiceError("unauthenticated", "Invalid or expired token")
		}
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to look up magic link token", err)
	}

	if err := s.queries.MarkMagicLinkTokenUsed(ctx, rows.ID); err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to mark token", err)
	}

	user, err := s.queries.GetUserByEmail(ctx, rows.Email)
	if err == pgx.ErrNoRows {
		displayName := strings.Split(rows.Email, "@")[0]
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:       rows.Email,
			DisplayName: displayName,
		})
		if err != nil {
			return nil, nil, middleware.WrapServiceError("internal_error", "Failed to create user", err)
		}
	} else if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to fetch user", err)
	}

	tokens, err := s.issueTokens(ctx, &user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *Service) RefreshSession(ctx context.Context, rawRefreshToken string) (*db.User, *TokenPair, error) {
	if rawRefreshToken == "" {
		return nil, nil, middleware.NewServiceError("unauthenticated", "Invalid refresh token")
	}

	hash := HashRefreshToken(rawRefreshToken)

	// GetRefreshTokenByHash already filters to revoked_at IS NULL AND expires_at > now(),
	// so a missing row covers "not found", "expired", and "revoked" alike.
	stored, err := s.queries.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, middleware.NewServiceError("unauthenticated", "Invalid refresh token")
		}
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to look up refresh token", err)
	}

	user, err := s.queries.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to fetch user", err)
	}

	tokens, err := s.issueTokens(ctx, &user)
	if err != nil {
		return nil, nil, err
	}

	// Rotate: revoke the token that was just used so it can't be replayed.
	if err := s.queries.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to revoke used refresh token", err)
	}

	return &user, tokens, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.queries.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return err
	}
	// Also revoke any outstanding handoff codes: a code that was minted but
	// never redeemed (e.g. the user closed the tab mid-redirect) must not be
	// usable to resurrect a session after logout.
	return s.queries.RevokeAllUserHandoffCodes(ctx, userID)
}

// CreateHandoffCode mints a single-use, short-lived code bound to userID and
// returns the raw value for the caller to put in a redirect URL. Only the
// SHA-256 hash is stored — see db/migrations' handoff_codes comment for why.
func (s *Service) CreateHandoffCode(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, hash, err := GenerateHandoffCode()
	if err != nil {
		return "", middleware.WrapServiceError("internal_error", "Failed to generate handoff code", err)
	}

	_, err = s.queries.CreateHandoffCode(ctx, db.CreateHandoffCodeParams{
		UserID:    userID,
		CodeHash:  hash,
		ExpiresAt: time.Now().Add(HandoffCodeTTL),
	})
	if err != nil {
		return "", middleware.WrapServiceError("internal_error", "Failed to store handoff code", err)
	}

	return raw, nil
}

// RedeemHandoffCode exchanges a raw handoff code for a fresh token pair. The
// underlying query is a conditional UPDATE ... WHERE used_at IS NULL RETURNING,
// which atomically claims the row — a concurrent double-submit of the same
// code can only have one winner, so this alone is what enforces single-use.
func (s *Service) RedeemHandoffCode(ctx context.Context, rawCode string) (*db.User, *TokenPair, error) {
	if rawCode == "" {
		return nil, nil, middleware.NewServiceError("unauthenticated", "Invalid or expired handoff code")
	}

	hash := HashHandoffCode(rawCode)

	row, err := s.queries.RedeemHandoffCode(ctx, hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, middleware.NewServiceError("unauthenticated", "Invalid or expired handoff code")
		}
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to redeem handoff code", err)
	}

	user, err := s.queries.GetUserByID(ctx, row.UserID)
	if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to fetch user", err)
	}

	tokens, err := s.issueTokens(ctx, &user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*db.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "User not found")
	}
	return &user, nil
}

func (s *Service) issueTokens(ctx context.Context, user *db.User) (*TokenPair, error) {
	accessToken, err := IssueAccessToken(s.jwtSecret, user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to issue access token", err)
	}

	rawRefresh, hashRefresh, err := GenerateRefreshToken()
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to generate refresh token", err)
	}

	_, err = s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	})
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to store refresh token", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

// refreshCookiePath must match the mounted route (/v1/auth/refresh). A cookie
// scoped to a narrower path than the endpoint is simply never sent, which makes
// every refresh attempt look like an expired session.
const refreshCookiePath = "/v1/auth/refresh"

// cookieAttrs picks Secure and SameSite together, because the two are coupled:
// browsers reject SameSite=None unless Secure is also set.
//
// Deployed, the web app and this API are separate Vercel projects on different
// origins, so the cookies are cross-site and must be SameSite=None or the
// browser withholds them from every fetch the frontend makes. Local
// development runs over plain HTTP, where SameSite=None is invalid, so it falls
// back to Lax — same-origin-ish local usage doesn't need None anyway.
func (s *Service) cookieAttrs() (bool, http.SameSite) {
	if s.appEnv == "development" {
		return false, http.SameSiteLaxMode
	}
	return true, http.SameSiteNoneMode
}

// CookieAttrs exposes cookieAttrs to callers outside the package (the
// handler's oauth_state cookie) so every cookie this service issues shares
// one Secure/SameSite decision instead of hardcoding its own.
func (s *Service) CookieAttrs() (bool, http.SameSite) {
	return s.cookieAttrs()
}

func (s *Service) SetTokenCookies(w http.ResponseWriter, tokens *TokenPair) {
	secure, sameSite := s.cookieAttrs()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   900,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     refreshCookiePath,
		MaxAge:   2592000,
	})
}

// ClearTokenCookies must repeat the same Secure/SameSite/Path attributes used
// when setting them. A deletion cookie whose attributes don't match writes a
// second cookie instead of overwriting the original, leaving the user logged in.
func (s *Service) ClearTokenCookies(w http.ResponseWriter) {
	secure, sameSite := s.cookieAttrs()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     refreshCookiePath,
		MaxAge:   -1,
	})
}

func (s *Service) GoogleAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.googleClientID)
	params.Set("redirect_uri", s.googleRedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}
