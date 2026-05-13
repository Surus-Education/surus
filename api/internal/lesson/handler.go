package lesson

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) ListLessons(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	lessons, err := h.service.ListLessons(r.Context(), courseID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"lessons": lessons})
}

func (h *Handler) GetLesson(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}

	lesson, err := h.service.GetLesson(r.Context(), courseID, lessonID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"lesson": lesson})
}

func (h *Handler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	var input CreateLessonInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	lesson, err := h.service.CreateLesson(r.Context(), courseID, input)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusCreated, map[string]any{"lesson": lesson})
}

func (h *Handler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}

	var input UpdateLessonInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	lesson, err := h.service.UpdateLesson(r.Context(), courseID, lessonID, input)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"lesson": lesson})
}

func (h *Handler) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}

	if err := h.service.DeleteLesson(r.Context(), courseID, lessonID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ReorderLessons(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid course ID")
		return
	}

	var input struct {
		Moves []ReorderMove `json:"moves" validate:"required,min=1"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.service.ReorderLessons(r.Context(), courseID, input.Moves); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkComplete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}
	if err := h.service.MarkComplete(r.Context(), user.ID, lessonID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnmarkComplete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid lesson ID")
		return
	}
	if err := h.service.UnmarkComplete(r.Context(), user.ID, lessonID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
