create schema workflows;

create table workflows.workflow_definitions
(
    id         text                                   not null
        primary key,
    name       text                                   not null,
    version    integer                                not null,
    definition jsonb                                  not null,
    created_at timestamp with time zone default now() not null,
    unique (name, version)
);

comment on table workflow_definitions is 'Workflow templates with definition of the execution graph';
comment on column workflow_definitions.definition is 'JSONB graph with adjacency list structure';
create index idx_workflow_definitions_name
    on workflow_definitions (name);

create table workflows.workflow_instances
(
    id           bigserial
        primary key,
    workflow_id  text                                   not null
        references workflow_definitions,
    status       text                                   not null
        constraint workflow_instances_status_check
            check (status = ANY
                   (ARRAY ['pending'::text, 'running'::text, 'completed'::text, 'failed'::text, 'rolling_back'::text, 'cancelled'::text, 'cancelling'::text, 'aborted'::text, 'dlq'::text])),
    input        jsonb,
    output       jsonb,
    error        text,
    started_at   timestamp with time zone,
    completed_at timestamp with time zone,
    created_at   timestamp with time zone default now() not null,
    updated_at   timestamp with time zone default now() not null
);

comment on table workflow_instances is 'Instances of running workflows';
comment on column workflow_instances.status is 'pending | running | completed | failed | rolling_back | cancelled | cancelling | aborted | dlq';
create index idx_workflow_instances_workflow_id
    on workflow_instances (workflow_id);
create index idx_workflow_instances_status
    on workflow_instances (status);
create index idx_workflow_instances_created_at
    on workflow_instances (created_at desc);

create table workflows.workflow_cancel_requests
(
    id           bigserial
        primary key,
    instance_id  bigint                                 not null
        unique
        references workflow_instances
            on delete cascade,
    requested_by text                                   not null,
    cancel_type  text                                   not null
        constraint workflow_cancel_requests_type_check
            check (cancel_type = ANY (ARRAY ['cancel'::text, 'abort'::text])),
    reason       text,
    created_at   timestamp with time zone default now() not null
);

create index idx_workflow_cancel_requests_instance_id
    on workflow_cancel_requests (instance_id);

create table workflows.workflow_dlq
(
    id          bigserial
        primary key,
    instance_id bigint                                 not null
        references workflow_instances
            on delete cascade,
    workflow_id text                                   not null
        references workflow_definitions,
    step_id     bigint                                 not null
        references workflow_steps
            on delete cascade,
    step_name   text                                   not null,
    step_type   text                                   not null,
    input       jsonb,
    error       text,
    reason      text,
    created_at  timestamp with time zone default now() not null
);

create index idx_dead_letter_instance_id
    on workflow_dlq (instance_id);
create index idx_dead_letter_workflow_id
    on workflow_dlq (workflow_id);
create index idx_dead_letter_created_at
    on workflow_dlq (created_at desc);

create table workflows.workflow_events
(
    id          bigserial
        primary key,
    instance_id bigint                                 not null
        references workflow_instances
            on delete cascade,
    step_id     bigint
        references workflow_steps
            on delete cascade,
    event_type  text                                   not null,
    payload     jsonb,
    created_at  timestamp with time zone default now() not null
);

comment on table workflow_events is 'Event log for auditing and debugging';
create index idx_workflow_events_instance_id
    on workflow_events (instance_id);
create index idx_workflow_events_created_at
    on workflow_events (created_at desc);
create index idx_workflow_events_event_type
    on workflow_events (event_type);

create table workflows.workflow_human_decisions
(
    id          bigserial
        primary key,
    instance_id bigint                                 not null
        references workflow_instances
            on delete cascade,
    step_id     bigint                                 not null
        references workflow_steps
            on delete cascade,
    decided_by  text                                   not null,
    decision    text                                   not null
        constraint workflow_human_decisions_decision_check
            check (decision = ANY (ARRAY ['confirmed'::text, 'rejected'::text])),
    comment     text,
    decided_at  timestamp with time zone default now() not null,
    created_at  timestamp with time zone default now() not null,
    unique (step_id, decided_by)
);

comment on table workflow_human_decisions is 'Stores user decisions for human-in-the-loop workflow steps (confirm/reject).';
comment on column workflow_human_decisions.decided_by is 'Arbitrary user identifier (e.g. username, email, external ID).';
comment on column workflow_human_decisions.decision is 'confirmed | rejected';
create index idx_workflow_human_decisions_instance_id
    on workflow_human_decisions (instance_id);
create index idx_workflow_human_decisions_step_id
    on workflow_human_decisions (step_id);

