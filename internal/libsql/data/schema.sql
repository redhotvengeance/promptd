CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  thread_id TEXT NOT NULL,
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

CREATE TABLE threads (
  id TEXT PRIMARY KEY,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE workspace_chunks (
  id TEXT PRIMARY KEY,
  workspace_path TEXT NOT NULL,
  filepath TEXT NOT NULL,
  content TEXT NOT NULL,
  embedding F32_BLOB,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workspace_chunks_workspace on workspace_chunks(workspace_path);
