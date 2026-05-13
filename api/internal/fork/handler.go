package fork

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erc-pham/surus/api/db"
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool, queries: db.New(pool)}
}

func (h *Handler) ForkCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	newCourse, err := h.forkCourseTransaction(r.Context(), courseID, user.ID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{"course": newCourse})
}

func (h *Handler) forkCourseTransaction(ctx context.Context, sourceCourseID, forkerID uuid.UUID) (*db.CreateForkedCourseRow, error) {
	source, err := h.queries.GetCourseByID(ctx, db.GetCourseByIDParams{
		ID:       sourceCourseID,
		ViewerID: pgtype.UUID{Bytes: forkerID, Valid: true},
	})
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "Course not found")
	}

	if source.Visibility == db.VisibilityPrivate && source.OwnerID != forkerID {
		return nil, middleware.NewServiceError("forbidden", "Cannot fork a private course")
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	newCourse, err := qtx.CreateForkedCourse(ctx, db.CreateForkedCourseParams{
		OwnerID:      forkerID,
		Title:        source.Title,
		Description:  source.Description,
		Tags:         source.Tags,
		ThumbnailUrl: source.ThumbnailUrl,
		ForkedFromID: pgtype.UUID{Bytes: sourceCourseID, Valid: true},
	})
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to create forked course", err)
	}

	lessons, err := qtx.ListLessonsByCourse(ctx, sourceCourseID)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to list source lessons", err)
	}

	idMapping := make(map[uuid.UUID]uuid.UUID)

	for _, lesson := range lessons {
		newParentID := pgtype.UUID{}
		if lesson.ParentID.Valid {
			if mappedID, ok := idMapping[lesson.ParentID.Bytes]; ok {
				newParentID = pgtype.UUID{Bytes: mappedID, Valid: true}
			}
		}

		newLesson, err := qtx.CreateLesson(ctx, db.CreateLessonParams{
			CourseID:  newCourse.ID,
			ParentID:  newParentID,
			Position:  lesson.Position,
			Type:      lesson.Type,
			Title:     lesson.Title,
		})
		if err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to copy lesson", err)
		}
		idMapping[lesson.ID] = newLesson.ID

		switch lesson.Type {
		case db.LessonTypeVideo:
			video, err := qtx.GetVideoLesson(ctx, lesson.ID)
			if err == nil {
				qtx.CreateVideoLesson(ctx, db.CreateVideoLessonParams{
					LessonID:     newLesson.ID,
					Provider:     video.Provider,
					ProviderID:   video.ProviderID,
					StartSeconds: video.StartSeconds,
					EndSeconds:   video.EndSeconds,
					CuratorNotes: video.CuratorNotes,
					SourceUrl:    video.SourceUrl,
				})
			}
		case db.LessonTypePage:
			page, err := qtx.GetPageLesson(ctx, lesson.ID)
			if err == nil {
				qtx.CreatePageLesson(ctx, db.CreatePageLessonParams{
					LessonID: newLesson.ID,
					Content:  page.Content,
				})
			}
		case db.LessonTypeQuiz:
			quiz, err := qtx.GetQuizLesson(ctx, lesson.ID)
			if err == nil {
				// Generate new question IDs
				var questions []json.RawMessage
				json.Unmarshal(quiz.Questions, &questions)
				qtx.CreateQuizLesson(ctx, db.CreateQuizLessonParams{
					LessonID:  newLesson.ID,
					Questions: quiz.Questions,
				})
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to commit fork", err)
	}

	return &newCourse, nil
}
