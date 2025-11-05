-- Add membership.manage permission to project_manager role
insert into workflows_manager.role_permissions (role_id, permission_id)
values ('348a558f-b00a-419b-a8a4-699f6f8eae3f', 'bb1bc5da-fbe2-4b7d-8088-9ff7642034ce')
on conflict (role_id, permission_id) do nothing;
