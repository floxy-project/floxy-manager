import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { authFetch } from '../utils/api';
import { WorkflowGraph } from '../components/WorkflowGraph';
import { JsonViewer } from '../components/JsonViewer';

interface WorkflowDefinition {
    id: string;
    name: string;
    version: number;
    definition: any;
    created_at: string;
}

interface WorkflowInstance {
    id: number;
    workflow_id: string;
    status: string;
    input: any;
    output: any;
    error: string | null;
    started_at: string | null;
    completed_at: string | null;
    created_at: string;
    updated_at: string;
}

interface InstancesResponse {
    items: WorkflowInstance[];
    page: number;
    page_size: number;
    total: number;
}

export const WorkflowDetail: React.FC = () => {
    const { tenantId, projectId, id } = useParams<{ tenantId: string; projectId: string; id: string }>();
    const [workflow, setWorkflow] = useState<WorkflowDefinition | null>(null);
    const [instances, setInstances] = useState<WorkflowInstance[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize] = useState(20);
    const [totalItems, setTotalItems] = useState(0);

    useEffect(() => {
        if (tenantId && projectId && id) {
            fetchData();
        }
    }, [tenantId, projectId, id, currentPage]);

    const fetchData = async () => {
        try {
            setLoading(true);
            const [workflowRes, instancesRes] = await Promise.all([
                authFetch(`/api/v1/workflows/${id}?tenant_id=${tenantId}&project_id=${projectId}`),
                authFetch(`/api/v1/workflows/${id}/instances?tenant_id=${tenantId}&project_id=${projectId}&page=${currentPage}&page_size=${pageSize}`)
            ]);

            if (!workflowRes.ok) {
                throw new Error('Failed to fetch workflow');
            }

            const workflowData = await workflowRes.json();
            setWorkflow(workflowData);

            if (instancesRes.ok) {
                const instancesData: InstancesResponse = await instancesRes.json();
                setInstances(instancesData.items || []);
                setTotalItems(instancesData.total || 0);
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        } finally {
            setLoading(false);
        }
    };

    const totalPages = Math.ceil(totalItems / pageSize);

    if (loading) {
        return <div className="loading">Loading workflow details...</div>;
    }

    if (error) {
        return <div className="error">Error: {error}</div>;
    }

    if (!workflow) {
        return <div className="error">Workflow not found</div>;
    }

    return (
        <div>
            <div style={{ marginBottom: '1rem' }}>
                <Link to={`/tenants/${tenantId}/projects/${projectId}/workflows`} className="btn btn-nav">← Back to Workflows</Link>
            </div>

            <h1>{workflow.name} v{workflow.version}</h1>

            <div className="card">
                <h2>Workflow Definition</h2>
                <WorkflowGraph
                    definition={workflow.definition}
                    title={`${workflow.name} v${workflow.version} - Definition`}
                    showLegend={false}
                />
            </div>

            <div className="card">
                <h2>Raw Definition JSON</h2>
                <JsonViewer data={workflow.definition} />
            </div>

            <div className="card">
                <h2>Instances</h2>
                {instances.length === 0 ? (
                    <p>No instances found for this workflow</p>
                ) : (
                    <>
                        <div className="overflow-x-auto">
                            <table className="table">
                                <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Status</th>
                                    <th>Created</th>
                                    <th>Started</th>
                                    <th>Completed</th>
                                    <th>Actions</th>
                                </tr>
                                </thead>
                                <tbody>
                                {instances.map((instance) => (
                                    <tr key={instance.id}>
                                        <td>{instance.id}</td>
                                        <td>
                            <span className={`status ${instance.status}`}>
                              {instance.status}
                            </span>
                                        </td>
                                        <td>{new Date(instance.created_at).toLocaleString()}</td>
                                        <td>{instance.started_at ? new Date(instance.started_at).toLocaleString() : '-'}</td>
                                        <td>{instance.completed_at ? new Date(instance.completed_at).toLocaleString() : '-'}</td>
                                        <td>
                                            <Link 
                                                to={`/tenants/${tenantId}/projects/${projectId}/instances/${instance.id}`} 
                                                className="btn btn-primary"
                                            >
                                                View Details
                                            </Link>
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
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                                    </svg>
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
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                                    </svg>
                                </button>
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    );
};
