ALTER TABLE tasks
ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL;

UPDATE tasks t
SET created_by = p.owner_id
FROM projects p
WHERE t.project_id = p.id
  AND t.created_by IS NULL;
