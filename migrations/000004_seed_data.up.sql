INSERT INTO users (id, name, email, password)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'Test User',
  'test@example.com',
  '$2a$12$Uhd7XjpZfm8y.fc.ybua1uyMO.wW4crRY9SklgsF/S3guFI7hF5wO'
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id)
VALUES (
  '22222222-2222-2222-2222-222222222222',
  'Seed Project',
  'Project created by seed migration',
  '11111111-1111-1111-1111-111111111111'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, due_date, created_by)
VALUES
  (
    '33333333-3333-3333-3333-333333333331',
    'Prepare API docs',
    'Document auth and task endpoints',
    'todo',
    'medium',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    CURRENT_DATE + INTERVAL '3 days',
    '11111111-1111-1111-1111-111111111111'
  ),
  (
    '33333333-3333-3333-3333-333333333332',
    'Implement task filters',
    'Add status and assignee query support',
    'in_progress',
    'high',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    CURRENT_DATE + INTERVAL '5 days',
    '11111111-1111-1111-1111-111111111111'
  ),
  (
    '33333333-3333-3333-3333-333333333333',
    'Ship backend v1',
    'Finalize core CRUD endpoints',
    'done',
    'low',
    '22222222-2222-2222-2222-222222222222',
    NULL,
    CURRENT_DATE + INTERVAL '7 days',
    '11111111-1111-1111-1111-111111111111'
  )
ON CONFLICT (id) DO NOTHING;
