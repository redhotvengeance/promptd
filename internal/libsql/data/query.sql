-- name: GetThread :one
SELECT * FROM threads
WHERE id = ?;

-- name: ListThreads :many
SELECT * FROM threads
ORDER BY updated_at ASC;

-- name: CreateThread :exec
INSERT INTO threads (id) VALUES (?);

-- name: DeleteThread :exec
DELETE FROM threads WHERE id = ?;
