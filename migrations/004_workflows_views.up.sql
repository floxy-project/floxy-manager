-- v_workflow_definitions
create or replace view workflows_manager.v_workflow_definitions as
select
    p.tenant_id,
    pw.project_id,
    wd.*
from workflows.workflow_definitions wd
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_instances
create or replace view workflows_manager.v_workflow_instances as
select
    p.tenant_id,
    pw.project_id,
    wi.*
from workflows.workflow_instances wi
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_steps
create or replace view workflows_manager.v_workflow_steps as
select
    p.tenant_id,
    pw.project_id,
    ws.*
from workflows.workflow_steps ws
         join workflows.workflow_instances wi
              on wi.id = ws.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_events
create or replace view workflows_manager.v_workflow_events as
select
    p.tenant_id,
    pw.project_id,
    we.*
from workflows.workflow_events we
         join workflows.workflow_instances wi
              on wi.id = we.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_human_decisions
create or replace view workflows_manager.v_workflow_human_decisions as
select
    p.tenant_id,
    pw.project_id,
    whd.*
from workflows.workflow_human_decisions whd
         join workflows.workflow_instances wi
              on wi.id = whd.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_join_state
create or replace view workflows_manager.v_workflow_join_state as
select
    p.tenant_id,
    pw.project_id,
    wjs.*
from workflows.workflow_join_state wjs
         join workflows.workflow_instances wi
              on wi.id = wjs.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_queue
create or replace view workflows_manager.v_workflow_queue as
select
    p.tenant_id,
    pw.project_id,
    wq.*
from workflows.workflow_queue wq
         join workflows.workflow_instances wi
              on wi.id = wq.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_cancel_requests
create or replace view workflows_manager.v_workflow_cancel_requests as
select
    p.tenant_id,
    pw.project_id,
    wcr.*
from workflows.workflow_cancel_requests wcr
         join workflows.workflow_instances wi
              on wi.id = wcr.instance_id
         join workflows.workflow_definitions wd
              on wd.id = wi.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_dlq
create or replace view workflows_manager.v_workflow_dlq as
select
    p.tenant_id,
    pw.project_id,
    wdlq.*
from workflows.workflow_dlq wdlq
         join workflows.workflow_definitions wd
              on wd.id = wdlq.workflow_id
         join workflows.workflow_instances wi
              on wi.id = wdlq.instance_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_active_workflows
create or replace view workflows_manager.v_active_workflows as
select
    p.tenant_id,
    pw.project_id,
    aw.*
from workflows.active_workflows aw
         join workflows.workflow_definitions wd
              on wd.id = aw.workflow_id
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;

-- v_workflow_stats
create or replace view workflows_manager.v_workflow_stats as
select
    p.tenant_id,
    pw.project_id,
    ws.*
from workflows.workflow_stats ws
         join workflows.workflow_definitions wd
              on wd.name = ws.name and wd.version = ws.version
         join workflows_manager.project_workflows pw
              on pw.workflow_definition_id = wd.id
         join workflows_manager.projects p
              on p.id = pw.project_id;
