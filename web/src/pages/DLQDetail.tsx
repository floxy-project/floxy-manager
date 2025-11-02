import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';

interface DeadLetterItem {
  id: number;
  instance_id: number;
  workflow_id: string;
  step_id: number;
  step_name: string;
  step_type: string;
  input: any;
  error: string | null;
  reason: string;
  created_at: string;
}

export const DLQDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [dlqItem, setDlqItem] = useState<DeadLetterItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isRequeuing, setIsRequeuing] = useState(false);
  const [editedInput, setEditedInput] = useState('');
  const [inputError, setInputError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      fetchDLQItem();
    }
  }, [id]);

  const fetchDLQItem = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/dlq/${id}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch DLQ item');
      }
      
      const data: DeadLetterItem = await response.json();
      setDlqItem(data);
      setEditedInput(JSON.stringify(data.input, null, 2));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const validateJSON = (jsonString: string): boolean => {
    try {
      JSON.parse(jsonString);
      return true;
    } catch {
      return false;
    }
  };

  const handleInputChange = (value: string) => {
    setEditedInput(value);
    if (value.trim() === '') {
      setInputError(null);
      return;
    }
    
    if (!validateJSON(value)) {
      setInputError('Invalid JSON format');
    } else {
      setInputError(null);
    }
  };

  const handleRequeue = async () => {
    if (!dlqItem || isRequeuing) return;
    
    if (editedInput.trim() === '') {
      alert('Please provide input data');
      return;
    }
    
    if (!validateJSON(editedInput)) {
      alert('Please fix JSON format errors');
      return;
    }

    setIsRequeuing(true);
    try {
      const newInput = JSON.parse(editedInput);
      const response = await fetch(`/api/dlq/${id}/requeue`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ new_input: newInput }),
      });

      // Check if response is successful (2xx status codes)
      if (response.status >= 200 && response.status < 300) {
        alert('Item successfully requeued!');
        // Redirect to DLQ list
        window.location.href = '/dlq';
      } else {
        // Handle error responses
        let errorMessage = 'Failed to requeue item';
        
        try {
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage = errorData.message;
          }
        } catch (parseError) {
          errorMessage = response.statusText || errorMessage;
        }
        
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error('Error requeuing item:', error);
      alert(`Error requeuing item: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsRequeuing(false);
    }
  };

  if (loading) {
    return <div className="loading">Loading DLQ item...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  if (!dlqItem) {
    return <div className="error">DLQ item not found</div>;
  }

  return (
    <div>
      <div style={{ marginBottom: '1rem' }}>
        <Link to="/dlq" className="btn btn-nav">← Back to DLQ</Link>
      </div>

      <h1>DLQ Item {dlqItem.id}</h1>

      <div className="card">
        <h2>Item Information</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
          <div>
            <strong>ID:</strong> {dlqItem.id}
          </div>
          <div>
            <strong>Instance ID:</strong> 
            <Link to={`/instances/${dlqItem.instance_id}`} style={{ marginLeft: '0.5rem' }}>
              {dlqItem.instance_id}
            </Link>
          </div>
          <div>
            <strong>Workflow ID:</strong> {dlqItem.workflow_id}
          </div>
          <div>
            <strong>Step ID:</strong> {dlqItem.step_id}
          </div>
          <div>
            <strong>Step Name:</strong> {dlqItem.step_name}
          </div>
          <div>
            <strong>Step Type:</strong> 
            <span className="status failed" style={{ marginLeft: '0.5rem' }}>
              {dlqItem.step_type}
            </span>
          </div>
          <div>
            <strong>Reason:</strong> {dlqItem.reason}
          </div>
          <div>
            <strong>Created:</strong> {new Date(dlqItem.created_at).toLocaleString()}
          </div>
        </div>
      </div>

      {dlqItem.error && (
        <div className="card">
          <h2>Error Details</h2>
          <div className="json-viewer" style={{ backgroundColor: '#fef2f2', borderColor: '#fecaca' }}>
            {dlqItem.error}
          </div>
        </div>
      )}

      <div className="card">
        <h2>Requeue Item</h2>
        <p style={{ marginBottom: '1rem', color: '#6b7280' }}>
          Edit the input data below and click "Requeue" to send this item back to the workflow queue.
        </p>
        
        <div className="form-group">
          <label htmlFor="input-data">Input Data (JSON):</label>
          <textarea
            id="input-data"
            value={editedInput}
            onChange={(e) => handleInputChange(e.target.value)}
            rows={15}
            style={{ 
              width: '100%', 
              padding: '0.75rem', 
              border: inputError ? '1px solid #ef4444' : '1px solid #d1d5db', 
              borderRadius: '8px',
              fontFamily: 'JetBrains Mono, Fira Code, Courier New, monospace',
              fontSize: '0.9rem',
              backgroundColor: inputError ? '#fef2f2' : 'white'
            }}
          />
          {inputError && (
            <div style={{ color: '#ef4444', fontSize: '0.875rem', marginTop: '0.5rem' }}>
              {inputError}
            </div>
          )}
        </div>

        <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
          <button
            className="btn btn-secondary"
            onClick={() => setEditedInput(JSON.stringify(dlqItem.input, null, 2))}
            disabled={isRequeuing}
          >
            Reset to Original
          </button>
          <button
            className="btn btn-primary"
            onClick={handleRequeue}
            disabled={isRequeuing || !!inputError}
          >
            {isRequeuing ? 'Requeuing...' : 'Requeue Item'}
          </button>
        </div>
      </div>
    </div>
  );
};
