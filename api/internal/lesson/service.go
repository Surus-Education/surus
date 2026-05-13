package lesson

import (
	"context"
	"encoding/json"

	"github.com/erc-pham/surus/api/db"
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, queries: db.New(pool)}
}

func (s *Service) ListLessons(ctx context.Context, courseID uuid.UUID) ([]LessonResponse, error) {
	rows, err := s.queries.ListLessonsByCourse(ctx, courseID)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to list lessons", err)
	}

	results := make([]LessonResponse, len(rows))
	for i, row := range rows {
		resp := toLessonResponse(row)

		switch db.LessonType(row.Type) {
		case db.LessonTypeVideo:
			video, err := s.queries.GetVideoLesson(ctx, row.ID)
			if err == nil {
				resp.Video = &VideoDetail{
					Provider:   string(video.Provider),
					ProviderID: video.ProviderID,
					SourceURL:  video.SourceUrl,
				}
				if video.StartSeconds.Valid {
					resp.Video.StartSeconds = &video.StartSeconds.Int32
				}
				if video.EndSeconds.Valid {
					resp.Video.EndSeconds = &video.EndSeconds.Int32
				}
				if video.CuratorNotes != nil {
					resp.Video.CuratorNotes = json.RawMessage(video.CuratorNotes)
				}
			}
		case db.LessonTypePage:
			page, err := s.queries.GetPageLesson(ctx, row.ID)
			if err == nil {
				resp.Page = &PageDetail{Content: json.RawMessage(page.Content)}
			}
		case db.LessonTypeQuiz:
			quiz, err := s.queries.GetQuizLesson(ctx, row.ID)
			if err == nil {
				resp.Quiz = &QuizDetail{Questions: json.RawMessage(quiz.Questions)}
			}
		}

		results[i] = resp
	}
	return results, nil
}

func (s *Service) GetLesson(ctx context.Context, courseID, lessonID uuid.UUID) (*LessonResponse, error) {
	row, err := s.queries.GetLessonByID(ctx, db.GetLessonByIDParams{ID: lessonID, CourseID: courseID})
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "Lesson not found")
	}

	resp := toLessonResponse(row)

	switch db.LessonType(row.Type) {
	case db.LessonTypeVideo:
		video, err := s.queries.GetVideoLesson(ctx, row.ID)
		if err == nil {
			resp.Video = &VideoDetail{
				Provider:   string(video.Provider),
				ProviderID: video.ProviderID,
				SourceURL:  video.SourceUrl,
			}
			if video.StartSeconds.Valid {
				resp.Video.StartSeconds = &video.StartSeconds.Int32
			}
			if video.EndSeconds.Valid {
				resp.Video.EndSeconds = &video.EndSeconds.Int32
			}
			if video.CuratorNotes != nil {
				resp.Video.CuratorNotes = json.RawMessage(video.CuratorNotes)
			}
		}
	case db.LessonTypePage:
		page, err := s.queries.GetPageLesson(ctx, row.ID)
		if err == nil {
			resp.Page = &PageDetail{Content: json.RawMessage(page.Content)}
		}
	case db.LessonTypeQuiz:
		quiz, err := s.queries.GetQuizLesson(ctx, row.ID)
		if err == nil {
			resp.Quiz = &QuizDetail{Questions: json.RawMessage(quiz.Questions)}
		}
	}

	return &resp, nil
}

func (s *Service) CreateLesson(ctx context.Context, courseID uuid.UUID, input CreateLessonInput) (*LessonResponse, error) {
	var parentID pgtype.UUID
	if input.ParentID != nil {
		pid, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, middleware.NewServiceError("validation_error", "Invalid parent_id")
		}
		parentID = pgtype.UUID{Bytes: pid, Valid: true}
	}

	position := int32(0)
	if input.Position != nil {
		position = *input.Position
	} else {
		var parentParam pgtype.UUID
		if input.ParentID != nil {
			pid, _ := uuid.Parse(*input.ParentID)
			parentParam = pgtype.UUID{Bytes: pid, Valid: true}
		}
		maxPos, err := s.queries.GetMaxPosition(ctx, db.GetMaxPositionParams{
			CourseID: courseID,
			ParentID: parentParam,
		})
		if err == nil {
			position = maxPos + 1
		}
	}

	row, err := s.queries.CreateLesson(ctx, db.CreateLessonParams{
		CourseID:  courseID,
		ParentID:  parentID,
		Position:  position,
		Type:      db.LessonType(input.Type),
		Title:     input.Title,
	})
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to create lesson", err)
	}

	switch db.LessonType(input.Type) {
	case db.LessonTypeVideo:
		if input.Video == nil {
			return nil, middleware.NewServiceError("validation_error", "Video details required for video lesson")
		}
		params := db.CreateVideoLessonParams{
			LessonID:   row.ID,
			Provider:   db.VideoProvider(input.Video.Provider),
			ProviderID: input.Video.ProviderID,
			SourceUrl:  input.Video.SourceURL,
		}
		if input.Video.StartSeconds != nil {
			params.StartSeconds = pgtype.Int4{Int32: *input.Video.StartSeconds, Valid: true}
		}
		if input.Video.EndSeconds != nil {
			params.EndSeconds = pgtype.Int4{Int32: *input.Video.EndSeconds, Valid: true}
		}
		if input.Video.CuratorNotes != nil {
			params.CuratorNotes = input.Video.CuratorNotes
		}
		if err := s.queries.CreateVideoLesson(ctx, params); err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to create video lesson", err)
		}
	case db.LessonTypePage:
		if input.Page == nil {
			return nil, middleware.NewServiceError("validation_error", "Page content required for page lesson")
		}
		if err := s.queries.CreatePageLesson(ctx, db.CreatePageLessonParams{
			LessonID: row.ID,
			Content:  input.Page.Content,
		}); err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to create page lesson", err)
		}
	case db.LessonTypeQuiz:
		if input.Quiz == nil {
			return nil, middleware.NewServiceError("validation_error", "Quiz questions required for quiz lesson")
		}
		if err := s.queries.CreateQuizLesson(ctx, db.CreateQuizLessonParams{
			LessonID:  row.ID,
			Questions: input.Quiz.Questions,
		}); err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to create quiz lesson", err)
		}
	}

	return s.GetLesson(ctx, courseID, row.ID)
}

