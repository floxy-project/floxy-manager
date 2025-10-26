import React, { useState, useEffect } from 'react';

interface WorkflowStats {
    workflow_name: string;
    version: number;
    total_instances: number;
    completed_instances: number;
    failed_instances: number;
    running_instances: number;
    average_duration: number;
}

export const Stats: React.FC = () => {
    const [stats, setStats] = useState<WorkflowStats[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const response = await fetch('/api/stats');
                if (!response.ok) {
                    throw new Error('Failed to fetch stats');
                }
                const data = await response.json();
                setStats(data);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    if (loading) {
        return <div className="loading">Loading statistics...</div>;
    }

    if (error) {
        return <div className="error">Error: {error}</div>;
    }

    const formatDuration = (nanoseconds: number) => {
        if (nanoseconds === 0 || nanoseconds === null) return '-';
        
        // Convert nanoseconds to seconds
        const seconds = nanoseconds / 1000000000;
        
        if (seconds < 60) return `${seconds.toFixed(1)}s`;
        if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
        return `${(seconds / 3600).toFixed(1)}h`;
    };

    const getSuccessRate = (completed: number, total: number) => {
        if (total === 0) return '0.0';
        return ((completed / total) * 100).toFixed(1);
    };

    return (
        <div>
            <h1>Workflow Statistics</h1>

            <div className="card">
                {stats.length === 0 ? (
                    <p>No statistics available</p>
                ) : (
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
                        {stats.map((stat, index) => (
                            <tr key={`${stat.workflow_name}-${stat.version}`}>
                                <td>{stat.workflow_name}</td>
                                <td>v{stat.version}</td>
                                <td>{stat.total_instances}</td>
                                <td>{stat.completed_instances}</td>
                                <td>{stat.failed_instances}</td>
                                <td>{stat.running_instances}</td>
                                <td>
                                    {getSuccessRate(stat.completed_instances, stat.total_instances)}%
                                </td>
                                <td>{formatDuration(stat.average_duration)}</td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                )}
            </div>

            <div className="card">
                <h2>Summary</h2>
                <div className="stats-grid">
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => sum + stat.total_instances, 0)}
                        </div>
                        <div className="stat-label">Total Instances</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => sum + stat.completed_instances, 0)}
                        </div>
                        <div className="stat-label">Completed</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => sum + stat.failed_instances, 0)}
                        </div>
                        <div className="stat-label">Failed</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => sum + stat.running_instances, 0)}
                        </div>
                        <div className="stat-label">Running</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.length}
                        </div>
                        <div className="stat-label">Workflow Types</div>
                    </div>
                </div>
            </div>
        </div>
    );
};
