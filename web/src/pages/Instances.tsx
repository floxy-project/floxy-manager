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
                  <td>{Math.round((new Date(instance.updated_at).getTime() - new Date(instance.started_at).getTime()) / 1000)}s</td>
                  <td>
                    {instance.completed_steps}/{instance.total_steps}
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
                          backgroundColor: instance.status === 'dlq' ? '#dc3545' : '#007bff',
                          borderRadius: '4px'
                        }}
                      />
                    </div>
                  </td>
                  <td>{new Date(instance.started_at).toLocaleString()}</td>
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
