-- name: CreateFifMeta :one
INSERT INTO fifs (id, user_id, title, description )
VALUES(
    ?,
    ?,
    ?,
    ?
)RETURNING *;

-- name: GetFifById :one
SELECT * FROM fifs WHERE id = ?;

-- name: SetFifAudioUrl :one
UPDATE fifs SET audio_s3 = ?
WHERE id = ?
RETURNING *;

-- name: SetFifUrl :one
UPDATE fifs SET s3_url= ?
WHERE id = ?
RETURNING *;

-- name: UpdateFif :one
UPDATE fifs
SET
    title = ?,
    description = ?,
    user_id = ?,
    s3_url = ?
WHERE id = ?
RETURNING *;


-- name: GetFifsByUser :many
SELECT * FROM fifs WHERE user_id = ?;
