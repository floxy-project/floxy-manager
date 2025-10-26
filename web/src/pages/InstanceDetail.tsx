import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { WorkflowGraph } from '../components/WorkflowGraph';
import { DecisionModal } from '../components/DecisionModal';
import { InstanceActionModal } from '../components/InstanceActionModal';

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

interface WorkflowStep {
  id: number;
  instance_id: number;
  step_name: string;
  step_type: string;
  status: string;
  input: any;
  output: any;
  error: string | null;
  retry_count: number;
  max_retries: number;
  compensation_retry_count: number;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
}

interface WorkflowEvent {
  id: number;
  instance_id: number;
  step_id: number | null;
  event_type: string;
  payload: any;
  created_at: string;
}

interface WorkflowDefinition {
  id: string;
  name: string;
  version: number;
  definition: any;
  created_at: string;
}

export const InstanceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [instance, setInstance] = useState<WorkflowInstance | null>(null);
  const [steps, setSteps] = useState<WorkflowStep[]>([]);
  const [events, setEvents] = useState<WorkflowEvent[]>([]);
  const [workflowDefinition, setWorkflowDefinition] = useState<WorkflowDefinition | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'steps' | 'events' | 'graph'>('steps');
  const [showDecisionModal, setShowDecisionModal] = useState(false);
  const [showActionModal, setShowActionModal] = useState(false);
  const [actionType, setActionType] = useState<'cancel' | 'abort'>('cancel');

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [instanceRes, stepsRes, eventsRes] = await Promise.all([
          fetch(`/api/instances/${id}`),
          fetch(`/api/instances/${id}/steps`),
          fetch(`/api/instances/${id}/events`)
        ]);

        if (!instanceRes.ok) {
          throw new Error('Failed to fetch instance');
        }

        const instanceData = await instanceRes.json();
        setInstance(instanceData);

        if (stepsRes.ok) {
          const stepsData = await stepsRes.json();
          setSteps(stepsData);
        }

        if (eventsRes.ok) {
          const eventsData = await eventsRes.json();
          setEvents(eventsData);
        }

        // Fetch workflow definition
        const workflowRes = await fetch(`/api/workflows/${instanceData.workflow_id}`);
        if (workflowRes.ok) {
          const workflowData = await workflowRes.json();
          setWorkflowDefinition(workflowData);
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

  // Function to determine if decision buttons are needed
  const needsDecision = () => {
    if (!instance || !steps.length) return false;
    
    // Look for the current step (the one that's currently running or waiting)
    // that is in waiting_decision status with type "human"
    const currentStep = steps.find(step => 
      instance.status === 'running' && 
      step.step_type === 'human' && 
      step.status === 'waiting_decision'
    );
    
    return !!currentStep;
  };

  // Function to determine if instance is active (can be cancelled or aborted)
  const isActiveInstance = () => {
    if (!instance) return false;
    
    // Active statuses that can be cancelled or aborted
    const activeStatuses = ['pending', 'running', 'rolling_back', 'cancelling'];
    return activeStatuses.includes(instance.status);
  };

  // Functions for handling decisions (called after successful API call)
  const handleDecisionConfirm = (message: string) => {
    // Refresh data after successful decision
    window.location.reload();
  };

  const handleDecisionReject = (message: string) => {
    // Refresh data after successful decision
    window.location.reload();
  };

  // Functions for handling instance actions
  const handleInstanceAction = (reason: string) => {
    // Refresh data after successful action
    window.location.reload();
  };

  const openActionModal = (type: 'cancel' | 'abort') => {
    setActionType(type);
    setShowActionModal(true);
  };

  if (loading) {
    return <div className="loading">Loading instance details...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  if (!instance) {
    return <div className="error">Instance not found</div>;
  }

  return (
    <div>
      <div style={{ marginBottom: '1rem' }}>
        <Link to="/instances" className="btn btn-nav">← Back to Instances</Link>
      </div>

      <h1>Instance {instance.id}</h1>
      
      <div className="card">
        <h2>Instance Information</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
          <div>
            <strong>Workflow ID:</strong> {instance.workflow_id}
          </div>
          <div>
            <strong>Status:</strong> 
            <span className={`status ${instance.status}`} style={{ marginLeft: '0.5rem' }}>
              {instance.status}
            </span>
          </div>
          <div>
            <strong>Created:</strong> {new Date(instance.created_at).toLocaleString()}
          </div>
          <div>
            <strong>Started:</strong> {instance.started_at ? new Date(instance.started_at).toLocaleString() : '-'}
          </div>
          <div>
            <strong>Completed:</strong> {instance.completed_at ? new Date(instance.completed_at).toLocaleString() : '-'}
          </div>
          {instance.error && (
            <div style={{ gridColumn: '1 / -1' }}>
              <strong>Error:</strong> 
              <div className="json-viewer" style={{ marginTop: '0.5rem' }}>
                {instance.error}
              </div>
            </div>
          )}
        </div>
        
        {/* Decision buttons */}
        {needsDecision() && (
          <div className="decision-buttons">
            <div style={{ flex: 1 }}>
              <strong>Decision Required:</strong>
              <p style={{ margin: '0.5rem 0', color: '#6b7280', fontSize: '0.9rem' }}>
                The instance is waiting for your decision to continue execution.
              </p>
            </div>
            <button
              className="btn btn-primary"
              onClick={() => setShowDecisionModal(true)}
            >
              Make Decision
            </button>
          </div>
        )}

        {/* Instance action buttons */}
        {isActiveInstance() && (
          <div className="decision-buttons">
            <div style={{ flex: 1 }}>
              <strong>Instance Actions:</strong>
              <p style={{ margin: '0.5rem 0', color: '#6b7280', fontSize: '0.9rem' }}>
                This instance is active and can be cancelled or aborted.
              </p>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                className="btn btn-warning"
                onClick={() => openActionModal('cancel')}
              >
                Cancel Instance
              </button>
              <button
                className="btn btn-danger"
                onClick={() => openActionModal('abort')}
              >
                Abort Instance
              </button>
            </div>
          </div>
        )}
      </div>

      <div className="card">
        <h2>Input Data</h2>
        <div className="json-viewer">
          {JSON.stringify(instance.input, null, 2)}
        </div>
      </div>

      {instance.output && (
        <div className="card">
          <h2>Output Data</h2>
          <div className="json-viewer">
            {JSON.stringify(instance.output, null, 2)}
          </div>
        </div>
      )}

      <div className="card">
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
          <button 
            className={`btn ${activeTab === 'steps' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('steps')}
          >
            Steps
          </button>
          <button 
            className={`btn ${activeTab === 'events' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('events')}
          >
            Events
          </button>
          <button 
            className={`btn ${activeTab === 'graph' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('graph')}
          >
            Graph
          </button>
        </div>

        {activeTab === 'steps' && (
          <div>
            <h3>Workflow Steps</h3>
            {steps.length === 0 ? (
              <p>No steps found</p>
            ) : (
              <ul className="step-list">
                {steps.map((step) => (
                  <li key={step.id} className="step-item">
                    <div className="step-header">
                      <span className="step-name">{step.step_name}</span>
                      <div>
                        <span className="step-type">{step.step_type}</span>
                        <span className={`status ${step.status}`} style={{ marginLeft: '0.5rem' }}>
                          {step.status}
                        </span>
                      </div>
                    </div>
                    <div className="step-details">
                      <div>
                        <strong>Retries:</strong> {step.retry_count}/{step.max_retries}
                      </div>
                      <div>
                        <strong>Started:</strong> {step.started_at ? new Date(step.started_at).toLocaleString() : '-'}
                      </div>
                      <div>
                        <strong>Completed:</strong> {step.completed_at ? new Date(step.completed_at).toLocaleString() : '-'}
                      </div>
                      {step.compensation_retry_count > 0 && (
                        <div>
                          <strong>Compensation Retries:</strong> {step.compensation_retry_count}
                        </div>
                      )}
                    </div>
                    {step.error && (
                      <div style={{ marginTop: '0.5rem' }}>
                        <strong>Error:</strong>
                        <div className="json-viewer" style={{ marginTop: '0.25rem' }}>
                          {step.error}
                        </div>
                      </div>
                    )}
                    {step.input && (
                      <div style={{ marginTop: '0.5rem' }}>
                        <strong>Input:</strong>
                        <div className="json-viewer" style={{ marginTop: '0.25rem' }}>
                          {JSON.stringify(step.input, null, 2)}
                        </div>
                      </div>
                    )}
                    {step.output && (
                      <div style={{ marginTop: '0.5rem' }}>
                        <strong>Output:</strong>
                        <div className="json-viewer" style={{ marginTop: '0.25rem' }}>
                          {JSON.stringify(step.output, null, 2)}
                        </div>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {activeTab === 'events' && (
          <div>
            <h3>Workflow Events</h3>
            {events.length === 0 ? (
              <p>No events found</p>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>Type</th>
                    <th>Step ID</th>
                    <th>Payload</th>
                  </tr>
                </thead>
                <tbody>
                  {events.map((event) => (
                    <tr key={event.id}>
                      <td>{new Date(event.created_at).toLocaleString()}</td>
                      <td>{event.event_type}</td>
                      <td>{event.step_id || '-'}</td>
                      <td>
                        {event.payload && (
                          <div className="json-viewer" style={{ maxHeight: '100px' }}>
                            {JSON.stringify(event.payload, null, 2)}
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'graph' && (
          <div>
            <h3>Workflow Execution Graph</h3>
            {workflowDefinition ? (
              <WorkflowGraph 
                definition={workflowDefinition.definition} 
                steps={steps}
                title={`Instance ${instance?.id} - ${workflowDefinition.name} v${workflowDefinition.version}`}
              />
            ) : (
              <p>Loading workflow definition...</p>
            )}
          </div>
        )}
      </div>

      {/* Decision modal */}
      <DecisionModal
        isOpen={showDecisionModal}
        onClose={() => setShowDecisionModal(false)}
        onConfirm={handleDecisionConfirm}
        onReject={handleDecisionReject}
        instanceId={id || ''}
      />

      {/* Instance action modal */}
      <InstanceActionModal
        isOpen={showActionModal}
        onClose={() => setShowActionModal(false)}
        onAction={handleInstanceAction}
        instanceId={id || ''}
        actionType={actionType}
      />
    </div>
  );
};
