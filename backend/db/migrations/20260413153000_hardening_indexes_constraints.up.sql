CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_assignee ON tasks(project_id, assignee_id);

ALTER TABLE users
    ADD CONSTRAINT users_name_non_empty CHECK (btrim(name) <> '');

ALTER TABLE users
    ADD CONSTRAINT users_email_non_empty CHECK (btrim(email) <> '');

ALTER TABLE projects
    ADD CONSTRAINT projects_name_non_empty CHECK (btrim(name) <> '');

ALTER TABLE tasks
    ADD CONSTRAINT tasks_title_non_empty CHECK (btrim(title) <> '');
