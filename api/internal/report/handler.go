package report

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erc-pham/surus/api/db"
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	validate *validator.Validate
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool, queries: db.New(pool), validate: validator.New()}
}

type ReportInput struct {
	Category string `json:"category" validate:"required,oneof=incorrect harmful copyright other"`
	Body     string `json:"body" validate:"max=2000"`
}

func (h *Handler) ReportCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	var input ReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	existing, err := h.queries.GetExistingReport(r.Context(), db.GetExistingReportParams{
		ReporterID: user.ID,
		TargetType: db.ReportTargetTypeCourse,
		TargetID:   courseID,
	})
	if err == nil && existing.ID != uuid.Nil {
		middleware.RespondError(w, http.StatusConflict, "conflict", "You've already reported this")
		return
	}

	report, err := h.queries.CreateReport(r.Context(), db.CreateReportParams{
		ReporterID: user.ID,
		TargetType: db.ReportTargetTypeCourse,
		TargetID:   courseID,
		Category:   db.ReportCategory(input.Category),
		Body:       pgtype.Text{String: input.Body, Valid: input.Body != ""},
	})
	if err != nil {
		middleware.HandleServiceError(w, r, middleware.WrapServiceError("internal_error", "Failed to create report", err))
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{"report": report})
}

func (h *Handler) ReportLesson(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}

	var input ReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	existing, err := h.queries.GetExistingReport(r.Context(), db.GetExistingReportParams{
		ReporterID: user.ID,
		TargetType: db.ReportTargetTypeLesson,
		TargetID:   lessonID,
	})
	if err == nil && existing.ID != uuid.Nil {
		middleware.RespondError(w, http.StatusConflict, "conflict", "You've already reported this")
		return
	}

	report, err := h.queries.CreateReport(r.Context(), db.CreateReportParams{
		ReporterID: user.ID,
		TargetType: db.ReportTargetTypeLesson,
		TargetID:   lessonID,
		Category:   db.ReportCategory(input.Category),
		Body:       pgtype.Text{String: input.Body, Valid: input.Body != ""},
	})
	if err != nil {
		middleware.HandleServiceError(w, r, middleware.WrapServiceError("internal_error", "Failed to create report", err))
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{"report": report})
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}
	limitStr := r.URL.Query().Get("limit")
	limit := int32(50)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = int32(l)
	}

	reports, err := h.queries.ListReportsByStatus(r.Context(), db.ListReportsByStatusParams{
		Status: db.ReportStatus(status),
		Limit:  limit,
	})
	if err != nil {
		middleware.HandleServiceError(w, r, middleware.WrapServiceError("internal_error", "Failed to list reports", err))
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"data": reports, "next_cursor": nil})
}

func (h *Handler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "reportId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid report ID")
		return
	}

	var input struct {
		Status string `json:"status" validate:"required,oneof=reviewed actioned dismissed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	report, err := h.queries.UpdateReportStatus(r.Context(), db.UpdateReportStatusParams{
		ID:         reportID,
		Status:     db.ReportStatus(input.Status),
		ReviewedBy: pgtype.UUID{Bytes: user.ID, Valid: true},
	})
	if err != nil {
		middleware.HandleServiceError(w, r, middleware.WrapServiceError("not_found", "Report not found", err))
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"report": report})
}

func (h *Handler) AdminRoutes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()
	r.Use(authCfg.RequireAdmin)
	r.Get("/", h.ListReports)
	r.Patch("/{reportId}", h.UpdateReport)
	return r
}
