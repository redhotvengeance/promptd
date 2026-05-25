-- Create "threads" table
CREATE TABLE `threads` (`id` text NULL, `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP), `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP), PRIMARY KEY (`id`));
