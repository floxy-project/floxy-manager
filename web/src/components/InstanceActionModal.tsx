import React, { useState } from 'react';

interface InstanceActionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAction: (reason: string) => void;
  instanceId: string;
  actionType: 'cancel' | 'abort';
}

export const InstanceActionModal: React.FC<InstanceActionModalProps> = ({
  isOpen,
  onClose,
  onAction,
  instanceId,
  actionType
}) => {
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const response = await fetch(`/api/instances/${instanceId}/${actionType}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ reason }),
      });

      // Check if response is successful (2xx status codes)
      if (response.status >= 200 && response.status < 300) {
        onAction(reason);
        setReason('');
        onClose();
      } else {
        // Handle error responses (404, 409, 422, 500, etc.)
        let errorMessage = `Failed to ${actionType} instance`;
        
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
      console.error(`Error ${actionType}ing instance:`, error);
      alert(`Error ${actionType}ing instance: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  const actionTitle = actionType === 'cancel' ? 'Cancel Instance' : 'Abort Instance';
  const actionDescription = actionType === 'cancel' 
    ? 'Cancel this workflow instance. The instance will be gracefully stopped.'
    : 'Abort this workflow instance immediately. This action cannot be undone.';

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{actionTitle}</h3>
          <button className="btn btn-close" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          <p>{actionDescription}</p>
          
          <div className="form-group">
            <label htmlFor="action-reason">Reason (optional):</label>
            <textarea
              id="action-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder={`Enter a reason for ${actionType}ing this instance...`}
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
            className={`btn ${actionType === 'cancel' ? 'btn-warning' : 'btn-danger'}`}
            onClick={handleSubmit}
            disabled={isSubmitting}
            style={{ marginLeft: '0.5rem' }}
          >
            {isSubmitting ? `${actionTitle}ing...` : actionTitle}
          </button>
        </div>
      </div>
    </div>
  );
};
