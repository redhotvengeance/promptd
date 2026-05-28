-- Create "workspace_chunks" table
CREATE TABLE `workspace_chunks` (`id` text NULL, `workspace_path` text NOT NULL, `filepath` text NOT NULL, `content` text NOT NULL, `embedding` F32_BLOB NULL, `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP), PRIMARY KEY (`id`));
-- Create index "idx_workspace_chunks_workspace" to table: "workspace_chunks"
CREATE INDEX `idx_workspace_chunks_workspace` ON `workspace_chunks` (`workspace_path`);
