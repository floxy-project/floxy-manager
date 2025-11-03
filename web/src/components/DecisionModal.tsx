import React, { useState } from 'react';
import { Modal } from './Modal';

import { authFetch } from '../utils/api';

interface DecisionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (message: string) => void;
  onReject: (message: string) => void;
  instanceId: string;
  projectId: string; // explicit projectId to pass RBAC middleware
}

export const DecisionModal: React.FC<DecisionModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onReject,
  instanceId,
  projectId,
}) => {
  const [message, setMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (action: 'confirm' | 'reject') => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const response = await authFetch(`/api/instances/${instanceId}/make-decision/${action}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Project-ID': String(projectId),
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

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
        <h3 className="text-xl font-bold text-slate-900 pr-8">Make Decision</h3>
        <button 
          className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900 text-xl leading-none"
          onClick={onClose}
        >
          ×
        </button>
      </div>
      
      <div className="px-6 py-6 flex-1 overflow-y-auto">
        <p className="mb-4 text-slate-700">The instance is waiting for your decision. Please choose an action and add a comment (optional):</p>
        
        <div className="mb-4">
          <label htmlFor="decision-message" className="block mb-2 font-semibold text-slate-700">Message:</label>
          <textarea
            id="decision-message"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder="Enter a comment for your decision..."
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
          className="btn btn-danger"
          onClick={() => handleSubmit('reject')}
          disabled={isSubmitting}
        >
          {isSubmitting ? 'Rejecting...' : 'Reject'}
        </button>
        <button
          className="btn btn-success"
          onClick={() => handleSubmit('confirm')}
          disabled={isSubmitting}
        >
          {isSubmitting ? 'Confirming...' : 'Confirm'}
        </button>
      </div>
    </Modal>
  );
};
