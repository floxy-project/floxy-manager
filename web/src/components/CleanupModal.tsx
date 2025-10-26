import React, { useState } from 'react';

interface CleanupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCleanup: (daysToKeep: number) => void;
}

export const CleanupModal: React.FC<CleanupModalProps> = ({
  isOpen,
  onClose,
  onCleanup
}) => {
  const [daysToKeep, setDaysToKeep] = useState(30);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (isSubmitting || daysToKeep < 1) return;
    
    setIsSubmitting(true);
    try {
      const response = await fetch('/api/cleanup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ days_to_keep: daysToKeep }),
      });

      // Check if response is successful (2xx status codes)
      if (response.status >= 200 && response.status < 300) {
        const result = await response.json();
        onCleanup(daysToKeep);
        alert(`Cleanup completed successfully! Deleted ${result.deleted_count} old workflows (kept workflows newer than ${result.days_to_keep} days).`);
        setDaysToKeep(30);
        onClose();
      } else {
        // Handle error responses (404, 409, 422, 500, etc.)
        let errorMessage = 'Failed to cleanup workflows';
        
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
      console.error('Error cleaning up workflows:', error);
      alert(`Error cleaning up workflows: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Cleanup Old Workflows</h3>
          <button className="btn btn-close" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          <p>This will permanently delete workflow instances older than the specified number of days. This action cannot be undone.</p>
          
          <div className="form-group">
            <label htmlFor="days-to-keep">Days to keep:</label>
            <input
              id="days-to-keep"
              type="number"
              min="1"
              value={daysToKeep}
              onChange={(e) => setDaysToKeep(Math.max(1, parseInt(e.target.value) || 1))}
              style={{ 
                width: '100%', 
                padding: '0.75rem', 
                border: '1px solid #d1d5db', 
                borderRadius: '8px',
                fontSize: '1rem'
              }}
            />
            <small style={{ color: '#6b7280', fontSize: '0.875rem' }}>
              Workflows older than this number of days will be deleted. Minimum: 1 day.
            </small>
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
            onClick={handleSubmit}
            disabled={isSubmitting || daysToKeep < 1}
            style={{ marginLeft: '0.5rem' }}
          >
            {isSubmitting ? 'Cleaning up...' : 'Cleanup Workflows'}
          </button>
        </div>
      </div>
    </div>
  );
};
