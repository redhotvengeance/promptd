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

-- name: DeleteWorkspace :exec
DELETE FROM workspace_chunks WHERE workspace_path = ?;

-- name: InsertChunk :exec
INSERT INTO workspace_chunks (id, workspace_path, filepath, content, embedding)
VALUES (?, ?, ?, ?, vector32(?));

-- name: SearchChunks :many
SELECT id, filepath, content, vector_distance_cos(embedding, vector32(?)) as distance
FROM workspace_chunks
WHERE workspace_path = ?
ORDER BY distance ASC
LIMIT ?;
