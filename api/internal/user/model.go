package user

import (
	"time"

	"github.com/erc-pham/surus/api/db"
	"github.com/google/uuid"
)

type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Bio         *string    `json:"bio"`
	AvatarURL   *string    `json:"avatar_url"`
	IsAdmin     bool       `json:"is_admin"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToUserResponse(u *db.User) *UserResponse {
	resp := &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsAdmin:     u.IsAdmin,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
	if u.Bio.Valid {
		resp.Bio = &u.Bio.String
	}
	if u.AvatarUrl.Valid {
		resp.AvatarURL = &u.AvatarUrl.String
	}
	return resp
}

type PublicUserResponse struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Bio         *string   `json:"bio"`
	AvatarURL   *string   `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToPublicUserResponse(u *db.User) *PublicUserResponse {
	resp := &PublicUserResponse{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
	if u.Bio.Valid {
		resp.Bio = &u.Bio.String
	}
	if u.AvatarUrl.Valid {
		resp.AvatarURL = &u.AvatarUrl.String
	}
	return resp
}
