import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

interface DeadLetterItem {
  id: number;
  instance_id: number;
  workflow_id: string;
  step_id: number;
  step_name: string;
  step_type: string;
  input: any;
  error: string | null;
  reason: string;
  created_at: string;
}

interface DLQListResponse {
  items: DeadLetterItem[];
  page: number;
  page_size: number;
  total: number;
}

export const DLQ: React.FC = () => {
  const [dlqItems, setDlqItems] = useState<DeadLetterItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [totalItems, setTotalItems] = useState(0);

  useEffect(() => {
    fetchDLQItems();
  }, [currentPage]);

  const fetchDLQItems = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/dlq?page=${currentPage}&page_size=${pageSize}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch DLQ items');
      }
      
      const data: DLQListResponse = await response.json();
      setDlqItems(data.items);
      setTotalItems(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const totalPages = Math.ceil(totalItems / pageSize);

  if (loading) {
    return <div className="loading">Loading DLQ items...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Dead Letter Queue</h1>
        <div style={{ color: '#6b7280', fontSize: '0.9rem' }}>
          Total: {totalItems} items
        </div>
      </div>

      <div className="card">
        {dlqItems.length === 0 ? (
          <p>No items in Dead Letter Queue</p>
        ) : (
          <>
            <table className="table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Instance ID</th>
                  <th>Workflow</th>
                  <th>Step</th>
                  <th>Type</th>
                  <th>Error</th>
                  <th>Reason</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {dlqItems.map((item) => (
                  <tr key={item.id}>
                    <td>{item.id}</td>
                    <td>
                      <Link to={`/instances/${item.instance_id}`} className="btn btn-primary" style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}>
                        {item.instance_id}
                      </Link>
                    </td>
                    <td>{item.workflow_id}</td>
                    <td>{item.step_name}</td>
                    <td>
                      <span className="status failed">{item.step_type}</span>
                    </td>
                    <td>
                      {item.error ? (
                        <div style={{ maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          {item.error}
                        </div>
                      ) : '-'}
                    </td>
                    <td>{item.reason}</td>
                    <td>{new Date(item.created_at).toLocaleString()}</td>
                    <td>
                      <Link to={`/dlq/${item.id}`} className="btn btn-primary">
                        View & Requeue
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Pagination */}
            {totalPages > 1 && (
              <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', marginTop: '2rem' }}>
                <button
                  className="btn btn-secondary"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1}
                >
                  Previous
                </button>
                <span style={{ display: 'flex', alignItems: 'center', color: '#6b7280' }}>
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  className="btn btn-secondary"
                  onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                  disabled={currentPage === totalPages}
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};
