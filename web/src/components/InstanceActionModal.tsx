import React, { useState } from 'react';
import { Modal } from './Modal';

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

  const actionTitle = actionType === 'cancel' ? 'Cancel Instance' : 'Abort Instance';
  const actionDescription = actionType === 'cancel' 
    ? 'Cancel this workflow instance. The instance will be gracefully stopped.'
    : 'Abort this workflow instance immediately. This action cannot be undone.';

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
        <h3 className="text-xl font-bold text-slate-900 pr-8">{actionTitle}</h3>
        <button 
          className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900 text-xl leading-none"
          onClick={onClose}
        >
          ×
        </button>
      </div>
      
      <div className="px-6 py-6 flex-1 overflow-y-auto">
        <p className="mb-4 text-slate-700">{actionDescription}</p>
        
        <div className="mb-4">
          <label htmlFor="action-reason" className="block mb-2 font-semibold text-slate-700">Reason (optional):</label>
          <textarea
            id="action-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={`Enter a reason for ${actionType}ing this instance...`}
            rows={3}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-y"
          />
        </div>
      </div>
      
      <div className="px-6 py-4 border-t border-slate-200 flex justify-end gap-3">
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
        >
          {isSubmitting ? `${actionTitle}ing...` : actionTitle}
        </button>
      </div>
    </Modal>
  );
};
