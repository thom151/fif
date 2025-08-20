-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at) 
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;


-- name: GetToken :one
SELECT * FROM refresh_tokens WHERE token = ?;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = ?, 
updated_at = ? 
where token = ?
RETURNING * ;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = ?
AND revoked_at IS NULL
AND expires_at > ?;
