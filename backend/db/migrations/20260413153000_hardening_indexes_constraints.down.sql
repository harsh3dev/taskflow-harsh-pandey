DROP INDEX IF EXISTS idx_tasks_project_assignee;
DROP INDEX IF EXISTS idx_projects_owner_id;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_title_non_empty;

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_name_non_empty;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_non_empty;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_name_non_empty;
