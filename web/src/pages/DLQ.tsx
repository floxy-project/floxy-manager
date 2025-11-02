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
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading DLQ items...</span>
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

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1>Dead Letter Queue</h1>
        <div className="px-4 py-2 rounded-lg" style={{
          background: 'rgba(255, 255, 255, 0.5)',
          backdropFilter: 'blur(12px) saturate(150%)',
          WebkitBackdropFilter: 'blur(12px) saturate(150%)',
          border: '1px solid rgba(255, 255, 255, 0.4)',
          boxShadow: '0 4px 16px rgba(0, 0, 0, 0.06)'
        }}>
          <span className="text-sm font-medium text-slate-700 dark:text-[#ff4500]400">
            Total: <span className="font-bold">{totalItems}</span> items
          </span>
        </div>
      </div>

      <div className="card">
        {dlqItems.length === 0 ? (
          <div className="text-center py-8 text-slate-500 dark:text-[#ff4500]500">
            <svg className="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p>No items in Dead Letter Queue</p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
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
                      <td className="font-mono text-xs">{item.id}</td>
                      <td>
                        <Link 
                          to={`/instances/${item.instance_id}`} 
                          className="btn btn-primary text-xs py-1 px-2 font-mono"
                        >
                          {item.instance_id}
                        </Link>
                      </td>
                      <td>
                        <div className="font-medium">{item.workflow_id}</div>
                      </td>
                      <td>
                        <div className="font-medium">{item.step_name}</div>
                        <div className="text-xs text-slate-500 dark:text-[#ff4500]500 font-mono mt-0.5">
                          Step #{item.step_id}
                        </div>
                      </td>
                      <td>
                        <span className="status failed">{item.step_type}</span>
                      </td>
                      <td className="max-w-xs">
                        {item.error ? (
                          <div className="text-sm text-red-600 dark:text-red-400 truncate" title={item.error}>
                            {item.error}
                          </div>
                        ) : (
                          <span className="text-slate-400">-</span>
                        )}
                      </td>
                      <td className="text-sm text-slate-600 dark:text-[#ff4500]500">{item.reason}</td>
                      <td className="text-sm text-slate-600 dark:text-[#ff4500]500">
                        {new Date(item.created_at).toLocaleString()}
                      </td>
                      <td>
                        <Link to={`/dlq/${item.id}`} className="btn btn-primary text-xs py-1.5 px-3">
                          View & Requeue
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
