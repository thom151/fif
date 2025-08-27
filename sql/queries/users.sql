-- name: CreateUser :one
INSERT INTO users (id, username, email, password_hash, avatar_url, voice_url, is_admin)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;


-- name: GetUserByEmail :one
SELECT * FROM users WHERE  email = ?;


-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;



-- name: SetUserAvatarAndVoiceID :one
UPDATE users SET avatar_url = ?, voice_url = ?, updated_at = ? WHERE id = ? RETURNING *;
