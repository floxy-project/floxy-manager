import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { CleanupModal } from '../components/CleanupModal';

interface SummaryStats {
  total_workflows: number;
  completed_workflows: number;
  failed_workflows: number;
  running_workflows: number;
  pending_workflows: number;
  active_workflows: number;
}

interface ActiveWorkflow {
  id: number;
  workflow_id: string;
  status: string;
  created_at: string;
  updated_at: string;
  duration_seconds: number;
  total_steps: number;
  completed_steps: number;
  failed_steps: number;
  running_steps: number;
  compensation_steps: number;
  rolled_back_steps: number;
}

export const Dashboard: React.FC = () => {
  const [summary, setSummary] = useState<SummaryStats | null>(null);
  const [activeWorkflows, setActiveWorkflows] = useState<ActiveWorkflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCleanupModal, setShowCleanupModal] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [summaryRes, activeRes] = await Promise.all([
          fetch('/api/stats/summary'),
          fetch('/api/instances/active')
        ]);

        if (!summaryRes.ok || !activeRes.ok) {
          throw new Error('Failed to fetch data');
        }

        const summaryData = await summaryRes.json();
        const activeData = await activeRes.json();

        setSummary(summaryData);
        setActiveWorkflows(activeData && Array.isArray(activeData) ? activeData : []);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleCleanup = (daysToKeep: number) => {
    // Refresh data after successful cleanup
    window.location.reload();
  };

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div>
      <h1>Floxy Dashboard</h1>
      
      {summary && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-number">{summary.total_workflows}</div>
            <div className="stat-label">Total Workflows</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{summary.active_workflows}</div>
            <div className="stat-label">Active Workflows</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{summary.completed_workflows}</div>
            <div className="stat-label">Completed</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{summary.failed_workflows}</div>
            <div className="stat-label">Failed</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{summary.running_workflows}</div>
            <div className="stat-label">Running</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{summary.pending_workflows}</div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
      )}

      {/* Administrative Actions */}
      <div className="card">
        <h2>Administrative Actions</h2>
        <p style={{ marginBottom: '1rem', color: '#6b7280' }}>
          Manage workflow instances and perform system maintenance tasks.
        </p>
        <button
          className="btn btn-danger"
          onClick={() => setShowCleanupModal(true)}
        >
          Cleanup Old Workflows
        </button>
      </div>

      <div className="card">
        <h2>Active Workflows</h2>
        {!activeWorkflows || activeWorkflows.length === 0 ? (
          <p>No active workflows</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Workflow</th>
                <th>Status</th>
                <th>Duration</th>
                <th>Steps</th>
                <th>Progress</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {activeWorkflows.map((workflow) => (
                <tr key={workflow.id}>
                  <td>{workflow.id}</td>
                  <td>{workflow.workflow_id}</td>
                  <td>
                    <span className={`status ${workflow.status}`}>
                      {workflow.status}
                    </span>
                  </td>
                  <td>{Math.round(workflow.duration_seconds)}s</td>
                  <td>
                    {workflow.completed_steps}/{workflow.total_steps}
                  </td>
                  <td>
                    <div style={{ width: '100px', backgroundColor: '#e9ecef', borderRadius: '4px' }}>
                      <div 
                        style={{ 
                          width: `${(workflow.completed_steps / workflow.total_steps) * 100}%`,
                          height: '8px',
                          backgroundColor: '#007bff',
                          borderRadius: '4px'
                        }}
                      />
                    </div>
                  </td>
                  <td>
                    <Link to={`/instances/${workflow.id}`} className="btn btn-primary">
                      View
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Cleanup modal */}
      <CleanupModal
        isOpen={showCleanupModal}
        onClose={() => setShowCleanupModal(false)}
        onCleanup={handleCleanup}
      />
    </div>
  );
};
