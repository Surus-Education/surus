package course

import (
	"context"
	"strings"

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

func (s *Service) GetCourse(ctx context.Context, courseID uuid.UUID, viewerID *uuid.UUID) (*db.GetCourseByIDRow, error) {
	params := db.GetCourseByIDParams{ID: courseID}
	if viewerID != nil {
		params.ViewerID = pgtype.UUID{Bytes: *viewerID, Valid: true}
	}
	course, err := s.queries.GetCourseByID(ctx, params)
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "Course not found")
	}
	return &course, nil
}

func (s *Service) ListPublicCourses(ctx context.Context, cursor *string, limit int32) ([]db.ListPublicCoursesRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 24
	}
	courses, err := s.queries.ListPublicCourses(ctx, limit)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to list courses", err)
	}
	return courses, nil
}

func (s *Service) SearchCourses(ctx context.Context, query string, tags []string, limit int32) ([]db.SearchCoursesRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 24
	}
	if query != "" {
		courses, err := s.queries.SearchCourses(ctx, db.SearchCoursesParams{
			WebsearchToTsquery: query,
			Limit:              limit,
		})
		if err != nil {
			return nil, middleware.WrapServiceError("internal_error", "Failed to search courses", err)
		}
		return courses, nil
	}
	return nil, nil
}

func (s *Service) CreateCourse(ctx context.Context, ownerID uuid.UUID, input CreateCourseInput) (*db.CreateCourseRow, error) {
	tags := make([]string, len(input.Tags))
	for i, t := range input.Tags {
		tags[i] = strings.ToLower(strings.TrimSpace(t))
	}

	params := db.CreateCourseParams{
		OwnerID:     ownerID,
		Title:       input.Title,
		Description: input.Description,
		Tags:        tags,
		Visibility:  db.Visibility(input.Visibility),
	}
	if input.ThumbnailURL != nil {
		params.ThumbnailUrl = pgtype.Text{String: *input.ThumbnailURL, Valid: true}
	}

	course, err := s.queries.CreateCourse(ctx, params)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to create course", err)
	}
	return &course, nil
}

func (s *Service) UpdateCourse(ctx context.Context, courseID, ownerID uuid.UUID, input UpdateCourseInput) (*db.UpdateCourseRow, error) {
	params := db.UpdateCourseParams{ID: courseID, OwnerID: ownerID}
	if input.Title != nil {
		params.Title = pgtype.Text{String: *input.Title, Valid: true}
	}
	if input.Description != nil {
		params.Description = pgtype.Text{String: *input.Description, Valid: true}
	}
	if input.Tags != nil {
		tags := make([]string, len(input.Tags))
		for i, t := range input.Tags {
			tags[i] = strings.ToLower(strings.TrimSpace(t))
		}
		params.Tags = tags
	}
	if input.ThumbnailURL != nil {
		params.ThumbnailUrl = pgtype.Text{String: *input.ThumbnailURL, Valid: true}
	}
	if input.Visibility != nil {
		params.Visibility = db.NullVisibility{Visibility: db.Visibility(*input.Visibility), Valid: true}
	}

	course, err := s.queries.UpdateCourse(ctx, params)
	if err != nil {
		return nil, middleware.WrapServiceError("not_found", "Course not found or not owned by you", err)
	}
	return &course, nil
}

func (s *Service) DeleteCourse(ctx context.Context, courseID, ownerID uuid.UUID) error {
	return s.queries.DeleteCourse(ctx, db.DeleteCourseParams{ID: courseID, OwnerID: ownerID})
}

func (s *Service) SaveCourse(ctx context.Context, userID, courseID uuid.UUID) error {
	return s.queries.CreateSave(ctx, db.CreateSaveParams{UserID: userID, CourseID: courseID})
}

func (s *Service) UnsaveCourse(ctx context.Context, userID, courseID uuid.UUID) error {
	return s.queries.DeleteSave(ctx, db.DeleteSaveParams{UserID: userID, CourseID: courseID})
}

func (s *Service) ListUserCourses(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID) ([]db.ListUserCoursesRow, error) {
	params := db.ListUserCoursesParams{OwnerID: userID}
	if viewerID != nil {
		params.ViewerID = pgtype.UUID{Bytes: *viewerID, Valid: true}
	}
	return s.queries.ListUserCourses(ctx, params)
}
