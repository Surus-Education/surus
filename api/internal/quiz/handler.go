package quiz

import (
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

type AttemptInput struct {
	Answers map[string]any `json:"answers"`
}

type QuestionOption struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

type Question struct {
	ID          string           `json:"id"`
	Prompt      json.RawMessage  `json:"prompt"`
	Type        string           `json:"type"`
	Options     []QuestionOption `json:"options"`
	Explanation json.RawMessage  `json:"explanation"`
}

type QuestionResult struct {
	Correct     bool            `json:"correct"`
	Explanation json.RawMessage `json:"explanation,omitempty"`
}

func (h *Handler) SubmitAttempt(w http.ResponseWriter, r *http.Request) {
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
	_ = courseID

	var input AttemptInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	quiz, err := h.queries.GetQuizLesson(r.Context(), lessonID)
	if err != nil {
		middleware.RespondError(w, http.StatusNotFound, "not_found", "Quiz not found")
		return
	}

	var questions []Question
	if err := json.Unmarshal(quiz.Questions, &questions); err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "internal_error", "Failed to parse questions")
		return
	}

	results := make([]QuestionResult, len(questions))
	totalCorrect := 0
	totalGraded := 0

	for i, q := range questions {
		if q.Type == "short_answer" {
			results[i] = QuestionResult{Correct: false, Explanation: q.Explanation}
			continue
		}

		totalGraded++
		answer, ok := input.Answers[q.ID]
		if !ok {
			results[i] = QuestionResult{Correct: false, Explanation: q.Explanation}
			continue
		}

		correct := false
		answerStr, isStr := answer.(string)
		if isStr {
			for _, opt := range q.Options {
				if opt.ID == answerStr && opt.Correct {
					correct = true
					break
				}
			}
		}

		if correct {
			totalCorrect++
		}
		results[i] = QuestionResult{Correct: correct, Explanation: q.Explanation}
	}

	var score *float64
	if totalGraded > 0 {
		s := float64(totalCorrect) / float64(totalGraded) * 100
		score = &s
	}

	user := middleware.UserFromContext(r.Context())
	var attemptID *uuid.UUID
	if user != nil {
		answersJSON, _ := json.Marshal(input.Answers)
		var scoreNum pgtype.Numeric
		if score != nil {
			scoreNum.Valid = true
			// Store as text representation for pgtype.Numeric
			scoreNum.Int = nil
		}
		attempt, err := h.queries.CreateQuizAttempt(r.Context(), db.CreateQuizAttemptParams{
			UserID:   user.ID,
			LessonID: lessonID,
			Answers:  answersJSON,
			Score:    scoreNum,
		})
		if err == nil {
			attemptID = &attempt.ID
		}
	}

	resp := map[string]any{
		"question_results": results,
	}
	if score != nil {
		resp["score"] = *score
	}
	if attemptID != nil {
		resp["attempt_id"] = attemptID
	}

	middleware.RespondJSON(w, http.StatusOK, resp)
}

func (h *Handler) Routes(authCfg *middleware.AuthConfig) chi.Router {
	r := chi.NewRouter()
	r.Use(authCfg.OptionalAuth)
	r.Post("/", h.SubmitAttempt)
	return r
}
