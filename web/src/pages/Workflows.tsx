import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { AlertCircle, Edit, Plus, Clipboard, ChevronLeft, ChevronRight } from 'lucide-react';
import { authFetch } from '../utils/api';
import apiClient from '../utils/api';
import { useRBAC } from '../auth/permissions';
import { AssignWorkflowsModal } from '../components/AssignWorkflowsModal';
import { WorkflowBuilder, type WorkflowDefinition as BuilderWorkflowDefinition } from '../components/WorkflowBuilder';

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

export const Workflows: React.FC = () => {
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const [workflows, setWorkflows] = useState<WorkflowDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [totalItems, setTotalItems] = useState(0);
  const rbac = useRBAC(projectId);
  const [showAssignWorkflowsModal, setShowAssignWorkflowsModal] = useState(false);
  const [showWorkflowBuilder, setShowWorkflowBuilder] = useState(false);

  useEffect(() => {
    if (tenantId && projectId) {
      fetchWorkflows();
    }
  }, [tenantId, projectId, currentPage]);

  const fetchWorkflows = async () => {
    try {
      setLoading(true);
      const response = await authFetch(`/api/v1/workflows?tenant_id=${tenantId}&project_id=${projectId}&page=${currentPage}&page_size=${pageSize}`);
      if (!response.ok) {
        throw new Error('Failed to fetch workflows');
      }
      const data: WorkflowsResponse = await response.json();
      setWorkflows(data.items);
      setTotalItems(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const handleAssignWorkflows = () => {
    fetchWorkflows(); // Refresh workflows list after assignment
  };

  const totalPages = Math.ceil(totalItems / pageSize);

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading workflows...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error">
        <div className="flex items-center justify-center gap-2">
          <AlertCircle className="w-5 h-5" />
          <span>Error: {error}</span>
        </div>
      </div>
    );
  }

  // Check if user can view workflows in this project
  if (!rbac.canViewProject()) {
    return (
      <div>
        <h1>Workflow Definitions</h1>
        <div className="card">
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <p>You don't have permission to view workflows in this project.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1>Workflow Definitions</h1>
        {rbac.canManageProject() && (
          <div className="flex gap-3">
            <button
              className="btn btn-primary"
              onClick={() => setShowWorkflowBuilder(true)}
            >
              <Edit className="w-4 h-4" />
              Create Workflow
            </button>
            <button
              className="btn btn-outline"
              onClick={() => setShowAssignWorkflowsModal(true)}
            >
              <Plus className="w-4 h-4" />
              Assign Workflows
            </button>
          </div>
        )}
      </div>
      
      <div className="card">
        {workflows.length === 0 ? (
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <Clipboard className="w-12 h-12 mx-auto mb-3 opacity-50" />
            <p>No workflow definitions found</p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Version</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {workflows.map((workflow) => (
                    <tr key={workflow.id}>
                      <td>
                        <div className="font-medium">{workflow.name}</div>
                        <div className="text-xs text-slate-500 dark:text-[#ff4500]500 font-mono mt-0.5">
                          {workflow.id}
                        </div>
                      </td>
                      <td>
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 dark:bg-[#3e3e42] text-slate-700 dark:text-[#ff4500]400">
                          v{workflow.version}
                        </span>
                      </td>
                      <td className="text-sm text-slate-600 dark:text-[#ff4500]500">
                        {new Date(workflow.created_at).toLocaleString()}
                      </td>
                      <td>
                        {rbac.canViewProject() ? (
                          <Link 
                            to={`/tenants/${tenantId}/projects/${projectId}/workflows/${workflow.id}`} 
                            className="btn btn-primary text-xs py-1.5 px-3"
                          >
                            View Details
                          </Link>
                        ) : (
                          <span className="text-xs text-slate-400">No access</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4 mt-6 pt-6 border-t border-slate-200 dark:border-slate-700">
                <button
                  className="btn btn-outline"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </button>
                <div className="px-4 py-2 rounded-lg" style={{
                  background: 'rgba(255, 255, 255, 0.5)',
                  backdropFilter: 'blur(12px) saturate(150%)',
                  WebkitBackdropFilter: 'blur(12px) saturate(150%)',
                  border: '1px solid rgba(255, 255, 255, 0.4)',
                  boxShadow: '0 4px 16px rgba(0, 0, 0, 0.06)'
                }}>
                  <span className="text-sm font-medium text-slate-700 dark:text-[#ff4500]400">
                    Page {currentPage} of {totalPages}
                  </span>
                </div>
                <button
                  className="btn btn-outline"
                  onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                  disabled={currentPage === totalPages}
                >
                  Next
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      <AssignWorkflowsModal
        isOpen={showAssignWorkflowsModal}
        onClose={() => setShowAssignWorkflowsModal(false)}
        onAssign={handleAssignWorkflows}
        projectId={projectId || ''}
      />
      <WorkflowBuilder
        isOpen={showWorkflowBuilder}
        onClose={() => setShowWorkflowBuilder(false)}
        onSave={async (definition) => {
          if (!tenantId || !projectId) {
            alert('Tenant ID and Project ID are required');
            return;
          }

          try {
            // Convert definition to JSON format
            const jsonDefinition = {
              start: definition.definition.start,
              steps: definition.definition.steps,
              dlq_enabled: definition.definition.dlq_enabled || false,
            };

            await apiClient.createWorkflow(
              parseInt(tenantId),
              parseInt(projectId),
              definition.name,
              definition.version,
              jsonDefinition
            );

            // Refresh workflows list
            await fetchWorkflows();
            setShowWorkflowBuilder(false);
          } catch (err) {
            console.error('Failed to save workflow:', err);
            alert(`Failed to save workflow: ${err instanceof Error ? err.message : 'Unknown error'}`);
          }
        }}
      />
    </div>
  );
};
