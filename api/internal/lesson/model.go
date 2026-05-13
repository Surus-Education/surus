package lesson

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LessonResponse struct {
	ID          uuid.UUID        `json:"id"`
	CourseID    uuid.UUID        `json:"course_id"`
	ParentID    *uuid.UUID       `json:"parent_id"`
	Position    int32            `json:"position"`
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	EmbedBroken bool             `json:"embed_broken"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Video       *VideoDetail     `json:"video,omitempty"`
	Page        *PageDetail      `json:"page,omitempty"`
	Quiz        *QuizDetail      `json:"quiz,omitempty"`
}

type VideoDetail struct {
	Provider     string           `json:"provider"`
	ProviderID   string           `json:"provider_id"`
	StartSeconds *int32           `json:"start_seconds"`
	EndSeconds   *int32           `json:"end_seconds"`
	CuratorNotes json.RawMessage  `json:"curator_notes,omitempty"`
	SourceURL    string           `json:"source_url"`
}

type PageDetail struct {
	Content json.RawMessage `json:"content"`
}

type QuizDetail struct {
	Questions json.RawMessage `json:"questions"`
}

type CreateLessonInput struct {
	ParentID *string        `json:"parent_id"`
	Position *int32         `json:"position"`
	Type     string         `json:"type" validate:"required,oneof=video page quiz"`
	Title    string         `json:"title" validate:"required,min=1,max=200"`
	Video    *VideoInput    `json:"video"`
	Page     *PageInput     `json:"page"`
	Quiz     *QuizInput     `json:"quiz"`
}

type VideoInput struct {
	Provider     string          `json:"provider" validate:"required,oneof=youtube"`
	ProviderID   string          `json:"provider_id" validate:"required"`
	StartSeconds *int32          `json:"start_seconds"`
	EndSeconds   *int32          `json:"end_seconds"`
	CuratorNotes json.RawMessage `json:"curator_notes"`
	SourceURL    string          `json:"source_url" validate:"required,url"`
}

type PageInput struct {
	Content json.RawMessage `json:"content" validate:"required"`
}

type QuizInput struct {
	Questions json.RawMessage `json:"questions" validate:"required"`
}

type UpdateLessonInput struct {
	Title    *string         `json:"title" validate:"omitempty,min=1,max=200"`
	ParentID *string         `json:"parent_id"`
	Position *int32          `json:"position"`
	Video    *VideoInput     `json:"video"`
	Page     *PageInput      `json:"page"`
	Quiz     *QuizInput      `json:"quiz"`
}

type ReorderMove struct {
	LessonID    string  `json:"lesson_id" validate:"required,uuid"`
	NewParentID *string `json:"new_parent_id"`
	NewPosition int32   `json:"new_position" validate:"min=0"`
}
