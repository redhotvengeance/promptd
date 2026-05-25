-- Create "messages" table
CREATE TABLE `messages` (`id` text NULL, `thread_id` text NOT NULL, `role` text NOT NULL, `content` text NOT NULL, `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP), PRIMARY KEY (`id`), CONSTRAINT `0` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE);
