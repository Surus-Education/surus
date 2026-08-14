package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func IssueAccessToken(secret []byte, userID uuid.UUID, email string, isAdmin bool) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      userID.String(),
		"email":    email,
		"is_admin": isAdmin,
		"iat":      now.Unix(),
		"exp":      now.Add(AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// GenerateRefreshToken returns a random raw token plus a SHA-256 hex digest
// of it. The digest is deterministic rather than a salted bcrypt hash: the
// raw token already carries 256 bits of entropy from crypto/rand, so it
// doesn't need a salt, and a deterministic hash lets RefreshSession look the
// row up directly (WHERE token_hash = $1) instead of iterating every
// non-revoked token and comparing each one.
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken produces the same deterministic digest GenerateRefreshToken
// stores, so a presented raw token can be hashed and looked up directly.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GenerateMagicLinkToken returns a random raw token plus a SHA-256 hex digest
// of it, the same deterministic scheme GenerateRefreshToken uses and for the
// same reason: the raw token already carries 256 bits of entropy from
// crypto/rand, so a salted bcrypt hash buys nothing and only prevents the
// direct indexed lookup VerifyMagicLink needs (WHERE token_hash = $1).
func GenerateMagicLinkToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	return raw, HashMagicLinkToken(raw), nil
}

// HashMagicLinkToken produces the same deterministic digest
// GenerateMagicLinkToken stores, so a presented raw token can be hashed and
// looked up directly.
func HashMagicLinkToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
