create table if not exists workflows.product_info
(
    id         serial
        primary key,
    key        varchar(255) not null
        unique,
    value      text         not null,
    created_at timestamp with time zone default now()
);

create table if not exists workflows.license
(
    id           uuid                                   not null
        primary key,
    license_text text                                   not null,
    issued_at    timestamp with time zone               not null,
    expires_at   timestamp with time zone               not null,
    client_id    uuid                                   not null,
    type         varchar(50)                            not null,
    created_at   timestamp with time zone default now() not null,
    constraint license_dates_range
        check (issued_at <= expires_at)
);

create table if not exists workflows.license_history
(
    id           uuid                                   not null
        primary key,
    license_id   uuid                                   not null
        references workflows.license,
    license_text text                                   not null,
    issued_at    timestamp with time zone               not null,
    expires_at   timestamp with time zone               not null,
    client_id    uuid                                   not null,
    type         varchar(50)                            not null,
    created_at   timestamp with time zone default now() not null,
    constraint license_history_dates_range
        check (issued_at <= expires_at)
);

create table if not exists workflows.tenants
(
    id         serial
        primary key,
    name       varchar(255) not null
        unique,
    created_at timestamp with time zone default now()
);

create table if not exists workflows.projects
(
    id          serial not null primary key,
    name        varchar(128)                                       not null
        unique,
    description varchar(300),
    tenant_id   integer                                            not null
        references workflows.tenants
            on delete restrict,
    archived    boolean                  default false             not null,
    created_at  timestamp with time zone default now()             not null,
    updated_at  timestamp with time zone default now()             not null,
    archived_at timestamp with time zone
);

create table if not exists workflows.users
(
    id                  serial
        primary key,
    username            varchar(255)                                       not null
        unique,
    email               varchar(255)                                       not null
        unique,
    password_hash       varchar(255)                                       not null,
    is_superuser        boolean                  default false             not null,
    is_active           boolean                  default true              not null,
    created_at          timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at          timestamp with time zone default CURRENT_TIMESTAMP not null,
    last_login          timestamp with time zone default CURRENT_TIMESTAMP,
    is_tmp_password     boolean                  default true,
    two_fa_enabled      boolean                  default false             not null,
    two_fa_secret       text,
    two_fa_confirmed_at timestamp with time zone,
    is_external         boolean                  default false             not null,
    license_accepted    boolean                  default false             not null
);

create table if not exists workflows.roles
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    key         varchar(50)                                        not null
        unique,
    name        varchar(50)                                        not null,
    description varchar(300),
    created_at  timestamp with time zone default now()             not null
);

create table workflows.permissions
(
    id   uuid default gen_random_uuid() not null
        primary key,
    key  varchar(50)                    not null
        unique,
    name varchar(50)                    not null
);

create table if not exists workflows.role_permissions
(
    id            uuid default gen_random_uuid() not null
        primary key,
    role_id       uuid                           not null
        references workflows.roles
            on delete cascade,
    permission_id uuid                           not null
        references workflows.permissions
            on delete cascade,
    constraint role_permissions_unique
        unique (role_id, permission_id)
);

insert into workflows.roles (id, key, name, description, created_at)
values  ('c80633be-e36b-4fb8-8d5d-f4b05c449d37', 'project_owner', 'Project Owner', 'Full control of project', '2025-10-14 06:16:14.104571 +00:00'),
        ('348a558f-b00a-419b-a8a4-699f6f8eae3f', 'project_manager', 'Project Manager', 'Manage workflows', '2025-10-14 06:16:14.104571 +00:00'),
        ('fb1c0bf4-65bf-4141-8723-170dffb797fa', 'project_viewer', 'Project Viewer', 'Read-only', '2025-10-14 06:16:14.104571 +00:00')
on conflict (key) do nothing;

insert into workflows.permissions (id, key, name)
values  ('b0ca1ed0-07aa-4fa2-a843-8253799240ae', 'project.view', 'View project'),
        ('4465a5a3-ea1e-4ed8-a905-217368483ac4', 'project.manage', 'Manage project'),
        ('a6f3cc48-0698-4f12-8e89-7a9e02fd30b2', 'audit.view', 'View audit'),
        ('bb1bc5da-fbe2-4b7d-8088-9ff7642034ce', 'membership.manage', 'Manage memberships'),
        ('c100b088-0163-4ae1-82d6-f2eeace0621a', 'project.create', 'Create projects')
on conflict (key) do nothing;

