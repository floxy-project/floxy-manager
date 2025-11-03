import React, { useState } from 'react';
import { Modal } from './Modal';
import { authFetch } from '../utils/api';

interface CleanupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCleanup: (daysToKeep: number) => void;
  projectId: string; // used to include X-Project-ID for RBAC middleware
}

export const CleanupModal: React.FC<CleanupModalProps> = ({
  isOpen,
  onClose,
  onCleanup,
  projectId,
}) => {
  const [daysToKeep, setDaysToKeep] = useState(30);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (isSubmitting || daysToKeep < 1) return;
    
    setIsSubmitting(true);
    try {
      const response = await authFetch('/api/cleanup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Project-ID': String(projectId),
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

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
        <h3 className="text-xl font-bold text-slate-900 pr-8">Cleanup Old Workflows</h3>
        <button 
          className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900 text-xl leading-none"
          onClick={onClose}
        >
          ×
        </button>
      </div>
      
      <div className="px-6 py-6 flex-1 overflow-y-auto">
        <p className="mb-4 text-slate-700">This will permanently delete workflow instances older than the specified number of days. This action cannot be undone.</p>
        
        <div className="mb-4">
          <label htmlFor="days-to-keep" className="block mb-2 font-semibold text-slate-700">Days to keep:</label>
          <input
            id="days-to-keep"
            type="number"
            min="1"
            value={daysToKeep}
            onChange={(e) => setDaysToKeep(Math.max(1, parseInt(e.target.value) || 1))}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          />
          <small className="text-slate-500 text-sm mt-1 block">
            Workflows older than this number of days will be deleted. Minimum: 1 day.
          </small>
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
          onClick={handleSubmit}
          disabled={isSubmitting || daysToKeep < 1}
        >
          {isSubmitting ? 'Cleaning up...' : 'Cleanup Workflows'}
        </button>
      </div>
    </Modal>
  );
};
