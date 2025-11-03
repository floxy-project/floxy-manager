insert into workflows.projects (name, tenant_id, created_at) values ('Default', (SELECT id FROM workflows.tenants ORDER BY id DESC LIMIT 1), now()) on conflict (name) do nothing;

create table if not exists workflows.project_workflows (
    project_id integer references workflows.projects(id) not null,
    workflow_definition_id text references workflows.workflow_definitions(id) not null,
    primary key (project_id, workflow_definition_id)
);
