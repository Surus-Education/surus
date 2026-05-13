package user

import (
	"context"
	"time"

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

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, middleware.NewServiceError("not_found", "User not found")
	}
	return &user, nil
}

type UpdateInput struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=100"`
	Bio         *string `json:"bio" validate:"omitempty,max=500"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,url"`
}

func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, input UpdateInput) (*db.User, error) {
	params := db.UpdateUserParams{ID: userID}
	if input.DisplayName != nil {
		params.DisplayName = pgtype.Text{String: *input.DisplayName, Valid: true}
	}
	if input.Bio != nil {
		params.Bio = pgtype.Text{String: *input.Bio, Valid: true}
	}
	if input.AvatarURL != nil {
		params.AvatarUrl = pgtype.Text{String: *input.AvatarURL, Valid: true}
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, middleware.WrapServiceError("internal_error", "Failed to update user", err)
	}
	return &user, nil
}

func (s *Service) GetLibrary(ctx context.Context, userID uuid.UUID) (saved []db.ListSavedCoursesRow, created []db.ListUserOwnCoursesRow, err error) {
	saved, err = s.queries.ListSavedCourses(ctx, userID)
	if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to list saved courses", err)
	}
	created, err = s.queries.ListUserOwnCourses(ctx, userID)
	if err != nil {
		return nil, nil, middleware.WrapServiceError("internal_error", "Failed to list created courses", err)
	}
	return saved, created, nil
}

func (s *Service) ScheduleDeletion(ctx context.Context, userID uuid.UUID) error {
	deletionTime := time.Now().Add(30 * 24 * time.Hour)
	return s.queries.ScheduleUserDeletion(ctx, db.ScheduleUserDeletionParams{
		ID:                  userID,
		DeletionScheduledAt: pgtype.Timestamptz{Time: deletionTime, Valid: true},
	})
}

func (s *Service) CancelDeletion(ctx context.Context, userID uuid.UUID) error {
	return s.queries.CancelUserDeletion(ctx, userID)
}