insert into workflows.role_permissions (id, role_id, permission_id)
values  ('c243ca60-1307-4b92-83ea-a1c22791968b', 'c80633be-e36b-4fb8-8d5d-f4b05c449d37', 'b0ca1ed0-07aa-4fa2-a843-8253799240ae'),
        ('c2a2145e-d598-4e73-bf63-00cd0bfdb015', '348a558f-b00a-419b-a8a4-699f6f8eae3f', 'b0ca1ed0-07aa-4fa2-a843-8253799240ae'),
        ('a1545c30-9929-4594-9653-1af6072b92af', 'fb1c0bf4-65bf-4141-8723-170dffb797fa', 'b0ca1ed0-07aa-4fa2-a843-8253799240ae'),
        ('80468ccf-dd5d-4ddd-85b1-34465574b829', 'c80633be-e36b-4fb8-8d5d-f4b05c449d37', '4465a5a3-ea1e-4ed8-a905-217368483ac4'),
        ('e3c4b44b-ca8e-44e6-b6d9-e2cf2a3515fc', '348a558f-b00a-419b-a8a4-699f6f8eae3f', '4465a5a3-ea1e-4ed8-a905-217368483ac4'),
        ('734e6b92-7aaf-4091-9643-d542b9e09f9e', 'c80633be-e36b-4fb8-8d5d-f4b05c449d37', 'a6f3cc48-0698-4f12-8e89-7a9e02fd30b2'),
        ('98de5120-230b-45b8-b3dd-3bf61d5471c2', '348a558f-b00a-419b-a8a4-699f6f8eae3f', 'a6f3cc48-0698-4f12-8e89-7a9e02fd30b2'),
        ('598319f4-4ff9-47bf-9ab5-ea50cb929587', 'c80633be-e36b-4fb8-8d5d-f4b05c449d37', 'bb1bc5da-fbe2-4b7d-8088-9ff7642034ce')
on conflict (id) do nothing;

create table if not exists workflows.memberships
(
    id         serial not null primary key,
    project_id integer                                               not null
        references workflows.projects
            on delete cascade,
    user_id    integer                                            not null
        references workflows.users
            on delete cascade,
    role_id    uuid                                               not null
        references workflows.roles
            on delete restrict,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone default now()             not null,
    constraint membership_unique
        unique (project_id, user_id)
);

create index if not exists idx_memberships_project_id
    on workflows.memberships (project_id);

create index if not exists idx_memberships_user_id
    on workflows.memberships (user_id);

create index if not exists idx_memberships_role_id
    on workflows.memberships (role_id);

create table if not exists workflows.membership_audit
(
    id            bigserial
        primary key,
    membership_id uuid,
    actor_user_id integer,
    action        varchar(50)                            not null,
    old_value     jsonb,
    new_value     jsonb,
    created_at    timestamp with time zone default now() not null
);

create table if not exists workflows.ldap_sync_stats
(
    id              serial
        primary key,
    sync_session_id uuid                                                          not null
        unique,
    start_time      timestamp with time zone default now()                        not null,
    end_time        timestamp with time zone,
    duration        varchar(50),
    total_users     integer                  default 0                            not null,
    synced_users    integer                  default 0                            not null,
    errors          integer                  default 0                            not null,
    warnings        integer                  default 0                            not null,
    status          varchar(20)              default 'running'::character varying not null,
    error_message   text
);

create index if not exists idx_ldap_sync_stats_sync_session_id
    on workflows.ldap_sync_stats (sync_session_id);

create index if not exists idx_ldap_sync_stats_start_time
    on workflows.ldap_sync_stats (start_time);

create index if not exists idx_ldap_sync_stats_status
    on workflows.ldap_sync_stats (status);

create table if not exists workflows.ldap_sync_logs
(
    id                 serial
        primary key,
    timestamp          timestamp with time zone default now() not null,
    level              varchar(10)                            not null,
    message            text                                   not null,
    username           varchar(255),
    details            text,
    sync_session_id    uuid                                   not null,
    stack_trace        text,
    ldap_error_code    integer,
    ldap_error_message text
);

create index if not exists idx_ldap_sync_logs_timestamp
    on workflows.ldap_sync_logs (timestamp);

create index if not exists idx_ldap_sync_logs_level
    on workflows.ldap_sync_logs (level);

create index if not exists idx_ldap_sync_logs_sync_session_id
    on workflows.ldap_sync_logs (sync_session_id);

create index if not exists idx_ldap_sync_logs_username
    on workflows.ldap_sync_logs (username);

create or replace view workflows.v_user_project_permissions(user_id, project_id, role_key, permissions) as
SELECT m.user_id,
       m.project_id,
       r.key                                                                    AS role_key,
       json_agg(json_build_object('key', p.key, 'name', p.name) ORDER BY p.key) AS permissions
FROM workflows.memberships m
         JOIN workflows.roles r ON r.id = m.role_id
         LEFT JOIN workflows.role_permissions rp ON rp.role_id = r.id
         LEFT JOIN workflows.permissions p ON p.id = rp.permission_id
GROUP BY m.user_id, m.project_id, r.key
ORDER BY m.user_id, m.project_id;

create or replace view workflows.v_role_permissions(role_id, role_key, role_name, permissions) as
SELECT r.id                                                                     AS role_id,
       r.key                                                                    AS role_key,
       r.name                                                                   AS role_name,
       json_agg(json_build_object('key', p.key, 'name', p.name) ORDER BY p.key) AS permissions
FROM workflows.roles r
         LEFT JOIN workflows.role_permissions rp ON rp.role_id = r.id
         LEFT JOIN workflows.permissions p ON p.id = rp.permission_id
GROUP BY r.id, r.key, r.name
ORDER BY r.key;
