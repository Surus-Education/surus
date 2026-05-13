package course

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, validate: validator.New()}
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := int32(24)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = int32(l)
	}

	if query != "" {
		courses, err := h.service.SearchCourses(r.Context(), query, nil, limit)
		if err != nil {
			middleware.HandleServiceError(w, r, err)
			return
		}
		middleware.RespondJSON(w, http.StatusOK, map[string]any{"data": courses, "next_cursor": nil})
		return
	}

	courses, err := h.service.ListPublicCourses(r.Context(), nil, limit)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"data": courses, "next_cursor": nil})
}

func (h *Handler) GetCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	var viewerID *uuid.UUID
	if user := middleware.UserFromContext(r.Context()); user != nil {
		viewerID = &user.ID
	}

	course, err := h.service.GetCourse(r.Context(), courseID, viewerID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"course": course})
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var input CreateCourseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	course, err := h.service.CreateCourse(r.Context(), user.ID, input)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusCreated, map[string]any{"course": course})
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	var input UpdateCourseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	course, err := h.service.UpdateCourse(r.Context(), courseID, user.ID, input)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"course": course})
}

func (h *Handler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	if err := h.service.DeleteCourse(r.Context(), courseID, user.ID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SaveCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	if err := h.service.SaveCourse(r.Context(), user.ID, courseID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnsaveCourse(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	if err := h.service.UnsaveCourse(r.Context(), user.ID, courseID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
