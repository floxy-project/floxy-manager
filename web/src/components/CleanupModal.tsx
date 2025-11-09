import React, { useState } from 'react';
import { X } from 'lucide-react';
import { Modal } from './Modal';
import { authFetch } from '../utils/api';

interface CleanupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCleanup: () => void;
  projectId: string; // used to include X-Project-ID for RBAC middleware
}

export const CleanupModal: React.FC<CleanupModalProps> = ({
  isOpen,
  onClose,
  onCleanup,
  projectId,
}) => {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const response = await authFetch('/api/cleanup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Project-ID': String(projectId),
        },
        body: JSON.stringify({}),
      });

      // Check if response is successful (2xx status codes)
      if (response.status >= 200 && response.status < 300) {
        await response.json(); // Response is empty JSON "{}"
        onCleanup();
        alert('Cleanup completed successfully! Old partitions have been cleaned.');
        onClose();
      } else {
        // Handle error responses (404, 409, 422, 500, etc.)
        let errorMessage = 'Failed to cleanup partitions';
        
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
      console.error('Error cleaning up partitions:', error);
      alert(`Error cleaning up partitions: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
        <h3 className="text-xl font-bold text-slate-900 pr-8">Clean Old Partitions</h3>
        <button 
          className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900"
          onClick={onClose}
          aria-label="Close"
        >
          <X className="w-5 h-5" />
        </button>
      </div>
      
      <div className="px-6 py-6 flex-1 overflow-y-auto">
        <p className="mb-4 text-slate-700">This will permanently delete old partitions based on the configured retention policy. This action cannot be undone.</p>
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
          disabled={isSubmitting}
        >
          {isSubmitting ? 'Cleaning up...' : 'Clean Old Partitions'}
        </button>
      </div>
    </Modal>
  );
};
