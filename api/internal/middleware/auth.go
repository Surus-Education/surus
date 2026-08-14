package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userContextKey contextKey = "user"

type AuthUser struct {
	ID      uuid.UUID
	Email   string
	IsAdmin bool
}

type AuthConfig struct {
	JWTSecret []byte
}

func (cfg *AuthConfig) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := cfg.extractUser(r)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, "unauthenticated", "No valid session")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *AuthConfig) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := cfg.extractUser(r)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, "unauthenticated", "No valid session")
			return
		}
		if !user.IsAdmin {
			RespondError(w, http.StatusForbidden, "forbidden", "Admin access required")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *AuthConfig) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := cfg.extractUser(r)
		if err == nil {
			ctx := context.WithValue(r.Context(), userContextKey, user)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (cfg *AuthConfig) extractUser(r *http.Request) (*AuthUser, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return cfg.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	sub, _ := claims.GetSubject()
	id, err := uuid.Parse(sub)
	if err != nil {
		return nil, err
	}

	email, _ := claims["email"].(string)
	isAdmin, _ := claims["is_admin"].(bool)

	return &AuthUser{
		ID:      id,
		Email:   email,
		IsAdmin: isAdmin,
	}, nil
}

func UserFromContext(ctx context.Context) *AuthUser {
	user, _ := ctx.Value(userContextKey).(*AuthUser)
	return user
}

// originAllowed reports whether origin matches any entry in allowed. An entry
// may contain a single "*" wildcard, which matches any run of characters, so
// that per-branch preview hostnames can be permitted without listing each one.
//
// Keep wildcard entries narrow. A bare "https://*.vercel.app" would let any
// site hosted on vercel.app make credentialed requests against this API, which
// is a cross-site data leak. Include the project prefix and the team slug, e.g.
// "https://surus-web-git-*-myteam.vercel.app".
func originAllowed(origin string, allowed []string) bool {
	for _, pattern := range allowed {
		if pattern == origin {
			return true
		}
		star := strings.Index(pattern, "*")
		if star < 0 {
			continue
		}
		prefix, suffix := pattern[:star], pattern[star+1:]
		if len(origin) >= len(prefix)+len(suffix) &&
			strings.HasPrefix(origin, prefix) &&
			strings.HasSuffix(origin, suffix) {
			return true
		}
	}
	return false
}

// CORS echoes the request's Origin when it is allowlisted. The response cannot
// use "*" because credentials are enabled — browsers reject that combination,
// and the access_token cookie would never be sent.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Responses vary by Origin whether or not this one matched, so
			// caches must not serve one origin's response to another.
			w.Header().Add("Vary", "Origin")

			if origin := r.Header.Get("Origin"); origin != "" && originAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/health") {
			// slog logging handled elsewhere
		}
		next.ServeHTTP(w, r)
	})
}
