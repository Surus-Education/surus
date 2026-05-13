package course

import (
	"time"

	"github.com/google/uuid"
)

type CourseResponse struct {
	ID           uuid.UUID  `json:"id"`
	OwnerID      uuid.UUID  `json:"owner_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Tags         []string   `json:"tags"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	Visibility   string     `json:"visibility"`
	ForkedFromID *uuid.UUID `json:"forked_from_id"`
	ForkedAt     *time.Time `json:"forked_at"`
	EmbedBroken  bool       `json:"embed_broken"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateCourseInput struct {
	Title        string   `json:"title" validate:"required,min=1,max=200"`
	Description  string   `json:"description" validate:"max=5000"`
	Tags         []string `json:"tags" validate:"max=10,dive,max=50"`
	ThumbnailURL *string  `json:"thumbnail_url" validate:"omitempty,url"`
	Visibility   string   `json:"visibility" validate:"required,oneof=public unlisted private"`
}

type UpdateCourseInput struct {
	Title        *string  `json:"title" validate:"omitempty,min=1,max=200"`
	Description  *string  `json:"description" validate:"omitempty,max=5000"`
	Tags         []string `json:"tags" validate:"omitempty,max=10,dive,max=50"`
	ThumbnailURL *string  `json:"thumbnail_url" validate:"omitempty,url"`
	Visibility   *string  `json:"visibility" validate:"omitempty,oneof=public unlisted private"`
}
