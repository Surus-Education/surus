package user

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

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"user": ToPublicUserResponse(user)})
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.service.UpdateUser(r.Context(), authUser.ID, input)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"user": ToUserResponse(user)})
}

func (h *Handler) GetLibrary(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	saved, created, err := h.service.GetLibrary(r.Context(), authUser.ID)
	if err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"saved": saved, "created": created})
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	if err := h.service.ScheduleDeletion(r.Context(), authUser.ID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) RestoreMe(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	if err := h.service.CancelDeletion(r.Context(), authUser.ID); err != nil {
		middleware.HandleServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	// In full implementation, enqueue a River job
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetQuizAttempts(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	_ = authUser
	middleware.RespondJSON(w, http.StatusOK, map[string]any{"data": []any{}, "next_cursor": nil})
}
