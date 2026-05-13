-- name: GetCourseByID :one
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE id = $1
  AND (
    visibility = 'public'
    OR visibility = 'unlisted'
    OR owner_id = sqlc.narg('viewer_id')
  );

-- name: GetCourseByIDOwnerOnly :one
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE id = $1 AND owner_id = $2;

-- name: ListPublicCourses :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE visibility = 'public'
ORDER BY created_at DESC
LIMIT $1;

-- name: ListPublicCoursesCursor :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE visibility = 'public' AND created_at < $1
ORDER BY created_at DESC
LIMIT $2;

-- name: SearchCourses :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at,
       ts_rank(search_vector, websearch_to_tsquery('english', $1)) AS rank
FROM courses
WHERE visibility = 'public'
  AND search_vector @@ websearch_to_tsquery('english', $1)
ORDER BY rank DESC, created_at DESC
LIMIT $2;

-- name: SearchCoursesCursor :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at,
       ts_rank(search_vector, websearch_to_tsquery('english', $1)) AS rank
FROM courses
WHERE visibility = 'public'
  AND search_vector @@ websearch_to_tsquery('english', $1)
  AND created_at < $2
ORDER BY rank DESC, created_at DESC
LIMIT $3;

-- name: SearchCoursesByTags :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE visibility = 'public'
  AND tags @> $1::text[]
ORDER BY created_at DESC
LIMIT $2;

-- name: ListUserCourses :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE owner_id = $1
  AND (
    visibility = 'public'
    OR visibility = 'unlisted'
    OR owner_id = sqlc.narg('viewer_id')
  )
ORDER BY created_at DESC;

-- name: ListUserOwnCourses :many
SELECT id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at
FROM courses
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: CreateCourse :one
INSERT INTO courses (owner_id, title, description, tags, thumbnail_url, visibility)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at;

-- name: CreateForkedCourse :one
INSERT INTO courses (owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at)
VALUES ($1, $2, $3, $4, $5, 'private', $6, now())
RETURNING id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at;

-- name: UpdateCourse :one
UPDATE courses
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    tags = COALESCE(sqlc.narg('tags'), tags),
    thumbnail_url = COALESCE(sqlc.narg('thumbnail_url'), thumbnail_url),
    visibility = COALESCE(sqlc.narg('visibility'), visibility),
    updated_at = now()
WHERE id = $1 AND owner_id = $2
RETURNING id, owner_id, title, description, tags, thumbnail_url, visibility, forked_from_id, forked_at, embed_broken, created_at, updated_at;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1 AND owner_id = $2;

-- name: SetCourseEmbedBroken :exec
UPDATE courses SET embed_broken = $2, updated_at = now() WHERE id = $1;
