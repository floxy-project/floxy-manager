import React, { useState, useEffect } from 'react';

interface WorkflowStats {
    name: string;
    version: number;
    total_instances: string | number;
    completed: string | number;
    failed: string | number;
    running: string | number;
    avg_duration_seconds: string | number | null;
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

    const formatDuration = (seconds: string | number | null) => {
        if (seconds === null || seconds === '') return '-';
        const numSeconds = typeof seconds === 'string' ? parseFloat(seconds) : seconds;
        if (isNaN(numSeconds)) return '-';
        if (numSeconds < 60) return `${numSeconds.toFixed(1)}s`;
        if (numSeconds < 3600) return `${(numSeconds / 60).toFixed(1)}m`;
        return `${(numSeconds / 3600).toFixed(1)}h`;
    };

    const getSuccessRate = (completed: string | number, total: string | number) => {
        const numCompleted = typeof completed === 'string' ? parseInt(completed) : completed;
        const numTotal = typeof total === 'string' ? parseInt(total) : total;
        if (numTotal === 0) return '0.0';
        return ((numCompleted / numTotal) * 100).toFixed(1);
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
                            <tr key={`${stat.name}-${stat.version}`}>
                                <td>{stat.name}</td>
                                <td>v{stat.version}</td>
                                <td>{stat.total_instances}</td>
                                <td>{stat.completed}</td>
                                <td>{stat.failed}</td>
                                <td>{stat.running}</td>
                                <td>
                                    {getSuccessRate(stat.completed, stat.total_instances)}%
                                </td>
                                <td>{formatDuration(stat.avg_duration_seconds)}</td>
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
                            {stats.reduce((sum, stat) => {
                                const num = typeof stat.total_instances === 'string' ? parseInt(stat.total_instances) : stat.total_instances;
                                return sum + (isNaN(num) ? 0 : num);
                            }, 0)}
                        </div>
                        <div className="stat-label">Total Instances</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => {
                                const num = typeof stat.completed === 'string' ? parseInt(stat.completed) : stat.completed;
                                return sum + (isNaN(num) ? 0 : num);
                            }, 0)}
                        </div>
                        <div className="stat-label">Completed</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => {
                                const num = typeof stat.failed === 'string' ? parseInt(stat.failed) : stat.failed;
                                return sum + (isNaN(num) ? 0 : num);
                            }, 0)}
                        </div>
                        <div className="stat-label">Failed</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-number">
                            {stats.reduce((sum, stat) => {
                                const num = typeof stat.running === 'string' ? parseInt(stat.running) : stat.running;
                                return sum + (isNaN(num) ? 0 : num);
                            }, 0)}
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
