import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

interface WorkflowDefinition {
  id: string;
  name: string;
  version: number;
  definition: any;
  created_at: string;
}

export const Workflows: React.FC = () => {
  const [workflows, setWorkflows] = useState<WorkflowDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchWorkflows = async () => {
      try {
        const response = await fetch('/api/workflows');
        if (!response.ok) {
          throw new Error('Failed to fetch workflows');
        }
        const data = await response.json();
        setWorkflows(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchWorkflows();
  }, []);

  if (loading) {
    return <div className="loading">Loading workflows...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div>
      <h1>Workflow Definitions</h1>
      
      <div className="card">
        {workflows.length === 0 ? (
          <p>No workflow definitions found</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Version</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {workflows.map((workflow) => (
                <tr key={workflow.id}>
                  <td>{workflow.name}</td>
                  <td>v{workflow.version}</td>
                  <td>{new Date(workflow.created_at).toLocaleString()}</td>
                  <td>
                    <Link to={`/workflows/${workflow.id}`} className="btn btn-primary">
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