func (s *Service) UpdateLesson(ctx context.Context, courseID, lessonID uuid.UUID, input UpdateLessonInput) (*LessonResponse, error) {
	params := db.UpdateLessonParams{ID: lessonID, CourseID: courseID}
	if input.Title != nil {
		params.Title = pgtype.Text{String: *input.Title, Valid: true}
	}
	if input.ParentID != nil {
		pid, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, middleware.NewServiceError("validation_error", "Invalid parent_id")
		}
		params.ParentID = pgtype.UUID{Bytes: pid, Valid: true}
	}
	if input.Position != nil {
		params.Position = pgtype.Int4{Int32: *input.Position, Valid: true}
	}

	_, err := s.queries.UpdateLesson(ctx, params)
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "Lesson not found")
	}

	if input.Video != nil {
		videoParams := db.UpdateVideoLessonParams{LessonID: lessonID}
		if input.Video.ProviderID != "" {
			videoParams.ProviderID = pgtype.Text{String: input.Video.ProviderID, Valid: true}
		}
		if input.Video.SourceURL != "" {
			videoParams.SourceUrl = pgtype.Text{String: input.Video.SourceURL, Valid: true}
		}
		if input.Video.StartSeconds != nil {
			videoParams.StartSeconds = pgtype.Int4{Int32: *input.Video.StartSeconds, Valid: true}
		}
		if input.Video.EndSeconds != nil {
			videoParams.EndSeconds = pgtype.Int4{Int32: *input.Video.EndSeconds, Valid: true}
		}
		if input.Video.CuratorNotes != nil {
			videoParams.CuratorNotes = input.Video.CuratorNotes
		}
		s.queries.UpdateVideoLesson(ctx, videoParams)
	}
	if input.Page != nil {
		s.queries.UpdatePageLesson(ctx, db.UpdatePageLessonParams{LessonID: lessonID, Content: input.Page.Content})
	}
	if input.Quiz != nil {
		s.queries.UpdateQuizLesson(ctx, db.UpdateQuizLessonParams{LessonID: lessonID, Questions: input.Quiz.Questions})
	}

	return s.GetLesson(ctx, courseID, lessonID)
}

func (s *Service) DeleteLesson(ctx context.Context, courseID, lessonID uuid.UUID) error {
	return s.queries.DeleteLesson(ctx, db.DeleteLessonParams{ID: lessonID, CourseID: courseID})
}

func (s *Service) ReorderLessons(ctx context.Context, courseID uuid.UUID, moves []ReorderMove) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return middleware.WrapServiceError("internal_error", "Failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	for _, move := range moves {
		lessonID, _ := uuid.Parse(move.LessonID)
		var parentID pgtype.UUID
		if move.NewParentID != nil {
			pid, _ := uuid.Parse(*move.NewParentID)
			parentID = pgtype.UUID{Bytes: pid, Valid: true}
		}
		if err := qtx.UpdateLessonPosition(ctx, db.UpdateLessonPositionParams{
			ID:       lessonID,
			ParentID: parentID,
			Position: move.NewPosition,
		}); err != nil {
			return middleware.WrapServiceError("internal_error", "Failed to reorder", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) MarkComplete(ctx context.Context, userID, lessonID uuid.UUID) error {
	return s.queries.CreateCompletion(ctx, db.CreateCompletionParams{UserID: userID, LessonID: lessonID})
}

func (s *Service) UnmarkComplete(ctx context.Context, userID, lessonID uuid.UUID) error {
	return s.queries.DeleteCompletion(ctx, db.DeleteCompletionParams{UserID: userID, LessonID: lessonID})
}

func toLessonResponse(row db.Lesson) LessonResponse {
	resp := LessonResponse{
		ID:          row.ID,
		CourseID:    row.CourseID,
		Position:    row.Position,
		Type:        string(row.Type),
		Title:       row.Title,
		EmbedBroken: row.EmbedBroken,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.ParentID.Valid {
		id := uuid.UUID(row.ParentID.Bytes)
		resp.ParentID = &id
	}
	return resp
}
