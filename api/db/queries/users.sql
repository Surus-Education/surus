-- name: GetUserByID :one
SELECT id, email, display_name, bio, avatar_url, is_admin, deleted_at, deletion_scheduled_at, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, display_name, bio, avatar_url, is_admin, deleted_at, deletion_scheduled_at, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, display_name)
VALUES ($1, $2)
RETURNING id, email, display_name, bio, avatar_url, is_admin, deleted_at, deletion_scheduled_at, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET display_name = COALESCE(sqlc.narg('display_name'), display_name),
    bio = COALESCE(sqlc.narg('bio'), bio),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, display_name, bio, avatar_url, is_admin, deleted_at, deletion_scheduled_at, created_at, updated_at;

-- name: ScheduleUserDeletion :exec
UPDATE users
SET deletion_scheduled_at = $2, updated_at = now()
WHERE id = $1;

-- name: CancelUserDeletion :exec
UPDATE users
SET deletion_scheduled_at = NULL, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: FinalizeUserDeletion :exec
UPDATE users
SET deleted_at = now(),
    email = 'deleted-' || users.id::text || '@deleted.local',
    display_name = 'Deleted user',
    bio = NULL,
    avatar_url = NULL,
    updated_at = now()
WHERE users.id = $1 AND deletion_scheduled_at IS NOT NULL AND deletion_scheduled_at <= now() AND deleted_at IS NULL;

-- name: GetUsersWithPendingDeletion :many
SELECT id, email, display_name, bio, avatar_url, is_admin, deleted_at, deletion_scheduled_at, created_at, updated_at
FROM users
WHERE deletion_scheduled_at IS NOT NULL AND deletion_scheduled_at <= now() AND deleted_at IS NULL;

-- name: ReassignCoursesToDeletedUser :exec
UPDATE courses
SET owner_id = '00000000-0000-0000-0000-000000000001'
WHERE owner_id = $1 AND visibility = 'public';

-- name: DeletePrivateUnlistedCourses :exec
DELETE FROM courses
WHERE owner_id = $1 AND visibility != 'public';

-- name: DeleteUserRelatedData :exec
DELETE FROM saves WHERE user_id = $1;

-- name: DeleteUserCompletions :exec
DELETE FROM completions WHERE user_id = $1;

-- name: DeleteUserQuizAttempts :exec
DELETE FROM quiz_attempts WHERE user_id = $1;

-- name: DeleteUserOAuthAccounts :exec
DELETE FROM oauth_accounts WHERE user_id = $1;

-- name: DeleteUserMagicLinkTokens :exec
DELETE FROM magic_link_tokens WHERE email = (SELECT users.email FROM users WHERE users.id = $1);

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens WHERE user_id = $1;
