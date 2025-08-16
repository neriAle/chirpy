-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    CAST(DATEADD(DAY, 60, GETDATE()) as DATE),
    $2
)
RETURNING token;

-- name: GetUserFromRefreshToken :one
SELECT user_id, expires_at FROM refresh_tokens
WHERE token = $1 LIMIT 1;