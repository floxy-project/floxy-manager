import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { authFetch } from '../utils/api';
import { useRBAC } from '../auth/permissions';
import { Navigate } from 'react-router-dom';
import { AlertCircle, Loader2 } from 'lucide-react';

interface AuditLogEntry {
  id: number;
  entity: string;
  entity_id: string;
  username: string;
  action: string;
  created_at: string;
}

interface AuditLogListResponse {
  items: AuditLogEntry[];
  page: number;
  page_size: number;
  total: number;
}

export const AuditLog: React.FC = () => {
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const [auditLog, setAuditLog] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [totalItems, setTotalItems] = useState(0);
  const rbac = useRBAC(projectId);

  useEffect(() => {
    if (tenantId && projectId) {
      fetchAuditLog();
    }
  }, [tenantId, projectId, currentPage]);

  const fetchAuditLog = async () => {
    try {
      setLoading(true);
      const response = await authFetch(`/api/v1/audit-log?project_id=${projectId}&page=${currentPage}&page_size=${pageSize}`);
      
      if (!response.ok) {
        if (response.status === 403) {
          setError('Access denied to audit log');
          return;
        }
        throw new Error('Failed to fetch audit log');
      }
      
      const data: AuditLogListResponse = await response.json();
      setAuditLog(data.items);
      setTotalItems(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  if (!rbac.isSuperuser && !rbac.canViewAudit()) {
    return <Navigate to={`/tenants/${tenantId}/projects/${projectId}/dashboard`} replace />;
  }

  const totalPages = Math.ceil(totalItems / pageSize);

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <Loader2 className="w-5 h-5 animate-spin" />
          <span>Loading audit log...</span>
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

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case 'create':
        return 'text-green-600 dark:text-green-400';
      case 'update':
        return 'text-blue-600 dark:text-blue-400';
      case 'delete':
        return 'text-red-600 dark:text-red-400';
      case 'archive':
        return 'text-orange-600 dark:text-orange-400';
      default:
        return 'text-slate-600 dark:text-slate-400';
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1>Audit Log</h1>
        <div className="px-4 py-2 rounded-lg" style={{
          background: 'rgba(255, 255, 255, 0.5)',
          backdropFilter: 'blur(12px) saturate(150%)',
          WebkitBackdropFilter: 'blur(12px) saturate(150%)',
          border: '1px solid rgba(255, 255, 255, 0.4)',
          boxShadow: '0 4px 16px rgba(0, 0, 0, 0.06)'
        }}>
          <span className="text-sm font-medium text-slate-700 dark:text-[#ff4500]">
            Total: <span className="font-bold">{totalItems}</span> entries
          </span>
        </div>
      </div>

      {auditLog.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-slate-600 dark:text-slate-400">No audit log entries found</p>
        </div>
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-700">
                  <th className="text-left py-3 px-4 font-semibold text-slate-700 dark:text-slate-300">Time</th>
                  <th className="text-left py-3 px-4 font-semibold text-slate-700 dark:text-slate-300">User</th>
                  <th className="text-left py-3 px-4 font-semibold text-slate-700 dark:text-slate-300">Entity</th>
                  <th className="text-left py-3 px-4 font-semibold text-slate-700 dark:text-slate-300">Entity ID</th>
                  <th className="text-left py-3 px-4 font-semibold text-slate-700 dark:text-slate-300">Action</th>
                </tr>
              </thead>
              <tbody>
                {auditLog.map((entry) => (
                  <tr key={entry.id} className="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50">
                    <td className="py-3 px-4 text-slate-600 dark:text-slate-400">{formatDate(entry.created_at)}</td>
                    <td className="py-3 px-4 text-slate-700 dark:text-slate-300">{entry.username}</td>
                    <td className="py-3 px-4 text-slate-700 dark:text-slate-300">{entry.entity}</td>
                    <td className="py-3 px-4 text-slate-600 dark:text-slate-400">{entry.entity_id}</td>
                    <td className={`py-3 px-4 font-medium ${getActionColor(entry.action)}`}>
                      {entry.action}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-6">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="btn btn-outline"
              >
                Previous
              </button>
              <span className="text-sm text-slate-600 dark:text-slate-400">
                Page {currentPage} of {totalPages}
              </span>
              <button
                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
                className="btn btn-outline"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