create table workflows.workflow_join_state
(
    id             bigserial
        primary key,
    instance_id    bigint                                       not null
        references workflow_instances
            on delete cascade,
    join_step_name text                                         not null,
    waiting_for    jsonb                                        not null,
    completed      jsonb                    default '[]'::jsonb not null,
    failed         jsonb                    default '[]'::jsonb not null,
    join_strategy  text                     default 'all'::text not null,
    is_ready       boolean                  default false       not null,
    created_at     timestamp with time zone default now()       not null,
    updated_at     timestamp with time zone default now()       not null,
    unique (instance_id, join_step_name)
);

create index idx_workflow_join_state_instance
    on workflow_join_state (instance_id);
create table workflows.workflow_queue
(
    id           bigserial
        primary key,
    instance_id  bigint                                 not null
        references workflow_instances
            on delete cascade,
    step_id      bigint
        references workflow_steps
            on delete cascade,
    scheduled_at timestamp with time zone default now() not null,
    attempted_at timestamp with time zone,
    attempted_by text,
    priority     integer                  default 0     not null
);

comment on table workflow_queue is 'Queue of steps for workers to complete';
create index idx_workflow_queue_scheduled
    on workflow_queue (scheduled_at asc, priority desc)
    where (attempted_at IS NULL);
create index idx_workflow_queue_instance_id
    on workflow_queue (instance_id);

-- auto-generated definition
create table workflows.workflow_steps
(
    id                       bigserial
        primary key,
    instance_id              bigint                                             not null
        references workflow_instances
            on delete cascade,
    step_name                text                                               not null,
    step_type                text                                               not null
        constraint workflow_steps_step_type_check
            check (step_type = ANY
                   (ARRAY ['task'::text, 'parallel'::text, 'condition'::text, 'fork'::text, 'join'::text, 'save_point'::text, 'human'::text])),
    status                   text                                               not null
        constraint workflow_steps_status_check
            check (status = ANY
                   (ARRAY ['pending'::text, 'running'::text, 'completed'::text, 'failed'::text, 'skipped'::text, 'compensation'::text, 'rolled_back'::text, 'waiting_decision'::text, 'confirmed'::text, 'rejected'::text, 'paused'::text])),
    input                    jsonb,
    output                   jsonb,
    error                    text,
    retry_count              integer                  default 0                 not null,
    max_retries              integer                  default 3                 not null,
    started_at               timestamp with time zone,
    completed_at             timestamp with time zone,
    created_at               timestamp with time zone default now()             not null,
    compensation_retry_count integer                  default 0                 not null,
    idempotency_key          uuid                     default gen_random_uuid() not null
);

comment on table workflow_steps is 'Separate workflow execution steps';
comment on column workflow_steps.step_type is 'task | parallel | condition | fork | join | save_point | human';
comment on column workflow_steps.status is 'pending | running | completed | failed | skipped | compensation | rolled_back | waiting_decision | confirmed | rejected | paused';
comment on column workflow_steps.compensation_retry_count is 'Number of compensation retries for this step';
create index idx_workflow_steps_instance_id
    on workflow_steps (instance_id);
create index idx_workflow_steps_status
    on workflow_steps (status);
create index idx_workflow_steps_step_name
    on workflow_steps (step_name);

create view workflows.active_workflows
            (id, workflow_id, status, created_at, updated_at, duration_seconds, total_steps, completed_steps,
             failed_steps, running_steps)
as
SELECT wi.id,
       wi.workflow_id,
       wi.status,
       wi.created_at,
       wi.updated_at,
       EXTRACT(epoch FROM now() - wi.created_at)                 AS duration_seconds,
       count(ws.id)                                              AS total_steps,
       count(ws.id) FILTER (WHERE ws.status = 'completed'::text) AS completed_steps,
       count(ws.id) FILTER (WHERE ws.status = 'failed'::text)    AS failed_steps,
       count(ws.id) FILTER (WHERE ws.status = 'running'::text)   AS running_steps
FROM workflows.workflow_instances wi
         LEFT JOIN workflows.workflow_steps ws ON wi.id = ws.instance_id
WHERE wi.status = ANY (ARRAY ['pending'::text, 'running'::text, 'dlq'::text])
GROUP BY wi.id, wi.workflow_id, wi.status, wi.created_at, wi.updated_at;

create view workflows.workflow_stats (name, version, total_instances, completed, failed, running, avg_duration_seconds) as
SELECT wd.name,
       wd.version,
       count(wi.id)                                                                                          AS total_instances,
       count(wi.id) FILTER (WHERE wi.status = 'completed'::text)                                             AS completed,
       count(wi.id) FILTER (WHERE wi.status = 'failed'::text)                                                AS failed,
       count(wi.id) FILTER (WHERE wi.status = 'running'::text)                                               AS running,
       avg(EXTRACT(epoch FROM wi.completed_at - wi.created_at))
       FILTER (WHERE wi.status = 'completed'::text)                                                          AS avg_duration_seconds
FROM workflows.workflow_definitions wd
         LEFT JOIN workflows.workflow_instances wi ON wd.id = wi.workflow_id
GROUP BY wd.name, wd.version;