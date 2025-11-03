import React, { useState, useEffect } from 'react';
import { useAuth } from '../auth/AuthContext';
import { useRBAC } from '../auth/permissions';
import apiClient, { type UserListItem } from '../utils/api';
import { Navigate } from 'react-router-dom';
import LDAPConfigTab from '../components/ldap/LDAPConfigTab';
import LDAPSyncTab from '../components/ldap/LDAPSyncTab';
import LDAPLogsTab from '../components/ldap/LDAPLogsTab';

export const Admin: React.FC = () => {
  const { user } = useAuth();
  const rbac = useRBAC();
  const [activeTab, setActiveTab] = useState<'users' | 'ldap'>('users');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [users, setUsers] = useState<UserListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [updatingUserId, setUpdatingUserId] = useState<number | null>(null);
  const [deletingUserId, setDeletingUserId] = useState<number | null>(null);

  // Form state for creating users
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isSuperuserCheckbox, setIsSuperuserCheckbox] = useState(false);
  const [creating, setCreating] = useState(false);

  const isSuperuser = user?.is_superuser || rbac.isSuperuser;

  // Redirect if not superuser
  if (!isSuperuser) {
    return <Navigate to="/tenants" replace />;
  }

  useEffect(() => {
    if (activeTab === 'users') {
      fetchUsers();
    }
  }, [activeTab]);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.listUsers();
      setUsers(response.data);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to fetch users';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!username.trim() || !email.trim() || !password.trim()) {
      setError('All fields are required');
      return;
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setCreating(true);

    try {
      await apiClient.createUser({
        username: username.trim(),
        email: email.trim(),
        password: password,
        is_superuser: isSuperuserCheckbox,
      });

      setSuccess('User created successfully');
      setTimeout(() => setSuccess(null), 3000);
      setUsername('');
      setEmail('');
      setPassword('');
      setIsSuperuserCheckbox(false);
      setShowCreateModal(false);
      await fetchUsers();
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to create user';
      setError(errorMessage);
    } finally {
      setCreating(false);
    }
  };

  const handleToggleActive = async (userId: number, currentActive: boolean) => {
    setUpdatingUserId(userId);
    setError(null);

    try {
      const response = await apiClient.updateUserStatus(userId, { is_active: !currentActive });
      setUsers(users.map(u => u.id === userId ? response.data : u));
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to update user status';
      setError(errorMessage);
    } finally {
      setUpdatingUserId(null);
    }
  };

  const handleToggleSuperuser = async (userId: number, currentSuperuser: boolean) => {
    setUpdatingUserId(userId);
    setError(null);

    try {
      const response = await apiClient.updateUserStatus(userId, { is_superuser: !currentSuperuser });
      setUsers(users.map(u => u.id === userId ? response.data : u));
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to update user status';
      setError(errorMessage);
    } finally {
      setUpdatingUserId(null);
    }
  };

  const handleDeleteUser = async (userId: number, username: string) => {
    if (!window.confirm(`Are you sure you want to delete user "${username}"? This action cannot be undone.`)) {
      return;
    }

    setDeletingUserId(userId);
    setError(null);

    try {
      await apiClient.deleteUser(userId);
      setUsers(users.filter(u => u.id !== userId));
      setSuccess('User deleted successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to delete user';
      setError(errorMessage);
    } finally {
      setDeletingUserId(null);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Admin Panel</h1>

      {/* Tabs */}
      <div className="mb-6 border-b border-slate-200 dark:border-[#3e3e42]">
        <div className="flex gap-4">
          <button
            onClick={() => setActiveTab('users')}
            className={`px-4 py-2 font-medium transition-colors ${
              activeTab === 'users'
                ? 'text-slate-900 dark:text-[#ff6b35] border-b-2 border-slate-900 dark:border-[#ff6b35]'
                : 'text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
            }`}
          >
            Users
          </button>
          <button
            onClick={() => setActiveTab('ldap')}
            className={`px-4 py-2 font-medium transition-colors ${
              activeTab === 'ldap'
                ? 'text-slate-900 dark:text-[#ff6b35] border-b-2 border-slate-900 dark:border-[#ff6b35]'
                : 'text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
            }`}
          >
            LDAP
          </button>
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'users' && (
        <div>
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold">Users</h2>
            <button
              onClick={() => {
                setShowCreateModal(true);
                setError(null);
                setSuccess(null);
              }}
              className="btn btn-primary"
            >
              <svg className="w-5 h-5 inline mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create User
            </button>
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

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
            </div>
          ) : (
            <div className="card">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200 dark:border-[#3e3e42]">
                      <th className="text-left px-4 py-3 text-sm font-semibold">User</th>
                      <th className="text-left px-4 py-3 text-sm font-semibold">Type</th>
                      <th className="text-left px-4 py-3 text-sm font-semibold">Status</th>
                      <th className="text-left px-4 py-3 text-sm font-semibold">2FA</th>
                      <th className="text-left px-4 py-3 text-sm font-semibold">Last Login</th>
                      <th className="text-left px-4 py-3 text-sm font-semibold">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((userItem) => (
                      <tr key={userItem.id} className="border-b border-slate-100 dark:border-[#3e3e42]/50 hover:bg-slate-50 dark:hover:bg-[#2d2d30]">
                        <td className="px-4 py-3">
                          <div>
                            <div className="font-medium text-slate-900 dark:text-[#ff6b35]">{userItem.username}</div>
                            <div className="text-sm text-slate-600 dark:text-[#ff4500]">{userItem.email}</div>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                              userItem.is_external
                                ? 'bg-blue-100 text-blue-800 dark:bg-blue-950/30 dark:text-blue-400'
                                : 'bg-green-100 text-green-800 dark:bg-green-950/30 dark:text-green-400'
                            }`}
                          >
                            {userItem.is_external ? 'External' : 'Internal'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex flex-col gap-2">
                            <label className="flex items-center gap-2 cursor-pointer">
                              <input
                                type="checkbox"
                                checked={userItem.is_active}
                                onChange={() => handleToggleActive(userItem.id, userItem.is_active)}
                                disabled={updatingUserId === userItem.id}
                                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
                              />
                              <span className="text-sm">
                                {userItem.is_active ? (
                                  <span className="text-green-600 dark:text-green-400">Active</span>
                                ) : (
                                  <span className="text-red-600 dark:text-red-400">Inactive</span>
                                )}
                              </span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                              <input
                                type="checkbox"
                                checked={userItem.is_superuser}
                                onChange={() => handleToggleSuperuser(userItem.id, userItem.is_superuser)}
                                disabled={updatingUserId === userItem.id}
                                className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
                              />
                              <span className="text-sm">
                                {userItem.is_superuser ? (
                                  <span className="text-purple-600 dark:text-purple-400">Superuser</span>
                                ) : (
                                  <span className="text-slate-600 dark:text-[#ff4500]">Regular</span>
                                )}
                              </span>
                            </label>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {userItem.two_fa_enabled ? (
                            <span className="text-sm text-green-600 dark:text-green-400">Enabled</span>
                          ) : (
                            <span className="text-sm text-slate-400">Disabled</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <div className="text-sm text-slate-600 dark:text-[#ff4500]">
                            {userItem.last_login ? (
                              <div>
                                <div>{new Date(userItem.last_login).toLocaleDateString()}</div>
                                <div className="text-xs">{new Date(userItem.last_login).toLocaleTimeString()}</div>
                              </div>
                            ) : (
                              <span className="text-slate-400">Never</span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => handleDeleteUser(userItem.id, userItem.username)}
                            disabled={deletingUserId === userItem.id}
                            className="p-2 rounded-lg bg-red-50 dark:bg-red-950/30 hover:bg-red-100 dark:hover:bg-red-950/50 text-red-600 dark:text-red-400 disabled:opacity-50"
                            title="Delete user"
                          >
                            {deletingUserId === userItem.id ? (
                              <div className="w-4 h-4 border-2 border-red-600 dark:border-red-400 border-t-transparent rounded-full animate-spin"></div>
                            ) : (
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                            )}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {users.length === 0 && !loading && (
                <div className="text-center py-12">
                  <p className="text-slate-500 dark:text-[#ff4500]">No users found</p>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {activeTab === 'ldap' && <LDAPTabContent />}

      {/* Create User Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-md w-full bg-white dark:bg-[#1e1e1e]">
            <h2 className="text-xl font-semibold mb-4">Create Internal User</h2>
            <form onSubmit={handleCreateUser}>
              <div className="mb-4">
                <label htmlFor="username" className="block text-sm font-medium mb-2">
                  Username *
                </label>
                <input
                  type="text"
                  id="username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter username"
                  autoFocus
                  disabled={creating}
                  required
                />
              </div>

              <div className="mb-4">
                <label htmlFor="email" className="block text-sm font-medium mb-2">
                  Email *
                </label>
                <input
                  type="email"
                  id="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter email"
                  disabled={creating}
                  required
                />
              </div>

              <div className="mb-4">
                <label htmlFor="password" className="block text-sm font-medium mb-2">
                  Password *
                </label>
                <input
                  type="password"
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter password (min 6 characters)"
                  disabled={creating}
                  required
                  minLength={6}
                />
              </div>

              <div className="mb-6">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={isSuperuserCheckbox}
                    onChange={(e) => setIsSuperuserCheckbox(e.target.checked)}
                    disabled={creating}
                    className="w-4 h-4 rounded border-slate-300 dark:border-[#3e3e42] text-slate-600 focus:ring-slate-500"
                  />
                  <span className="text-sm font-medium">Superuser</span>
                </label>
                <p className="text-xs text-slate-500 dark:text-[#ff4500] mt-1">
                  Grant superuser privileges to this user
                </p>
              </div>

              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setUsername('');
                    setEmail('');
                    setPassword('');
                    setIsSuperuserCheckbox(false);
                    setError(null);
                    setSuccess(null);
                  }}
                  disabled={creating}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating || !username.trim() || !email.trim() || !password.trim()}
                  className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {creating ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

// LDAP Tab Content Component
const LDAPTabContent: React.FC = () => {
  const [ldapSubTab, setLdapSubTab] = useState<'config' | 'sync' | 'logs'>('config');

  return (
    <div>
      {/* LDAP Sub-tabs */}
      <div className="mb-6 border-b border-slate-200 dark:border-[#3e3e42]">
        <div className="flex gap-4">
          <button
            onClick={() => setLdapSubTab('config')}
            className={`px-4 py-2 font-medium transition-colors ${
              ldapSubTab === 'config'
                ? 'text-slate-900 dark:text-[#ff6b35] border-b-2 border-slate-900 dark:border-[#ff6b35]'
                : 'text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
            }`}
          >
            Configuration
          </button>
          <button
            onClick={() => setLdapSubTab('sync')}
            className={`px-4 py-2 font-medium transition-colors ${
              ldapSubTab === 'sync'
                ? 'text-slate-900 dark:text-[#ff6b35] border-b-2 border-slate-900 dark:border-[#ff6b35]'
                : 'text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
            }`}
          >
            Synchronization
          </button>
          <button
            onClick={() => setLdapSubTab('logs')}
            className={`px-4 py-2 font-medium transition-colors ${
              ldapSubTab === 'logs'
                ? 'text-slate-900 dark:text-[#ff6b35] border-b-2 border-slate-900 dark:border-[#ff6b35]'
                : 'text-slate-600 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
            }`}
          >
            Logs
          </button>
        </div>
      </div>

      {/* LDAP Sub-tab Content */}
      {ldapSubTab === 'config' && <LDAPConfigTab />}
      {ldapSubTab === 'sync' && <LDAPSyncTab />}
      {ldapSubTab === 'logs' && <LDAPLogsTab />}
    </div>
  );
};
