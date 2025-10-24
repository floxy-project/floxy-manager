import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { WorkflowGraph } from '../components/WorkflowGraph';

interface WorkflowDefinition {
  id: string;
  name: string;
  version: number;
  definition: any;
  created_at: string;
}

interface WorkflowInstance {
  id: number;
  workflow_id: string;
  status: string;
  input: any;
  output: any;
  error: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export const WorkflowDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [workflow, setWorkflow] = useState<WorkflowDefinition | null>(null);
  const [instances, setInstances] = useState<WorkflowInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [workflowRes, instancesRes] = await Promise.all([
          fetch(`/api/workflows/${id}`),
          fetch(`/api/workflows/${id}/instances`)
        ]);

        if (!workflowRes.ok) {
          throw new Error('Failed to fetch workflow');
        }

        const workflowData = await workflowRes.json();
        setWorkflow(workflowData);

        if (instancesRes.ok) {
          const instancesData = await instancesRes.json();
          setInstances(instancesData);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchData();
    }
  }, [id]);

  if (loading) {
    return <div className="loading">Loading workflow details...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  if (!workflow) {
    return <div className="error">Workflow not found</div>;
  }

  return (
    <div>
      <div style={{ marginBottom: '1rem' }}>
        <Link to="/workflows" className="btn btn-nav">← Back to Workflows</Link>
      </div>

      <h1>{workflow.name} v{workflow.version}</h1>
      
      <div className="card">
        <h2>Workflow Definition</h2>
        <WorkflowGraph 
          definition={workflow.definition} 
          title={`${workflow.name} v${workflow.version} - Definition`}
        />
      </div>

      <div className="card">
        <h2>Raw Definition JSON</h2>
        <div className="json-viewer">
          {JSON.stringify(workflow.definition, null, 2)}
        </div>
      </div>

      <div className="card">
        <h2>Instances</h2>
        {instances.length === 0 ? (
          <p>No instances found for this workflow</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Status</th>
                <th>Created</th>
                <th>Started</th>
                <th>Completed</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {instances.map((instance) => (
                <tr key={instance.id}>
                  <td>{instance.id}</td>
                  <td>
                    <span className={`status ${instance.status}`}>
                      {instance.status}
                    </span>
                  </td>
                  <td>{new Date(instance.created_at).toLocaleString()}</td>
                  <td>{instance.started_at ? new Date(instance.started_at).toLocaleString() : '-'}</td>
                  <td>{instance.completed_at ? new Date(instance.completed_at).toLocaleString() : '-'}</td>
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
