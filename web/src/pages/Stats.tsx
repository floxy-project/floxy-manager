import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { authFetch } from '../utils/api';

interface WorkflowStats {
    name: string;
    version: number;
    total_instances: number;
    completed_instances: number;
    failed_instances: number;
    running_instances: number;
    average_duration: number;
}

interface StatsResponse {
    items: WorkflowStats[];
    page: number;
    page_size: number;
    total: number;
}

export const Stats: React.FC = () => {
    const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
    const [stats, setStats] = useState<WorkflowStats[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize] = useState(20);
    const [totalItems, setTotalItems] = useState(0);

    useEffect(() => {
        if (tenantId && projectId) {
            fetchStats();
        }
    }, [tenantId, projectId, currentPage]);

    const fetchStats = async () => {
        try {
            setLoading(true);
            const response = await authFetch(`/api/v1/stats?tenant_id=${tenantId}&project_id=${projectId}&page=${currentPage}&page_size=${pageSize}`);
            if (!response.ok) {
                throw new Error('Failed to fetch stats');
            }
            const data: StatsResponse = await response.json();
            setStats(data.items || data);
            setTotalItems(data.total || data.items?.length || 0);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        } finally {
            setLoading(false);
        }
    };

    const totalPages = Math.ceil(totalItems / pageSize);

    const formatDuration = (nanoseconds: number) => {
        if (nanoseconds === 0 || nanoseconds === null) return '-';
        
        const seconds = nanoseconds / 1000000000;
        
        if (seconds < 60) return `${seconds.toFixed(1)}s`;
        if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
        return `${(seconds / 3600).toFixed(1)}h`;
    };

    const getSuccessRate = (completed: number, total: number) => {
        if (total === 0) return '0.0';
        return ((completed / total) * 100).toFixed(1);
    };

    if (loading) {
        return (
            <div className="loading">
                <div className="flex items-center justify-center gap-2">
                    <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
                    <span>Loading statistics...</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="error">
                <div className="flex items-center justify-center gap-2">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span>Error: {error}</span>
                </div>
            </div>
        );
    }

    const totalInstances = stats.reduce((sum, stat) => sum + stat.total_instances, 0);
    const totalCompleted = stats.reduce((sum, stat) => sum + stat.completed_instances, 0);
    const totalFailed = stats.reduce((sum, stat) => sum + stat.failed_instances, 0);
    const totalRunning = stats.reduce((sum, stat) => sum + stat.running_instances, 0);

    return (
        <div>
            <h1>Workflow Statistics</h1>

            <div className="stats-grid mb-6">
                <div className="stat-card">
                    <div className="stat-number">{totalInstances}</div>
                    <div className="stat-label">Total Instances</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number text-emerald-600 dark:text-emerald-400">{totalCompleted}</div>
                    <div className="stat-label">Completed</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number text-red-600 dark:text-red-400">{totalFailed}</div>
                    <div className="stat-label">Failed</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number text-amber-600 dark:text-amber-400">{totalRunning}</div>
                    <div className="stat-label">Running</div>
                </div>
                <div className="stat-card">
                    <div className="stat-number">{stats.length}</div>
                    <div className="stat-label">Workflow Types</div>
                </div>
            </div>

            <div className="card">
                {stats.length === 0 ? (
                    <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
                        <svg className="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                        </svg>
                        <p>No statistics available</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="table">
                            <thead>
                            <tr>
                                <th>Workflow</th>
                                <th>Version</th>
                                <th>Total Instances</th>
                                <th>Completed</th>
                                <th>Failed</th>
                                <th>Running</th>
                                <th>Success Rate</th>
                                <th>Avg Duration</th>
                            </tr>
                            </thead>
                            <tbody>
                            {stats.map((stat) => {
                                const successRate = parseFloat(getSuccessRate(stat.completed_instances, stat.total_instances));
                                return (
                                    <tr key={`${stat.name}-${stat.version}`}>
                                        <td className="font-medium">{stat.name}</td>
                                        <td>
                                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 dark:bg-[#3e3e42] text-slate-700 dark:text-[#ff4500]400">
                                                v{stat.version}
                                            </span>
                                        </td>
                                        <td>{stat.total_instances}</td>
                                        <td className="text-emerald-600 dark:text-emerald-400 font-medium">{stat.completed_instances}</td>
                                        <td className="text-red-600 dark:text-red-400 font-medium">{stat.failed_instances}</td>
                                        <td className="text-amber-600 dark:text-amber-400 font-medium">{stat.running_instances}</td>
                                        <td>
                                            <div className="flex items-center gap-2">
                                                <span className={`font-medium ${successRate >= 90 ? 'text-emerald-600 dark:text-emerald-400' : successRate >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400'}`}>
                                                    {successRate}%
                                                </span>
                                                <div className="progress-bar w-16">
                                                    <div 
                                                        className={`progress-fill ${successRate >= 90 ? 'success' : successRate >= 70 ? 'primary' : 'danger'}`}
                                                        style={{ width: `${successRate}%` }}
                                                    />
                                                </div>
                                            </div>
                                        </td>
                                        <td className="font-mono text-sm">{formatDuration(stat.average_duration)}</td>
                                    </tr>
                                );
                            })}
                            </tbody>
                        </table>
                    </div>
                )}

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
            </div>
        </div>
    );
};
