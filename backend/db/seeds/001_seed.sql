-- Seed credentials for local testing:
-- email: test@example.com
-- password: password123

INSERT INTO users (id, name, email, password, created_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Test User',
    'test@example.com',
    '$2y$12$g.GAsEN2lKtF0.qCj9Iw0eBiIYDjNxPsqUbI8zmbJq3QxTiKGSWji',
    NOW()
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id, created_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'TaskFlow Demo Project',
    'Seeded project for reviewer walkthroughs.',
    '11111111-1111-1111-1111-111111111111',
    NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (
    id,
    title,
    description,
    status,
    priority,
    project_id,
    assignee_id,
    creator_id,
    due_date,
    created_at,
    updated_at
)
VALUES
    (
        '33333333-3333-3333-3333-333333333331',
        'Draft API contract',
        'Prepare the initial auth and projects API contract.',
        'todo',
        'high',
        '22222222-2222-2222-2222-222222222222',
        '11111111-1111-1111-1111-111111111111',
        '11111111-1111-1111-1111-111111111111',
        CURRENT_DATE + 3,
        NOW(),
        NOW()
    ),
    (
        '33333333-3333-3333-3333-333333333332',
        'Build Kanban layout',
        'Create the three-column project detail layout.',
        'in_progress',
        'medium',
        '22222222-2222-2222-2222-222222222222',
        '11111111-1111-1111-1111-111111111111',
        '11111111-1111-1111-1111-111111111111',
        CURRENT_DATE + 5,
        NOW(),
        NOW()
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'Prepare Docker baseline',
        'Document the initial container plan and env vars.',
        'done',
        'low',
        '22222222-2222-2222-2222-222222222222',
        '11111111-1111-1111-1111-111111111111',
        '11111111-1111-1111-1111-111111111111',
        CURRENT_DATE + 7,
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;
