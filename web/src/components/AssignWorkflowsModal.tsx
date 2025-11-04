import React, { useState, useEffect } from 'react';
import { Modal } from './Modal';
import apiClient from '../utils/api';

interface WorkflowDefinition {
  id: string;
  name: string;
  version: number;
  definition: any;
  created_at: string;
}

interface WorkflowsResponse {
  items: WorkflowDefinition[];
  page: number;
  page_size: number;
  total: number;
}

interface AssignWorkflowsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAssign: () => void;
  projectId: string;
}

export const AssignWorkflowsModal: React.FC<AssignWorkflowsModalProps> = ({
  isOpen,
  onClose,
  onAssign,
  projectId,
}) => {
  const [unassignedWorkflows, setUnassignedWorkflows] = useState<WorkflowDefinition[]>([]);
  const [selectedWorkflows, setSelectedWorkflows] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [assignAll, setAssignAll] = useState(false);

  useEffect(() => {
    if (isOpen) {
      fetchUnassignedWorkflows();
    } else {
      // Reset state when modal closes
      setUnassignedWorkflows([]);
      setSelectedWorkflows(new Set());
      setError(null);
      setAssignAll(false);
    }
  }, [isOpen]);

  const fetchUnassignedWorkflows = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.listUnassignedWorkflows(1, 100);
      const data: WorkflowsResponse = response.data;
      setUnassignedWorkflows(data.items || []);
    } catch (err) {
      console.error('Error fetching unassigned workflows:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch unassigned workflows');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleWorkflow = (workflowId: string) => {
    const newSelected = new Set(selectedWorkflows);
    if (newSelected.has(workflowId)) {
      newSelected.delete(workflowId);
    } else {
      newSelected.add(workflowId);
    }
    setSelectedWorkflows(newSelected);
    setAssignAll(false);
  };

  const handleSelectAll = () => {
    if (selectedWorkflows.size === unassignedWorkflows.length) {
      setSelectedWorkflows(new Set());
      setAssignAll(false);
    } else {
      setSelectedWorkflows(new Set(unassignedWorkflows.map(w => w.id)));
      setAssignAll(false);
    }
  };

  const handleAssignAll = () => {
    setAssignAll(true);
    setSelectedWorkflows(new Set());
  };

  const handleSubmit = async () => {
    if (submitting) return;

    setSubmitting(true);
    setError(null);

    try {
      const workflowIds = assignAll ? undefined : Array.from(selectedWorkflows);
      const response = await apiClient.assignWorkflowsToProject(
        parseInt(projectId),
        workflowIds
      );

      const message = assignAll
        ? `Successfully assigned ${response.data.assigned_count} workflow(s) to this project.`
        : `Successfully assigned ${response.data.assigned_count} selected workflow(s) to this project.`;

      alert(message);
      onAssign();
      onClose();
    } catch (err) {
      console.error('Error assigning workflows:', err);
      setError(err instanceof Error ? err.message : 'Failed to assign workflows');
    } finally {
      setSubmitting(false);
    }
  };

  const hasSelection = assignAll || selectedWorkflows.size > 0;

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
        <h3 className="text-xl font-bold text-slate-900 pr-8">Assign Workflows to Project</h3>
        <button 
          className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900 text-xl leading-none"
          onClick={onClose}
        >
          ×
        </button>
      </div>
      
      <div className="px-6 py-6 flex-1 overflow-y-auto max-h-[60vh]">
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="w-5 h-5 border-2 border-slate-300 border-t-slate-600 rounded-full animate-spin"></div>
            <span className="ml-2 text-slate-600">Loading unassigned workflows...</span>
          </div>
        ) : unassignedWorkflows.length === 0 ? (
          <div className="text-center py-8 text-slate-500">
            <svg className="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p>No unassigned workflows found.</p>
            <p className="text-sm mt-2">All workflows are already assigned to projects.</p>
          </div>
        ) : (
          <>
            <div className="mb-4 flex items-center justify-between">
              <p className="text-slate-700">
                Select workflows to assign to this project. You can select specific workflows or assign all unassigned workflows.
              </p>
              <div className="flex gap-2">
                <button
                  className="btn btn-outline text-xs py-1.5 px-3"
                  onClick={handleSelectAll}
                >
                  {selectedWorkflows.size === unassignedWorkflows.length ? 'Deselect All' : 'Select All'}
                </button>
                <button
                  className="btn btn-primary text-xs py-1.5 px-3"
                  onClick={handleAssignAll}
                >
                  Assign All ({unassignedWorkflows.length})
                </button>
              </div>
            </div>

            {assignAll && (
              <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg text-blue-700 text-sm">
                <strong>Assign All Mode:</strong> All {unassignedWorkflows.length} unassigned workflow(s) will be assigned to this project.
              </div>
            )}

            <div className="border border-slate-200 rounded-lg overflow-hidden">
              <div className="max-h-[300px] overflow-y-auto">
                <table className="table w-full">
                  <thead className="bg-slate-50 sticky top-0">
                    <tr>
                      <th className="w-12">
                        <input
                          type="checkbox"
                          checked={selectedWorkflows.size === unassignedWorkflows.length && unassignedWorkflows.length > 0 && !assignAll}
                          onChange={handleSelectAll}
                          disabled={assignAll}
                        />
                      </th>
                      <th>Name</th>
                      <th>Version</th>
                      <th>Created</th>
                    </tr>
                  </thead>
                  <tbody>
                    {unassignedWorkflows.map((workflow) => (
                      <tr key={workflow.id}>
                        <td>
                          <input
                            type="checkbox"
                            checked={selectedWorkflows.has(workflow.id)}
                            onChange={() => handleToggleWorkflow(workflow.id)}
                            disabled={assignAll}
                          />
                        </td>
                        <td>
                          <div className="font-medium">{workflow.name}</div>
                          <div className="text-xs text-slate-500 font-mono mt-0.5">
                            {workflow.id}
                          </div>
                        </td>
                        <td>
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-700">
                            v{workflow.version}
                          </span>
                        </td>
                        <td className="text-sm text-slate-600">
                          {new Date(workflow.created_at).toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="mt-4 text-sm text-slate-600">
              {assignAll ? (
                <span>All {unassignedWorkflows.length} workflow(s) will be assigned.</span>
              ) : (
                <span>{selectedWorkflows.size} workflow(s) selected.</span>
              )}
            </div>
          </>
        )}
      </div>
      
      <div className="px-6 py-4 border-t border-slate-200 flex justify-end gap-3">
        <button
          className="btn btn-secondary"
          onClick={onClose}
          disabled={submitting}
        >
          Cancel
        </button>
        <button
          className="btn btn-primary"
          onClick={handleSubmit}
          disabled={submitting || loading || (!hasSelection && !assignAll)}
        >
          {submitting ? 'Assigning...' : assignAll ? `Assign All (${unassignedWorkflows.length})` : `Assign Selected (${selectedWorkflows.size})`}
        </button>
      </div>
    </Modal>
  );
};

