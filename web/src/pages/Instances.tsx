import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

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

  if (loading) {
    return <div className="loading">Loading instances...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div>
      <h1>Active Workflow Instances</h1>
      
      <div className="card">
        {instances.length === 0 ? (
          <p>No active instances found</p>
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
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {instances.map((instance) => (
                <tr key={instance.id}>
                  <td>{instance.id}</td>
                  <td>{instance.workflow_id}</td>
                  <td>
                    <span className={`status ${instance.status}`}>
                      {instance.status}
                    </span>
                  </td>
                  <td>{Math.round(instance.duration_seconds)}s</td>
                  <td>
                    {instance.completed_steps}/{instance.total_steps}
                    {instance.failed_steps > 0 && (
                      <span style={{ color: '#dc3545', marginLeft: '0.5rem' }}>
                        ({instance.failed_steps} failed)
                      </span>
                    )}
                    {instance.compensation_steps > 0 && (
                      <span style={{ color: '#6f42c1', marginLeft: '0.5rem' }}>
                        ({instance.compensation_steps} compensation)
                      </span>
                    )}
                    {instance.rolled_back_steps > 0 && (
                      <span style={{ color: '#fd7e14', marginLeft: '0.5rem' }}>
                        ({instance.rolled_back_steps} rolled back)
                      </span>
                    )}
                  </td>
                  <td>
                    <div style={{ width: '100px', backgroundColor: '#e9ecef', borderRadius: '4px' }}>
                      <div 
                        style={{ 
                          width: `${(instance.completed_steps / instance.total_steps) * 100}%`,
                          height: '8px',
                          backgroundColor: instance.failed_steps > 0 ? '#dc3545' : '#007bff',
                          borderRadius: '4px'
                        }}
                      />
                    </div>
                  </td>
                  <td>{new Date(instance.created_at).toLocaleString()}</td>
                  <td>
                    <Link to={`/instances/${instance.id}`} className="btn btn-primary">
                      View Details
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
