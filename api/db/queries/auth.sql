-- name: GetOAuthAccount :one
SELECT id, user_id, provider, provider_id, created_at
FROM oauth_accounts
WHERE provider = $1 AND provider_id = $2;

-- name: CreateOAuthAccount :one
INSERT INTO oauth_accounts (user_id, provider, provider_id)
VALUES ($1, $2, $3)
RETURNING id, user_id, provider, provider_id, created_at;

-- name: CreateMagicLinkToken :one
INSERT INTO magic_link_tokens (email, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, email, token_hash, expires_at, used_at, created_at;

-- name: GetMagicLinkTokenByHash :one
SELECT id, email, token_hash, expires_at, used_at, created_at
FROM magic_link_tokens
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now();

-- name: MarkMagicLinkTokenUsed :exec
UPDATE magic_link_tokens SET used_at = now() WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;
