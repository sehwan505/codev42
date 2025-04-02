-- create "projects" table
CREATE TABLE `projects` (
  `id` varchar(255) NOT NULL,
  `branch` varchar(100) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`, `branch`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "dev_plans" table
CREATE TABLE `dev_plans` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` varchar(255) NOT NULL,
  `branch` varchar(100) NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  `language` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_dev_plans_project` (`project_id`, `branch`),
  CONSTRAINT `fk_dev_plans_project` FOREIGN KEY (`project_id`, `branch`) REFERENCES `projects` (`id`, `branch`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "plans" table
CREATE TABLE `plans` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `dev_plan_id` bigint NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  `class_name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_dev_plans_plans` (`dev_plan_id`),
  CONSTRAINT `fk_dev_plans_plans` FOREIGN KEY (`dev_plan_id`) REFERENCES `dev_plans` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "annotations" table
CREATE TABLE `annotations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  `name` varchar(255) NOT NULL,
  `params` text NULL,
  `returns` text NULL,
  `description` text NULL,
  `plan_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_plans_annotations` (`plan_id`),
  CONSTRAINT `fk_plans_annotations` FOREIGN KEY (`plan_id`) REFERENCES `plans` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "files" table
CREATE TABLE `files` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` varchar(255) NULL,
  `project_branch` varchar(100) NULL,
  `file_path` longtext NOT NULL,
  `directory` longtext NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_projects_files` (`project_id`, `project_branch`),
  CONSTRAINT `fk_projects_files` FOREIGN KEY (`project_id`, `project_branch`) REFERENCES `projects` (`id`, `branch`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "codes" table
CREATE TABLE `codes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `file_id` bigint NOT NULL,
  `func_declaration` varchar(255) NOT NULL,
  `plan` text NULL,
  `code_chunk` text NOT NULL,
  `chunk_hash` varchar(64) NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_codes_file_id` (`file_id`),
  UNIQUE INDEX `uni_codes_chunk_hash` (`chunk_hash`),
  CONSTRAINT `fk_files_codes` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
