INSERT INTO workflows.project_workflows (project_id, workflow_definition_id)
SELECT p.id, wd.id
FROM workflows.projects p
JOIN workflows.workflow_definitions wd ON TRUE
WHERE p.name = 'Default'
  AND NOT EXISTS (
    SELECT 1
    FROM workflows.project_workflows pw
    WHERE pw.project_id = p.id
      AND pw.workflow_definition_id = wd.id
);
