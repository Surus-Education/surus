-- name: CreateCompletion :exec
INSERT INTO completions (user_id, lesson_id)
VALUES ($1, $2)
ON CONFLICT (user_id, lesson_id) DO NOTHING;

-- name: DeleteCompletion :exec
DELETE FROM completions WHERE user_id = $1 AND lesson_id = $2;

-- name: ListCompletionsByUserAndCourse :many
SELECT c.id, c.user_id, c.lesson_id, c.completed_at
FROM completions c
JOIN lessons l ON l.id = c.lesson_id
WHERE c.user_id = $1 AND l.course_id = $2;
