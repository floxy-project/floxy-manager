-- Add project_id column to audit_log table
alter table workflows_manager.audit_log
    add column if not exists project_id integer references workflows_manager.projects(id);

-- Create index for project_id
create index if not exists idx_audit_log_project_id on workflows_manager.audit_log (project_id);

-- Update existing records: try to determine project_id from entity and entity_id
-- For projects: entity_id is the project_id
update workflows_manager.audit_log
set project_id = entity_id::integer
where entity = 'project' and entity_id ~ '^[0-9]+$';

-- For workflows: get project_id from project_workflows table
update workflows_manager.audit_log al
set project_id = (
    select pw.project_id
    from workflows_manager.project_workflows pw
    where pw.workflow_definition_id = al.entity_id
    limit 1
)
where al.entity = 'workflow' and al.project_id is null;

-- For memberships: get project_id from memberships table
update workflows_manager.audit_log al
set project_id = (
    select m.project_id
    from workflows_manager.memberships m
    where m.id::text = al.entity_id
    limit 1
)
where al.entity = 'membership' and al.project_id is null;

