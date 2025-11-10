import React, { useState, useEffect } from 'react';
import { Loader2, Save, Network } from 'lucide-react';
import apiClient, { LDAPConfig } from '../../utils/api';

const LDAPConfigTab: React.FC = () => {
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [configLoading, setConfigLoading] = useState(true);

  const [formData, setFormData] = useState<LDAPConfig>({
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

  const [backgroundSyncEnabled, setBackgroundSyncEnabled] = useState(true);

  useEffect(() => {
    fetchLDAPConfig();
  }, []);

  const fetchLDAPConfig = async () => {
    try {
      setConfigLoading(true);
      const response = await apiClient.getLDAPConfig();
      const config = response.data;
      setFormData(config);
      setBackgroundSyncEnabled((config.sync_interval || 0) > 0);
    } catch (err: any) {
      console.error('Error fetching LDAP config:', err);
      // Keep default form data
    } finally {
      setConfigLoading(false);
    }
  };

  const handleInputChange = (field: keyof LDAPConfig) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    if (field === 'enabled' || field === 'start_tls' || field === 'insecure_tls') {
      setFormData(prev => ({
        ...prev,
        [field]: e.target.checked
      }));
    } else if (field === 'sync_interval') {
      setFormData(prev => ({
        ...prev,
        [field]: parseInt(e.target.value) || 0
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        [field]: e.target.value
      }));
    }
  };

  const handleBackgroundSyncChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const enabled = e.target.checked;
    setBackgroundSyncEnabled(enabled);
    if (!enabled) {
      setFormData(prev => ({
        ...prev,
        sync_interval: 0
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        sync_interval: prev.sync_interval || 3600
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      const configData = {
        ...formData,
        sync_interval: backgroundSyncEnabled ? formData.sync_interval : 0
      };
      
      await apiClient.updateLDAPConfig(configData);
      setSuccess('LDAP configuration updated successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to update LDAP configuration';
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
        url: formData.url,
        bind_dn: formData.bind_dn,
        bind_password: formData.bind_password,
        user_base_dn: formData.user_base_dn,
        user_filter: formData.user_filter,
        user_name_attr: formData.user_name_attr,
        start_tls: formData.start_tls,
        insecure_tls: formData.insecure_tls,
        timeout: formData.timeout
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

  if (configLoading) {
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
          LDAP Configuration
        </h2>
        <p className="text-sm text-slate-600 dark:text-[#ff4500] max-w-2xl">
          Configure LDAP server connection settings for user synchronization.
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

      <div className="card">
        <form onSubmit={handleSubmit}>
          <h3 className="text-lg font-semibold mb-4">Connection Settings</h3>

          <div className="mb-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.enabled}
                onChange={handleInputChange('enabled')}
                disabled={loading}
                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
              />
              <span className="text-sm font-medium">Enable LDAP Integration</span>
            </label>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            <div>
              <label htmlFor="url" className="block text-sm font-medium mb-2">
                LDAP Server URL *
              </label>
              <input
                type="text"
                id="url"
                value={formData.url}
                onChange={handleInputChange('url')}
                placeholder="ldap://ldap.example.com:389"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">LDAP server URL with protocol and port</p>
            </div>

            <div>
              <label htmlFor="bind_dn" className="block text-sm font-medium mb-2">
                Bind DN *
              </label>
              <input
                type="text"
                id="bind_dn"
                value={formData.bind_dn}
                onChange={handleInputChange('bind_dn')}
                placeholder="cn=admin,dc=example,dc=com"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">DN for binding to LDAP server</p>
            </div>

            <div>
              <label htmlFor="bind_password" className="block text-sm font-medium mb-2">
                Bind Password *
              </label>
              <input
                type="password"
                id="bind_password"
                value={formData.bind_password}
                onChange={handleInputChange('bind_password')}
                placeholder="Password for bind DN"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Password for the bind DN</p>
            </div>

            <div>
              <label htmlFor="timeout" className="block text-sm font-medium mb-2">
                Connection Timeout
              </label>
              <input
                type="text"
                id="timeout"
                value={formData.timeout}
                onChange={handleInputChange('timeout')}
                placeholder="30s"
                disabled={loading}
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Connection timeout (e.g., 30s, 1m)</p>
            </div>
          </div>

          <div className="mb-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={backgroundSyncEnabled}
                onChange={handleBackgroundSyncChange}
                disabled={loading}
                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
              />
              <span className="text-sm font-medium">Enable background synchronization of users</span>
            </label>
          </div>

          {backgroundSyncEnabled && (
            <div className="mb-6">
              <label htmlFor="sync_interval" className="block text-sm font-medium mb-2">
                Sync Interval (seconds)
              </label>
              <input
                type="number"
                id="sync_interval"
                value={formData.sync_interval}
                onChange={handleInputChange('sync_interval')}
                placeholder="3600"
                min={60}
                max={86400}
                disabled={loading}
                className="w-full md:w-1/2 px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Background sync interval in seconds (default: 3600 = 1 hour)</p>
            </div>
          )}

          <div className="mb-6 flex flex-wrap gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.start_tls}
                onChange={handleInputChange('start_tls')}
                disabled={loading}
                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
              />
              <span className="text-sm font-medium">Use StartTLS</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.insecure_tls}
                onChange={handleInputChange('insecure_tls')}
                disabled={loading}
                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
              />
              <span className="text-sm font-medium">Skip TLS Certificate Verification</span>
            </label>
          </div>

          <div className="border-t border-slate-200 dark:border-[#3e3e42] my-6"></div>

          <h3 className="text-lg font-semibold mb-4">User Configuration</h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            <div>
              <label htmlFor="user_base_dn" className="block text-sm font-medium mb-2">
                User Base DN *
              </label>
              <input
                type="text"
                id="user_base_dn"
                value={formData.user_base_dn}
                onChange={handleInputChange('user_base_dn')}
                placeholder="ou=users,dc=example,dc=com"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Base DN for user search</p>
            </div>

            <div>
              <label htmlFor="user_filter" className="block text-sm font-medium mb-2">
                User Filter *
              </label>
              <input
                type="text"
                id="user_filter"
                value={formData.user_filter}
                onChange={handleInputChange('user_filter')}
                placeholder="(objectClass=person)"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">LDAP filter for user search</p>
            </div>

            <div>
              <label htmlFor="user_name_attr" className="block text-sm font-medium mb-2">
                Username Attribute *
              </label>
              <input
                type="text"
                id="user_name_attr"
                value={formData.user_name_attr}
                onChange={handleInputChange('user_name_attr')}
                placeholder="uid"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Attribute containing username</p>
            </div>

            <div>
              <label htmlFor="user_email_attr" className="block text-sm font-medium mb-2">
                Email Attribute *
              </label>
              <input
                type="text"
                id="user_email_attr"
                value={formData.user_email_attr}
                onChange={handleInputChange('user_email_attr')}
                placeholder="mail"
                disabled={loading}
                required
                className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
              />
              <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">Attribute containing email address</p>
            </div>
          </div>

          <div className="flex gap-3 mt-6">
            <button
              type="submit"
              disabled={loading}
              className="btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Saving...
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Save className="w-4 h-4" />
                  Save Configuration
                </span>
              )}
            </button>
            <button
              type="button"
              onClick={handleTestConnection}
              disabled={loading || !formData.url || !formData.bind_dn}
              className="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Testing...
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Network className="w-4 h-4" />
                  Test Connection
                </span>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default LDAPConfigTab;

