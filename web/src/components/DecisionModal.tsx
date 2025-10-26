import React, { useState } from 'react';

interface DecisionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (message: string) => void;
  onReject: (message: string) => void;
  instanceId: string;
}

export const DecisionModal: React.FC<DecisionModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onReject,
  instanceId
}) => {
  const [message, setMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (action: 'confirm' | 'reject') => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const response = await fetch(`/api/instances/${instanceId}/make-decision/${action}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message }),
      });

      // Check if response is successful (2xx status codes)
      if (response.status >= 200 && response.status < 300) {
        // Call the parent handlers to refresh the page
        if (action === 'confirm') {
          onConfirm(message);
        } else {
          onReject(message);
        }
        
        setMessage('');
        onClose();
      } else {
        // Handle error responses (404, 409, 422, 500, etc.)
        let errorMessage = `Failed to ${action} decision`;
        
        try {
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage = errorData.message;
          }
        } catch (parseError) {
          // If we can't parse the error response, use the status text
          errorMessage = response.statusText || errorMessage;
        }
        
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error(`Error ${action}ing decision:`, error);
      alert(`Error ${action === 'confirm' ? 'confirming' : 'rejecting'} decision: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Make Decision</h3>
          <button className="btn btn-close" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          <p>The instance is waiting for your decision. Please choose an action and add a comment (optional):</p>
          
          <div className="form-group">
            <label htmlFor="decision-message">Message:</label>
            <textarea
              id="decision-message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Enter a comment for your decision..."
              rows={3}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            />
          </div>
        </div>
        
        <div className="modal-footer">
          <button
            className="btn btn-secondary"
            onClick={onClose}
            disabled={isSubmitting}
          >
            Cancel
          </button>
          <button
            className="btn btn-danger"
            onClick={() => handleSubmit('reject')}
            disabled={isSubmitting}
            style={{ marginLeft: '0.5rem' }}
          >
            {isSubmitting ? 'Rejecting...' : 'Reject'}
          </button>
          <button
            className="btn btn-success"
            onClick={() => handleSubmit('confirm')}
            disabled={isSubmitting}
            style={{ marginLeft: '0.5rem' }}
          >
            {isSubmitting ? 'Confirming...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  );
};
