import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { authFetch } from '../utils/api';
import { AlertCircle, Clipboard, ChevronLeft, ChevronRight, Loader2 } from 'lucide-react';

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

interface InstancesResponse {
  items: ActiveWorkflow[];
  page: number;
  page_size: number;
  total: number;
}

export const Instances: React.FC = () => {
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const [instances, setInstances] = useState<ActiveWorkflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [totalItems, setTotalItems] = useState(0);

  useEffect(() => {
    if (tenantId && projectId) {
      fetchInstances();
    }
  }, [tenantId, projectId, currentPage]);

  const fetchInstances = async () => {
    try {
      setLoading(true);
      const response = await authFetch(`/api/v1/active-workflows?tenant_id=${tenantId}&project_id=${projectId}&page=${currentPage}&page_size=${pageSize}`);
      if (!response.ok) {
        throw new Error('Failed to fetch instances');
      }
      const data: InstancesResponse = await response.json();
      setInstances(data.items);
      setTotalItems(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const totalPages = Math.ceil(totalItems / pageSize);

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
          <Loader2 className="w-5 h-5 animate-spin" />
          <span>Loading instances...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error">
        <div className="flex items-center justify-center gap-2">
          <AlertCircle className="w-5 h-5" />
          <span>Error: {error}</span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <h1>Active Workflow Instances</h1>
      
      <div className="card">
        {instances.length === 0 ? (
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <Clipboard className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p>No active instances found</p>
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
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {instances.map((instance) => {
                  const progress = getProgressPercentage(instance.completed_steps, instance.total_steps);
                  return (
                    <tr key={instance.instance_id}>
                      <td className="font-mono text-xs">{instance.instance_id}</td>
                      <td>
                        <div className="font-medium">{instance.workflow_name || instance.workflow_id}</div>
                        <div className="text-xs text-slate-500 dark:text-[#ff4500]500 font-mono mt-0.5">
                          {instance.workflow_id}
                        </div>
                      </td>
                      <td>
                        <span className={`status ${instance.status}`}>
                          {instance.status}
                        </span>
                      </td>
                      <td className="font-mono text-sm">{formatDuration(instance.started_at, instance.updated_at)}</td>
                      <td>
                        <div className="flex items-center gap-2 min-w-[120px]">
                          <div className="progress-bar flex-1">
                            <div 
                              className={`progress-fill ${instance.status === 'dlq' || instance.status === 'failed' ? 'danger' : 'primary'}`}
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <span className="text-xs font-medium text-slate-600 dark:text-[#ff4500]500 min-w-[35px]">
                            {progress}%
                          </span>
                        </div>
                      </td>
                      <td className="text-sm">
                        <span className="font-medium">{instance.completed_steps}</span>
                        <span className="text-slate-400"> / </span>
                        <span className="text-slate-600 dark:text-[#ff4500]500">{instance.total_steps}</span>
                        {instance.rolled_back_steps > 0 && (
                          <div className="text-xs text-amber-600 dark:text-amber-400 mt-0.5">
                            ({instance.rolled_back_steps} rolled back)
                          </div>
                        )}
                      </td>
                      <td className="text-sm text-slate-600 dark:text-[#ff4500]500">
                        {new Date(instance.started_at).toLocaleString()}
                      </td>
                      <td>
                        <Link 
                          to={`/tenants/${tenantId}/projects/${projectId}/instances/${instance.instance_id}`} 
                          className="btn btn-primary text-xs py-1.5 px-3"
                        >
                          View Details
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-4 mt-6 pt-6 border-t border-slate-200 dark:border-slate-700">
            <button
              className="btn btn-outline"
              onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
              disabled={currentPage === 1}
            >
              <ChevronLeft className="w-4 h-4" />
              Previous
            </button>
            <div className="px-4 py-2 rounded-lg" style={{
              background: 'rgba(255, 255, 255, 0.5)',
              backdropFilter: 'blur(12px) saturate(150%)',
              WebkitBackdropFilter: 'blur(12px) saturate(150%)',
              border: '1px solid rgba(255, 255, 255, 0.4)',
              boxShadow: '0 4px 16px rgba(0, 0, 0, 0.06)'
            }}>
              <span className="text-sm font-medium text-slate-700 dark:text-[#ff4500]400">
                Page {currentPage} of {totalPages}
              </span>
            </div>
            <button
              className="btn btn-outline"
              onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
              disabled={currentPage === totalPages}
            >
              Next
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
};
