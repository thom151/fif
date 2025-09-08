-- name: CreateMusicMeta :one
INSERT INTO musics(id, user_id, title, description )
VALUES(
    ?,
    ?,
    ?,
    ?
)RETURNING *;

-- name: GetMusicById :one
SELECT * FROM musics WHERE id = ?;

-- name: UpdateMusic :exec
UPDATE musics
SET
    title = ?,
    description = ?,
    user_id = ?,
    s3_url = ?
WHERE id = ?;


-- name: GetMusicByUser :many
SELECT * FROM musics WHERE user_id = ?;


-- name: DeleteMusic :one
DELETE FROM musics WHERE id = ? AND user_id = ? RETURNING *;
