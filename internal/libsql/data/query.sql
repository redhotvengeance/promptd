-- name: ListMessages :many
SELECT * FROM messages
WHERE thread_id = ?
ORDER BY created_at ASC;

-- name: CreateMessage :exec
INSERT INTO messages (id, thread_id, role, content)
VALUES (?, ?, ?, ?);

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = ?;

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
