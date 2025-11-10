-- Add workflow.create permission
insert into workflows_manager.permissions (id, key, name)
values ('d2e8f9a1-3b4c-4d5e-6f7a-8b9c0d1e2f3a', 'workflow.create', 'Create workflows')
on conflict (key) do nothing;

-- Add project_developer role
insert into workflows_manager.roles (id, key, name, description)
values ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 'project_developer', 'Project Developer', 'View project and create workflows')
on conflict (key) do nothing;

-- Grant project.view and workflow.create to project_developer role
insert into workflows_manager.role_permissions (role_id, permission_id)
values 
    ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 'b0ca1ed0-07aa-4fa2-a843-8253799240ae'), -- project.view
    ('a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d', 'd2e8f9a1-3b4c-4d5e-6f7a-8b9c0d1e2f3a')  -- workflow.create
on conflict (role_id, permission_id) do nothing;

