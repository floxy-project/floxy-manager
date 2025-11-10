import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import { X, Loader2 } from 'lucide-react';
import { Modal } from '../Modal';
import apiClient, { LDAPSyncStatus, LDAPSyncProgress, LDAPConfig } from '../../utils/api';
import LDAPSyncProgressBar from '../LDAPSyncProgressBar';

const LDAPSyncTab: React.FC = () => {
  const [syncStarted, setSyncStarted] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const [ldapConfig, setLdapConfig] = useState<LDAPConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  
  const [syncStatus, setSyncStatus] = useState<LDAPSyncStatus | null>(null);
  const [statusLoading, setStatusLoading] = useState(true);
  
  const [syncProgress, setSyncProgress] = useState<LDAPSyncProgress | null>(null);
  const [progressLoading, setProgressLoading] = useState(false);

  const isRunningRef = useRef(false);
  const syncStartedRef = useRef(false);

  const fetchLDAPConfig = useCallback(async () => {
    try {
      setConfigLoading(true);
      const response = await apiClient.getLDAPConfig();
      setLdapConfig(response.data);
    } catch (err: any) {
      console.error('Error fetching LDAP config:', err);
      setLdapConfig({
        enabled: false,
        url: '',
        bind_dn: '',
        bind_password: '',
        user_base_dn: '',
        user_filter: '',
        user_name_attr: '',
        user_email_attr: '',
        start_tls: false,
        insecure_tls: false,
        timeout: '30s',
        sync_interval: 3600,
      });
    } finally {
      setConfigLoading(false);
    }
  }, []);

  const fetchSyncStatus = useCallback(async (showLoading = false) => {
    try {
      if (showLoading) {
        setStatusLoading(true);
      }
      const response = await apiClient.getLDAPSyncStatus();
      const newStatus = response.data;
      
      // Only update state if data actually changed to prevent unnecessary re-renders
      setSyncStatus(prev => {
        if (!prev) return newStatus;
        // Compare key fields to avoid unnecessary updates
        if (
          prev.status === newStatus.status &&
          prev.is_running === newStatus.is_running &&
          prev.total_users === newStatus.total_users &&
          prev.synced_users === newStatus.synced_users &&
          prev.errors === newStatus.errors &&
          prev.warnings === newStatus.warnings &&
          prev.last_sync_time === newStatus.last_sync_time &&
          prev.last_sync_duration === newStatus.last_sync_duration
        ) {
          return prev; // Return previous state to prevent re-render
        }
        return newStatus;
      });
    } catch (err: any) {
      console.error('Error fetching LDAP sync status:', err);
      setSyncStatus(prev => {
        const errorStatus = {
          status: '',
          is_running: false,
          total_users: 0,
          synced_users: 0,
          errors: 0,
          warnings: 0,
        };
        // Only update if state actually changed
        if (!prev || prev.is_running !== false) {
          return errorStatus;
        }
        return prev;
      });
    } finally {
      if (showLoading) {
        setStatusLoading(false);
      }
    }
  }, []);

  const fetchSyncProgress = useCallback(async (showLoading = false) => {
    try {
      if (showLoading) {
        setProgressLoading(true);
      }
      const response = await apiClient.getLDAPSyncProgress();
      const newProgress = response.data;
      
      // Only update state if data actually changed to prevent unnecessary re-renders
      setSyncProgress(prev => {
        if (!prev) return newProgress;
        // Compare key fields to avoid unnecessary updates
        if (
          prev.is_running === newProgress.is_running &&
          prev.progress === newProgress.progress &&
          prev.current_step === newProgress.current_step &&
          prev.processed_items === newProgress.processed_items &&
          prev.total_items === newProgress.total_items &&
          prev.estimated_time === newProgress.estimated_time &&
          prev.start_time === newProgress.start_time
        ) {
          return prev; // Return previous state to prevent re-render
        }
        return newProgress;
      });
    } catch (err: any) {
      console.error('Error fetching LDAP sync progress:', err);
      setSyncProgress(prev => {
        const errorProgress = {
          is_running: false,
          progress: 0,
          processed_items: 0,
          total_items: 0,
        };
        // Only update if state actually changed
        if (!prev || prev.is_running !== false) {
          return errorProgress;
        }
        return prev;
      });
    } finally {
      if (showLoading) {
        setProgressLoading(false);
      }
    }
  }, []);

  // Update refs when state changes
  useEffect(() => {
    isRunningRef.current = !!syncStatus?.is_running;
    syncStartedRef.current = syncStarted;
  }, [syncStatus?.is_running, syncStarted]);

  // Fetch LDAP config
  useEffect(() => {
    fetchLDAPConfig();
  }, [fetchLDAPConfig]);

  // Fetch sync status and progress
  useEffect(() => {
    // Initial fetch with loading indicator
    fetchSyncStatus(true);
    
    const intervalId = setInterval(() => {
      // Periodic fetch without loading indicator to avoid flickering
      fetchSyncStatus(false);
      if (isRunningRef.current || syncStartedRef.current) {
        fetchSyncProgress(false);
      }
    }, 1000);
    
    return () => clearInterval(intervalId);
  }, [fetchSyncStatus, fetchSyncProgress]);

  // Keep modal open for a few seconds after sync completes to show results
  useEffect(() => {
    if (!syncStatus?.is_running && syncStarted) {
      // Keep modal open for 3 seconds after sync completes to show final results
      const timer = setTimeout(() => {
        setSyncStarted(false);
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [syncStatus?.is_running, syncStarted]);

  const handleSyncUsers = useCallback(async () => {
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      const response = await apiClient.syncLDAPUsers();
      setSuccess('User synchronization started successfully');
      setTimeout(() => setSuccess(null), 3000);
      setSyncStarted(true);
      setTimeout(() => {
        fetchSyncStatus(false);
        fetchSyncProgress(false);
      }, 1000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to start user synchronization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, [fetchSyncStatus, fetchSyncProgress]);

  const handleCancelSync = useCallback(async () => {
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      await apiClient.cancelLDAPSync();
      setSuccess('Synchronization cancelled successfully');
      setTimeout(() => setSuccess(null), 3000);
      fetchSyncStatus(false);
      fetchSyncProgress(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to cancel synchronization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, [fetchSyncStatus, fetchSyncProgress]);

  const handleTestConnection = useCallback(async () => {
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      const response = await apiClient.testLDAPConnection({
        url: ldapConfig?.url || '',
        bind_dn: ldapConfig?.bind_dn || '',
        bind_password: ldapConfig?.bind_password || '',
        user_base_dn: ldapConfig?.user_base_dn || '',
        user_filter: ldapConfig?.user_filter || '',
        user_name_attr: ldapConfig?.user_name_attr || '',
        start_tls: ldapConfig?.start_tls || false,
        insecure_tls: ldapConfig?.insecure_tls || false,
        timeout: ldapConfig?.timeout || '30s',
      });
      
      if (response.data.success) {
        setSuccess('LDAP connection test successful');
      } else {
        setError(response.data.message || 'LDAP connection test failed');
      }
      setTimeout(() => {
        setSuccess(null);
        setError(null);
      }, 3000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'LDAP connection test failed';
      setError(errorMessage);
      setTimeout(() => setError(null), 3000);
    } finally {
      setLoading(false);
    }
  }, [ldapConfig]);

  const formatDate = useCallback((dateString: string | null | undefined) => {
    if (!dateString) return 'Never';
    
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime()) || date.getFullYear() === 1) {
        return 'Never';
      }
      return date.toLocaleString();
    } catch {
      return 'Never';
    }
  }, []);

  const getStatusColor = useCallback((status: string | undefined, isRunning: boolean | undefined) => {
    if (isRunning) return 'bg-blue-100 text-blue-800 dark:bg-blue-950/30 dark:text-blue-400';
    if (!status) return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-400';
    
    switch (status.toLowerCase()) {
      case 'completed':
        return 'bg-green-100 text-green-800 dark:bg-green-950/30 dark:text-green-400';
      case 'failed':
        return 'bg-red-100 text-red-800 dark:bg-red-950/30 dark:text-red-400';
      case 'cancelled':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950/30 dark:text-yellow-400';
      case 'running':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-950/30 dark:text-blue-400';
      default:
        return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-400';
    }
  }, []);

  const getStatusLabel = useCallback((status: string | undefined, isRunning: boolean | undefined) => {
    if (isRunning) return 'Running';
    if (!status) return 'Idle';
    return status.charAt(0).toUpperCase() + status.slice(1);
  }, []);

  const isLDAPEnabled = useMemo(() => {
    return ldapConfig?.enabled && ldapConfig?.url && ldapConfig?.bind_dn;
  }, [ldapConfig?.enabled, ldapConfig?.url, ldapConfig?.bind_dn]);

  const isModalOpen = useMemo(() => {
    return !!syncStatus?.is_running || (syncStarted && !syncStatus?.is_running);
  }, [syncStatus?.is_running, syncStarted]);

  const handleCloseModal = useCallback(() => {
    if (!syncStatus?.is_running) {
      setSyncStarted(false);
      fetchSyncStatus(false);
    }
  }, [syncStatus?.is_running, fetchSyncStatus]);

  // Memoize computed values to prevent unnecessary re-renders
  const statusColor = useMemo(() => getStatusColor(syncStatus?.status, syncStatus?.is_running), [syncStatus?.status, syncStatus?.is_running]);
  const statusLabel = useMemo(() => getStatusLabel(syncStatus?.status, syncStatus?.is_running), [syncStatus?.status, syncStatus?.is_running]);
  const formattedLastSyncTime = useMemo(() => formatDate(syncStatus?.last_sync_time), [syncStatus?.last_sync_time]);

  // Memoize progress bar props to prevent unnecessary re-renders
  const progressBarProps = useMemo(() => ({
    isRunning: !!syncProgress?.is_running,
    progress: typeof syncProgress?.progress === 'number' ? syncProgress.progress : 0,
    currentStep: syncProgress?.current_step || undefined,
    processedItems: syncProgress?.processed_items || undefined,
    totalItems: syncProgress?.total_items || undefined,
    estimatedTime: syncProgress?.estimated_time || undefined,
    startTime: syncProgress?.start_time || undefined,
  }), [
    syncProgress?.is_running,
    syncProgress?.progress,
    syncProgress?.current_step,
    syncProgress?.processed_items,
    syncProgress?.total_items,
    syncProgress?.estimated_time,
    syncProgress?.start_time,
  ]);

  if (configLoading || statusLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-8 h-8 text-slate-600 dark:text-slate-400 animate-spin" />
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 pb-4 border-b border-slate-200 dark:border-[#3e3e42]">
        <h2 className="text-xl font-semibold mb-2 text-slate-900 dark:text-[#ff6b35]">
          LDAP Synchronization
        </h2>
        <p className="text-sm text-slate-600 dark:text-[#ff4500] max-w-2xl">
          Manage LDAP synchronization settings and manually trigger synchronization of users.
        </p>
      </div>

      {error && (
        <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
        </div>
      )}

      {success && (
        <div className="mb-4 p-4 rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800/50">
          <p className="text-sm text-green-700 dark:text-green-400">{success}</p>
        </div>
      )}

      {/* Sync Controls */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold mb-4">Synchronization Controls</h3>

        {!isLDAPEnabled && (
          <div className="mb-4 p-4 rounded-lg bg-yellow-50 dark:bg-yellow-950/30 border border-yellow-200 dark:border-yellow-800/50">
            <p className="text-sm text-yellow-700 dark:text-yellow-400">
              LDAP is not configured or disabled. Please configure LDAP in the Configuration tab first.
            </p>
          </div>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
          <button
            onClick={handleSyncUsers}
            disabled={syncStatus?.is_running || loading || !isLDAPEnabled}
            className="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <Loader2 className="w-4 h-4 animate-spin" />
                Starting...
              </span>
            ) : (
              'Sync Users'
            )}
          </button>
        </div>

      </div>

      {/* Sync Progress Modal */}
      <Modal 
        isOpen={isModalOpen} 
        onClose={handleCloseModal}
      >
        <div className="flex items-start justify-between px-6 pt-6 pb-4 border-b border-slate-200 relative">
          <h3 className="text-xl font-bold text-slate-900 pr-8">LDAP Synchronization Progress</h3>
          {!syncStatus?.is_running && (
            <button 
              className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors text-slate-600 hover:text-slate-900"
              onClick={handleCloseModal}
              aria-label="Close"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>
        
        <div className="px-6 py-6 flex-1 overflow-y-auto">
          {progressLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-8 h-8 text-slate-600 dark:text-slate-400 animate-spin" />
            </div>
          ) : (
            <LDAPSyncProgressBar {...progressBarProps} />
          )}
        </div>
        
        <div className="px-6 py-4 border-t border-slate-200 flex justify-end gap-3">
          {syncStatus?.is_running && (
            <button
              className="btn btn-danger"
              onClick={handleCancelSync}
              disabled={loading}
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Cancelling...
                </span>
              ) : (
                'Cancel Sync'
              )}
            </button>
          )}
          {!syncStatus?.is_running && (
            <button
              className="btn btn-primary"
              onClick={handleCloseModal}
            >
              Close
            </button>
          )}
        </div>
      </Modal>

      {/* Sync Status */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold mb-4">Sync Status</h3>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="card border border-slate-200 dark:border-[#3e3e42]">
            <div className="flex justify-between items-center mb-3">
              <h4 className="font-semibold text-slate-900 dark:text-[#ff6b35]">Last Sync</h4>
              <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                {statusLabel}
              </span>
            </div>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Last Run:</span>
                <span className="text-sm text-slate-900 dark:text-slate-300">{formattedLastSyncTime}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Duration:</span>
                <span className="text-sm text-slate-900 dark:text-slate-300">{syncStatus?.last_sync_duration || 'N/A'}</span>
              </div>
            </div>
          </div>

          <div className="card border border-slate-200 dark:border-[#3e3e42]">
            <h4 className="font-semibold mb-3 text-slate-900 dark:text-[#ff6b35]">Statistics</h4>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Total Users:</span>
                <span className="text-sm text-slate-900 dark:text-slate-300">{syncStatus?.total_users || 0}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Synced Users:</span>
                <span className="text-sm text-slate-900 dark:text-slate-300">{syncStatus?.synced_users || 0}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Errors:</span>
                <span className="text-sm text-red-600 dark:text-red-400">{syncStatus?.errors || 0}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Warnings:</span>
                <span className="text-sm text-yellow-600 dark:text-yellow-400">{syncStatus?.warnings || 0}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Connection Test */}
      <div className="card">
        <h3 className="text-lg font-semibold mb-4">Connection Test</h3>

        <p className="text-sm text-slate-600 dark:text-[#ff4500] mb-4">
          Test the connection to your LDAP server with the current configuration.
        </p>

        {!isLDAPEnabled && (
          <div className="mb-4 p-4 rounded-lg bg-yellow-50 dark:bg-yellow-950/30 border border-yellow-200 dark:border-yellow-800/50">
            <p className="text-sm text-yellow-700 dark:text-yellow-400">
              LDAP is not configured or disabled. Please configure LDAP in the Configuration tab first.
            </p>
          </div>
        )}

        <button
          onClick={handleTestConnection}
          disabled={loading || !isLDAPEnabled}
          className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? (
            <span className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              Testing...
            </span>
          ) : (
            'Test Connection'
          )}
        </button>
      </div>
    </div>
  );
};

export default LDAPSyncTab;

