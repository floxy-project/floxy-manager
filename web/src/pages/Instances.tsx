import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

interface ActiveWorkflow {
  id: number;
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

export const Instances: React.FC = () => {
  const [instances, setInstances] = useState<ActiveWorkflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchInstances = async () => {
      try {
        const response = await fetch('/api/instances/active');
        if (!response.ok) {
          throw new Error('Failed to fetch instances');
        }
        const data = await response.json();
        setInstances(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchInstances();
  }, []);

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
          <span>Loading instances...</span>
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
      <h1>Active Workflow Instances</h1>
      
      <div className="card">
        {instances.length === 0 ? (
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <svg className="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
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
                    <tr key={instance.id}>
                      <td className="font-mono text-xs">{instance.id}</td>
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
                        <Link to={`/instances/${instance.id}`} className="btn btn-primary text-xs py-1.5 px-3">
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
      </div>
    </div>
  );
};
