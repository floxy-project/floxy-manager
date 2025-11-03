import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CleanupModal } from '../components/CleanupModal';
import { authFetch } from '../utils/api';
import { useRBAC } from '../auth/permissions';

interface SummaryStats {
  total_workflows: number;
  completed_workflows: number;
  failed_workflows: number;
  running_workflows: number;
  pending_workflows: number;
  active_workflows: number;
}

interface ActiveWorkflow {
  instance_id: number;
  workflow_id: string;
  workflow_name: string;
  status: string;
  started_at: string;
  updated_at: string;
  current_step: string;
  total_steps: number;
  completed_steps: number;
  rolled_back_steps: number;
}

export const Dashboard: React.FC = () => {
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const [summary, setSummary] = useState<SummaryStats | null>(null);
  const rbac = useRBAC(projectId);
  const [activeWorkflows, setActiveWorkflows] = useState<ActiveWorkflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCleanupModal, setShowCleanupModal] = useState(false);

  useEffect(() => {
    if (tenantId && projectId) {
      fetchData();
    }
  }, [tenantId, projectId]);

  const fetchData = async () => {
    try {
      const [summaryRes, activeRes] = await Promise.all([
        authFetch(`/api/v1/stats?tenant_id=${tenantId}&project_id=${projectId}`),
        authFetch(`/api/v1/active-workflows?tenant_id=${tenantId}&project_id=${projectId}`)
      ]);

        if (!summaryRes.ok || !activeRes.ok) {
          throw new Error('Failed to fetch data');
        }

        const summaryDataRaw = await summaryRes.json();
        const activeDataRaw = await activeRes.json();

        const summaryData = Array.isArray(summaryDataRaw) ? summaryDataRaw : (summaryDataRaw.items || []);
        const activeData = Array.isArray(activeDataRaw) ? activeDataRaw : (activeDataRaw.items || []);

        const stats = Array.isArray(summaryData) ? summaryData : [];
        const totalWorkflows = stats.length;
        const totalInstances = stats.reduce((sum, stat) => sum + (stat.total_instances || 0), 0);
        const totalCompleted = stats.reduce((sum, stat) => sum + (stat.completed_instances || 0), 0);
        const totalFailed = stats.reduce((sum, stat) => sum + (stat.failed_instances || 0), 0);
        const totalRunning = stats.reduce((sum, stat) => sum + (stat.running_instances || 0), 0);
        const activeWorkflowsList = Array.isArray(activeData) ? activeData : [];
        const activeWorkflowsCount = activeWorkflowsList.length;
        
        setSummary({
          total_workflows: totalWorkflows,
          completed_workflows: totalCompleted,
          failed_workflows: totalFailed,
          running_workflows: totalRunning,
          pending_workflows: 0,
          active_workflows: activeWorkflowsCount,
        });

        setActiveWorkflows(activeWorkflowsList);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

  const handleCleanup = (daysToKeep: number) => {
    window.location.reload();
  };

  const formatDuration = (started: string, updated: string) => {
    const seconds = Math.round((new Date(updated).getTime() - new Date(started).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
    return `${Math.floor(seconds / 3600)}h`;
  };

  const getProgressPercentage = (completed: number, total: number) => {
    if (total === 0) return 0;
    return Math.round((completed / total) * 100);
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading dashboard...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error">
        <div className="flex items-center justify-center gap-2">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>Error: {error}</span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1>Dashboard</h1>
      </div>
      
      {summary && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-number">{summary.total_workflows}</div>
            <div className="stat-label">Total Workflows</div>
          </div>
          <div className="stat-card">
            <div className="stat-number text-blue-600 dark:text-blue-400">{summary.active_workflows}</div>
            <div className="stat-label">Active Workflows</div>
          </div>
          <div className="stat-card">
            <div className="stat-number text-emerald-600 dark:text-emerald-400">{summary.completed_workflows}</div>
            <div className="stat-label">Completed</div>
          </div>
          <div className="stat-card">
            <div className="stat-number text-red-600 dark:text-red-400">{summary.failed_workflows}</div>
            <div className="stat-label">Failed</div>
          </div>
          <div className="stat-card">
            <div className="stat-number text-amber-600 dark:text-amber-400">{summary.running_workflows}</div>
            <div className="stat-label">Running</div>
          </div>
          <div className="stat-card">
            <div className="stat-number text-slate-600 dark:text-[#ff4500]500">{summary.pending_workflows}</div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
      )}

      {rbac.canManageProject() && (
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2>Administrative Actions</h2>
          </div>
          <p className="text-slate-600 dark:text-[#ff4500]500 mb-4">
            Manage workflow instances and perform system maintenance tasks.
          </p>
          <button
            className="btn btn-danger"
            onClick={() => setShowCleanupModal(true)}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Cleanup Old Workflows
          </button>
        </div>
      )}

      <div className="card">
        <h2>Active Workflows</h2>
        {!activeWorkflows || activeWorkflows.length === 0 ? (
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <svg className="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p>No active workflows</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Workflow</th>
                  <th>Status</th>
                  <th>Duration</th>
                  <th>Progress</th>
                  <th>Steps</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {activeWorkflows.map((workflow) => {
                  const progress = getProgressPercentage(workflow.completed_steps, workflow.total_steps);
                  return (
                    <tr key={workflow.instance_id}>
                      <td className="font-mono text-xs">{workflow.instance_id}</td>
                      <td>
                        <div className="font-medium">{workflow.workflow_name || workflow.workflow_id}</div>
                        <div className="text-xs text-slate-500 dark:text-[#ff4500]500 font-mono mt-0.5">
                          {workflow.workflow_id}
                        </div>
                      </td>
                      <td>
                        <span className={`status ${workflow.status}`}>
                          {workflow.status}
                        </span>
                      </td>
                      <td className="font-mono text-sm">{formatDuration(workflow.started_at, workflow.updated_at)}</td>
                      <td>
                        <div className="flex items-center gap-2 min-w-[120px]">
                          <div className="progress-bar flex-1">
                            <div 
                              className={`progress-fill ${workflow.status === 'failed' || workflow.status === 'dlq' ? 'danger' : 'primary'}`}
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <span className="text-xs font-medium text-slate-600 dark:text-[#ff4500]500 min-w-[35px]">
                            {progress}%
                          </span>
                        </div>
                      </td>
                      <td className="text-sm">
                        <span className="font-medium">{workflow.completed_steps}</span>
                        <span className="text-slate-400"> / </span>
                        <span className="text-slate-600 dark:text-[#ff4500]500">{workflow.total_steps}</span>
                      </td>
                      <td>
                        <Link to={`/tenants/${tenantId}/projects/${projectId}/instances/${workflow.instance_id}`} className="btn btn-primary text-xs py-1.5 px-3">
                          View
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <CleanupModal
        isOpen={showCleanupModal}
        onClose={() => setShowCleanupModal(false)}
        onCleanup={handleCleanup}
        projectId={projectId || ''}
      />
    </div>
  );
};
