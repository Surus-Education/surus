-- name: GetLessonByID :one
SELECT id, course_id, parent_id, position, type, title, embed_broken, created_at, updated_at
FROM lessons
WHERE id = $1 AND course_id = $2;

-- name: ListLessonsByCourse :many
SELECT id, course_id, parent_id, position, type, title, embed_broken, created_at, updated_at
FROM lessons
WHERE course_id = $1
ORDER BY COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid), position;

-- name: CreateLesson :one
INSERT INTO lessons (course_id, parent_id, position, type, title)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, course_id, parent_id, position, type, title, embed_broken, created_at, updated_at;

-- name: UpdateLesson :one
UPDATE lessons
SET title = COALESCE(sqlc.narg('title'), title),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    position = COALESCE(sqlc.narg('position'), position),
    updated_at = now()
WHERE id = $1 AND course_id = $2
RETURNING id, course_id, parent_id, position, type, title, embed_broken, created_at, updated_at;

-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1 AND course_id = $2;

-- name: UpdateLessonPosition :exec
UPDATE lessons SET parent_id = $2, position = $3, updated_at = now()
WHERE id = $1;

-- name: SetLessonEmbedBroken :exec
UPDATE lessons SET embed_broken = $2, updated_at = now() WHERE id = $1;

-- name: GetMaxPosition :one
SELECT COALESCE(MAX(position), -1)::integer AS max_pos
FROM lessons
WHERE course_id = $1
  AND COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid) = COALESCE(sqlc.narg('parent_id'), '00000000-0000-0000-0000-000000000000'::uuid);

-- name: GetVideoLesson :one
SELECT lesson_id, provider, provider_id, start_seconds, end_seconds, curator_notes, source_url
FROM video_lessons
WHERE lesson_id = $1;

-- name: CreateVideoLesson :exec
INSERT INTO video_lessons (lesson_id, provider, provider_id, start_seconds, end_seconds, curator_notes, source_url)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateVideoLesson :exec
UPDATE video_lessons
SET provider_id = COALESCE(sqlc.narg('provider_id'), provider_id),
    start_seconds = sqlc.narg('start_seconds'),
    end_seconds = sqlc.narg('end_seconds'),
    curator_notes = sqlc.narg('curator_notes'),
    source_url = COALESCE(sqlc.narg('source_url'), source_url)
WHERE lesson_id = $1;

-- name: GetPageLesson :one
SELECT lesson_id, content
FROM page_lessons
WHERE lesson_id = $1;

-- name: CreatePageLesson :exec
INSERT INTO page_lessons (lesson_id, content)
VALUES ($1, $2);

-- name: UpdatePageLesson :exec
UPDATE page_lessons SET content = $2 WHERE lesson_id = $1;

-- name: GetQuizLesson :one
SELECT lesson_id, questions
FROM quiz_lessons
WHERE lesson_id = $1;

-- name: CreateQuizLesson :exec
INSERT INTO quiz_lessons (lesson_id, questions)
VALUES ($1, $2);

-- name: UpdateQuizLesson :exec
UPDATE quiz_lessons SET questions = $2 WHERE lesson_id = $1;

-- name: GetAllVideoLessons :many
SELECT vl.lesson_id, vl.provider, vl.provider_id, vl.start_seconds, vl.end_seconds, vl.curator_notes, vl.source_url
FROM video_lessons vl
JOIN lessons l ON l.id = vl.lesson_id
JOIN courses c ON c.id = l.course_id
WHERE c.owner_id != '00000000-0000-0000-0000-000000000001';
