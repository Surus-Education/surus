package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return raw, string(hashed), nil
}

func VerifyRefreshToken(raw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

func GenerateMagicLinkToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return raw, string(hashed), nil
}
