DROP TRIGGER IF EXISTS set_tasks_updated_at ON tasks;

DROP FUNCTION IF EXISTS taskflow_set_updated_at();

DROP INDEX IF EXISTS idx_users_email_lower_unique;

ALTER TABLE auth_sessions
    DROP CONSTRAINT IF EXISTS auth_sessions_expires_after_creation;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_description_non_blank;

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_description_non_blank;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_normalized;
