import React, { useState, useEffect } from 'react';
import { Eye, X } from 'lucide-react';
import apiClient, { LDAPSyncLog, LDAPSyncLogsResult } from '../../utils/api';

const LDAPLogsTab: React.FC = () => {
  const [logs, setLogs] = useState<LDAPSyncLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [filters, setFilters] = useState({
    level: '',
    syncId: '',
    username: '',
    from: '',
    to: '',
    limit: 50
  });

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [selectedLog, setSelectedLog] = useState<LDAPSyncLog | null>(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  useEffect(() => {
    fetchLogs();
  }, [filters, page, rowsPerPage]);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      setError(null);

      const fromDate = filters.from ? new Date(filters.from).toISOString() : undefined;
      const toDate = filters.to ? new Date(filters.to).toISOString() : undefined;

      const response = await apiClient.getLDAPSyncLogs(
        filters.limit,
        filters.level || undefined,
        filters.syncId || undefined,
        filters.username || undefined,
        fromDate,
        toDate
      );

      setLogs(response.data.logs);
      setTotal(response.data.total);
    } catch (err: any) {
      console.error('Error fetching LDAP sync logs:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to fetch LDAP sync logs';
      setError(errorMessage);
      setLogs([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (field: keyof typeof filters) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const value = field === 'limit' ? parseInt(e.target.value) || 50 : e.target.value;
    setFilters(prev => ({
      ...prev,
      [field]: value
    }));
    setPage(0);
  };

  const handleChangePage = (newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setRowsPerPage(parseInt(e.target.value, 10));
    setPage(0);
  };

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
        return 'bg-red-100 text-red-800 dark:bg-red-950/30 dark:text-red-400';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950/30 dark:text-yellow-400';
      case 'info':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-950/30 dark:text-blue-400';
      default:
        return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-400';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const handleViewDetails = async (logId: number) => {
    try {
      const response = await apiClient.getLDAPSyncLogDetails(logId);
      setSelectedLog(response.data);
      setShowDetailsModal(true);
    } catch (err: any) {
      console.error('Error fetching log details:', err);
    }
  };

  const paginatedLogs = logs.slice(page * rowsPerPage, (page + 1) * rowsPerPage);
  const totalPages = Math.ceil(logs.length / rowsPerPage);

  return (
    <div>
      <div className="mb-6 pb-4 border-b border-slate-200 dark:border-[#3e3e42]">
        <h2 className="text-xl font-semibold mb-2 text-slate-900 dark:text-[#ff6b35]">
          LDAP Sync Logs
        </h2>
        <p className="text-sm text-slate-600 dark:text-[#ff4500] max-w-2xl">
          View detailed logs of LDAP synchronization operations and troubleshoot issues.
        </p>
      </div>

      {error && (
        <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* Filters */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold mb-4">Filters</h3>

        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <div>
            <label htmlFor="level" className="block text-sm font-medium mb-2">
              Level
            </label>
            <select
              id="level"
              value={filters.level}
              onChange={handleFilterChange('level')}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            >
              <option value="">All Levels</option>
              <option value="error">Error</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>

          <div>
            <label htmlFor="syncId" className="block text-sm font-medium mb-2">
              Sync ID
            </label>
            <input
              type="text"
              id="syncId"
              value={filters.syncId}
              onChange={handleFilterChange('syncId')}
              placeholder="Enter sync ID"
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            />
          </div>

          <div>
            <label htmlFor="username" className="block text-sm font-medium mb-2">
              Username
            </label>
            <input
              type="text"
              id="username"
              value={filters.username}
              onChange={handleFilterChange('username')}
              placeholder="Enter username"
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            />
          </div>

          <div>
            <label htmlFor="limit" className="block text-sm font-medium mb-2">
              Limit
            </label>
            <input
              type="number"
              id="limit"
              value={filters.limit}
              onChange={handleFilterChange('limit')}
              min={1}
              max={1000}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            />
          </div>

          <div>
            <label htmlFor="from" className="block text-sm font-medium mb-2">
              From Date
            </label>
            <input
              type="datetime-local"
              id="from"
              value={filters.from}
              onChange={handleFilterChange('from')}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            />
          </div>

          <div>
            <label htmlFor="to" className="block text-sm font-medium mb-2">
              To Date
            </label>
            <input
              type="datetime-local"
              id="to"
              value={filters.to}
              onChange={handleFilterChange('to')}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={fetchLogs}
              className="btn btn-secondary"
            >
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Logs Table */}
      <div className="card">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="w-8 h-8 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-[#3e3e42]">
                    <th className="text-left px-4 py-3 text-sm font-semibold">Timestamp</th>
                    <th className="text-left px-4 py-3 text-sm font-semibold">Level</th>
                    <th className="text-left px-4 py-3 text-sm font-semibold">Message</th>
                    <th className="text-left px-4 py-3 text-sm font-semibold">Username</th>
                    <th className="text-left px-4 py-3 text-sm font-semibold">Sync ID</th>
                    <th className="text-left px-4 py-3 text-sm font-semibold">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedLogs.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="px-4 py-8 text-center">
                        <p className="text-slate-500 dark:text-[#ff4500]">No logs found</p>
                      </td>
                    </tr>
                  ) : (
                    paginatedLogs.map((log) => (
                      <tr
                        key={log.id}
                        className="border-b border-slate-100 dark:border-[#3e3e42]/50 hover:bg-slate-50 dark:hover:bg-[#2d2d30]"
                      >
                        <td className="px-4 py-3">
                          <span className="text-sm text-slate-900 dark:text-slate-300">
                            {formatDate(log.timestamp)}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getLevelColor(log.level)}`}>
                            {log.level || 'unknown'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <p className="text-sm text-slate-900 dark:text-slate-300 max-w-xs truncate">
                            {log.message || ''}
                          </p>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-sm text-slate-900 dark:text-slate-300">
                            {log.username || '-'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-xs font-mono text-slate-600 dark:text-[#ff4500]">
                            {log.sync_session_id || '-'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {(log.details || log.stack_trace) && (
                            <button
                              onClick={() => handleViewDetails(log.id)}
                              className="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-[#2d2d30] text-slate-600 dark:text-[#ff4500]"
                              title="View Details"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between px-4 py-3 border-t border-slate-200 dark:border-[#3e3e42]">
              <div className="flex items-center gap-4">
                <label className="text-sm text-slate-600 dark:text-[#ff4500]">
                  Rows per page:
                </label>
                <select
                  value={rowsPerPage}
                  onChange={handleChangeRowsPerPage}
                  className="px-3 py-1.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">
                  Showing {page * rowsPerPage + 1} to {Math.min((page + 1) * rowsPerPage, logs.length)} of {logs.length}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleChangePage(page - 1)}
                  disabled={page === 0}
                  className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">
                  Page {page + 1} of {totalPages || 1}
                </span>
                <button
                  onClick={() => handleChangePage(page + 1)}
                  disabled={page >= totalPages - 1}
                  className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Details Modal */}
      {showDetailsModal && selectedLog && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-2xl w-full bg-white dark:bg-[#1e1e1e] max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Log Details</h2>
              <button
                onClick={() => {
                  setShowDetailsModal(false);
                  setSelectedLog(null);
                }}
                className="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-[#2d2d30] text-slate-600 dark:text-[#ff4500]"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Timestamp</label>
                <p className="text-slate-900 dark:text-slate-300">{formatDate(selectedLog.timestamp)}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Level</label>
                <p>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getLevelColor(selectedLog.level)}`}>
                    {selectedLog.level}
                  </span>
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Message</label>
                <p className="text-slate-900 dark:text-slate-300">{selectedLog.message}</p>
              </div>
              {selectedLog.username && (
                <div>
                  <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Username</label>
                  <p className="text-slate-900 dark:text-slate-300">{selectedLog.username}</p>
                </div>
              )}
              {selectedLog.details && (
                <div>
                  <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Details</label>
                  <p className="text-slate-900 dark:text-slate-300 whitespace-pre-wrap">{selectedLog.details}</p>
                </div>
              )}
              {selectedLog.stack_trace && (
                <div>
                  <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">Stack Trace</label>
                  <pre className="text-xs text-slate-900 dark:text-slate-300 bg-slate-100 dark:bg-[#2d2d30] p-3 rounded overflow-x-auto">
                    {selectedLog.stack_trace}
                  </pre>
                </div>
              )}
              {selectedLog.ldap_error_code && (
                <div>
                  <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">LDAP Error Code</label>
                  <p className="text-slate-900 dark:text-slate-300">{selectedLog.ldap_error_code}</p>
                </div>
              )}
              {selectedLog.ldap_error_message && (
                <div>
                  <label className="text-sm font-medium text-slate-600 dark:text-[#ff4500]">LDAP Error Message</label>
                  <p className="text-slate-900 dark:text-slate-300">{selectedLog.ldap_error_message}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default LDAPLogsTab;

