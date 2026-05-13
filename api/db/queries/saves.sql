-- name: CreateSave :exec
INSERT INTO saves (user_id, course_id)
VALUES ($1, $2)
ON CONFLICT (user_id, course_id) DO NOTHING;

-- name: DeleteSave :exec
DELETE FROM saves WHERE user_id = $1 AND course_id = $2;

-- name: ListSavedCourses :many
SELECT c.id, c.owner_id, c.title, c.description, c.tags, c.thumbnail_url, c.visibility, c.forked_from_id, c.forked_at, c.embed_broken, c.created_at, c.updated_at
FROM saves s
JOIN courses c ON c.id = s.course_id
WHERE s.user_id = $1
ORDER BY s.created_at DESC;

-- name: IsSaved :one
SELECT EXISTS(SELECT 1 FROM saves WHERE user_id = $1 AND course_id = $2) AS is_saved;
