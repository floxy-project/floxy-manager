import React, { useEffect, useState } from 'react';
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

  // Fetch LDAP config
  useEffect(() => {
    fetchLDAPConfig();
  }, []);

  // Fetch sync status and progress
  useEffect(() => {
    fetchSyncStatus();
    const interval = setInterval(() => {
      fetchSyncStatus();
      if (syncStatus?.is_running || syncStarted) {
        fetchSyncProgress();
      }
    }, syncStatus?.is_running || syncStarted ? 1000 : 5000);
    
    return () => clearInterval(interval);
  }, [syncStatus?.is_running, syncStarted]);

  // Reset syncStarted when sync is no longer running
  useEffect(() => {
    if (!syncStatus?.is_running && syncStarted) {
      setSyncStarted(false);
    }
  }, [syncStatus?.is_running, syncStarted]);

  const fetchLDAPConfig = async () => {
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
  };

  const fetchSyncStatus = async () => {
    try {
      setStatusLoading(true);
      const response = await apiClient.getLDAPSyncStatus();
      setSyncStatus(response.data);
    } catch (err: any) {
      console.error('Error fetching LDAP sync status:', err);
      setSyncStatus({
        status: '',
        is_running: false,
        total_users: 0,
        synced_users: 0,
        errors: 0,
        warnings: 0,
      });
    } finally {
      setStatusLoading(false);
    }
  };

  const fetchSyncProgress = async () => {
    try {
      setProgressLoading(true);
      const response = await apiClient.getLDAPSyncProgress();
      setSyncProgress(response.data);
    } catch (err: any) {
      console.error('Error fetching LDAP sync progress:', err);
      setSyncProgress({
        is_running: false,
        progress: 0,
        processed_items: 0,
        total_items: 0,
      });
    } finally {
      setProgressLoading(false);
    }
  };

  const handleSyncUsers = async () => {
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      const response = await apiClient.syncLDAPUsers();
      setSuccess('User synchronization started successfully');
      setTimeout(() => setSuccess(null), 3000);
      setSyncStarted(true);
      setTimeout(() => {
        fetchSyncStatus();
        fetchSyncProgress();
      }, 1000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to start user synchronization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleCancelSync = async () => {
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      await apiClient.cancelLDAPSync();
      setSuccess('Synchronization cancelled successfully');
      setTimeout(() => setSuccess(null), 3000);
      fetchSyncStatus();
      fetchSyncProgress();
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to cancel synchronization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async () => {
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
  };

  const formatDate = (dateString: string | null | undefined) => {
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
  };

  const getStatusColor = (status: string | undefined, isRunning: boolean | undefined) => {
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
  };

  const getStatusLabel = (status: string | undefined, isRunning: boolean | undefined) => {
    if (isRunning) return 'Running';
    if (!status) return 'Idle';
    return status.charAt(0).toUpperCase() + status.slice(1);
  };

  const isLDAPEnabled = ldapConfig?.enabled && ldapConfig?.url && ldapConfig?.bind_dn;

  if (configLoading || statusLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="w-8 h-8 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
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
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                Starting...
              </span>
            ) : (
              'Sync Users'
            )}
          </button>
        </div>

        {syncStatus?.is_running && (
          <button
            onClick={handleCancelSync}
            disabled={loading || !isLDAPEnabled}
            className="btn btn-secondary text-red-600 dark:text-red-400 border-red-600 dark:border-red-400 hover:bg-red-50 dark:hover:bg-red-950/30"
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <div className="w-4 h-4 border-2 border-red-600 dark:border-red-400 border-t-transparent rounded-full animate-spin"></div>
                Cancelling...
              </span>
            ) : (
              'Cancel Sync'
            )}
          </button>
        )}
      </div>

      {/* Sync Progress */}
      {syncStatus?.is_running && (
        progressLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="w-8 h-8 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          </div>
        ) : (
          <LDAPSyncProgressBar
            isRunning={!!syncProgress?.is_running}
            progress={typeof syncProgress?.progress === 'number' ? syncProgress.progress : 0}
            currentStep={syncProgress?.current_step}
            processedItems={syncProgress?.processed_items}
            totalItems={syncProgress?.total_items}
            estimatedTime={syncProgress?.estimated_time}
            startTime={syncProgress?.start_time || undefined}
          />
        )
      )}

      {/* Sync Status */}
      <div className="card mb-6">
        <h3 className="text-lg font-semibold mb-4">Sync Status</h3>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="card border border-slate-200 dark:border-[#3e3e42]">
            <div className="flex justify-between items-center mb-3">
              <h4 className="font-semibold text-slate-900 dark:text-[#ff6b35]">Last Sync</h4>
              <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(syncStatus?.status, syncStatus?.is_running)}`}>
                {getStatusLabel(syncStatus?.status, syncStatus?.is_running)}
              </span>
            </div>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-slate-600 dark:text-[#ff4500]">Last Run:</span>
                <span className="text-sm text-slate-900 dark:text-slate-300">{formatDate(syncStatus?.last_sync_time)}</span>
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
              <div className="w-4 h-4 border-2 border-slate-600 dark:border-slate-400 border-t-transparent rounded-full animate-spin"></div>
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

