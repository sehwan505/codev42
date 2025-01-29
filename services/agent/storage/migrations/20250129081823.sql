-- Create "projects" table
CREATE TABLE `projects` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `description` text NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "files" table
CREATE TABLE `files` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint NULL,
  `file_path` longtext NOT NULL,
  `directory` longtext NOT NULL,
  `created_at` bigint unsigned NULL,
  `updated_at` bigint unsigned NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_projects_files` (`project_id`),
  CONSTRAINT `fk_projects_files` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "codes" table
CREATE TABLE `codes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `file_id` bigint NOT NULL,
  `func_name` varchar(255) NOT NULL,
  `plan` text NULL,
  `code_chunk` text NOT NULL,
  `chunk_hash` varchar(64) NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_codes_file_id` (`file_id`),
  UNIQUE INDEX `uni_codes_chunk_hash` (`chunk_hash`),
  CONSTRAINT `fk_files_codes` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
