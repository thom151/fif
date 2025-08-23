-- name: CreateBrollMeta :one
INSERT INTO brolls (id, user_id, title, description )
VALUES(
    ?,
    ?,
    ?,
    ?
)RETURNING *;

-- name: GetBrollById :one
SELECT * FROM brolls WHERE id = ?;

-- name: UpdateBroll :exec
UPDATE brolls
SET
    title = ?,
    description = ?,
    user_id = ?,
    s3_url = ?
WHERE id = ?;


-- name: GetBrollByUser :many
SELECT * FROM brolls WHERE user_id = ?;


-- name: DeleteBroll :one
DELETE FROM brolls WHERE id = ? AND user_id = ? RETURNING *;
