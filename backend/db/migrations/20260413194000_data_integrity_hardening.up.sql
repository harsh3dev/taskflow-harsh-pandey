CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower_unique
    ON users (lower(email));

ALTER TABLE users
    ADD CONSTRAINT users_email_normalized CHECK (email = lower(btrim(email)));

ALTER TABLE projects
    ADD CONSTRAINT projects_description_non_blank CHECK (description IS NULL OR btrim(description) <> '');

ALTER TABLE tasks
    ADD CONSTRAINT tasks_description_non_blank CHECK (description IS NULL OR btrim(description) <> '');

ALTER TABLE auth_sessions
    ADD CONSTRAINT auth_sessions_expires_after_creation CHECK (expires_at > created_at);

CREATE OR REPLACE FUNCTION taskflow_set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_tasks_updated_at ON tasks;

CREATE TRIGGER set_tasks_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION taskflow_set_updated_at();
