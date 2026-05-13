-- name: CreateQuizAttempt :one
INSERT INTO quiz_attempts (user_id, lesson_id, answers, score)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, lesson_id, answers, score, created_at;

-- name: ListQuizAttemptsByUser :many
SELECT id, user_id, lesson_id, answers, score, created_at
FROM quiz_attempts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListQuizAttemptsByUserAndLesson :many
SELECT id, user_id, lesson_id, answers, score, created_at
FROM quiz_attempts
WHERE user_id = $1 AND lesson_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListQuizAttemptsByUserCursor :many
SELECT id, user_id, lesson_id, answers, score, created_at
FROM quiz_attempts
WHERE user_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListQuizAttemptsByUserAndLessonCursor :many
SELECT id, user_id, lesson_id, answers, score, created_at
FROM quiz_attempts
WHERE user_id = $1 AND lesson_id = $2 AND created_at < $3
ORDER BY created_at DESC
LIMIT $4;
