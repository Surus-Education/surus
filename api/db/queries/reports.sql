-- name: CreateReport :one
INSERT INTO reports (reporter_id, target_type, target_id, category, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, reporter_id, target_type, target_id, category, body, status, reviewed_by, reviewed_at, created_at;

-- name: GetExistingReport :one
SELECT id, reporter_id, target_type, target_id, category, body, status, reviewed_by, reviewed_at, created_at
FROM reports
WHERE reporter_id = $1 AND target_type = $2 AND target_id = $3
LIMIT 1;

-- name: ListReportsByStatus :many
SELECT id, reporter_id, target_type, target_id, category, body, status, reviewed_by, reviewed_at, created_at
FROM reports
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListReportsByStatusCursor :many
SELECT id, reporter_id, target_type, target_id, category, body, status, reviewed_by, reviewed_at, created_at
FROM reports
WHERE status = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: UpdateReportStatus :one
UPDATE reports
SET status = $2, reviewed_by = $3, reviewed_at = now()
WHERE id = $1
RETURNING id, reporter_id, target_type, target_id, category, body, status, reviewed_by, reviewed_at, created_at;
