import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { useRBAC } from '../auth/permissions';
import apiClient, { type Membership, type Role, type UserListItem } from '../utils/api';
import { Navigate } from 'react-router-dom';
import { Plus, Trash2, Loader2 } from 'lucide-react';

export const Memberships: React.FC = () => {
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const [memberships, setMemberships] = useState<Membership[]>([]);
  const [allUsers, setAllUsers] = useState<UserListItem[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number | ''>('');
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [adding, setAdding] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const { user } = useAuth();
  const rbac = useRBAC(projectId ? Number(projectId) : undefined);

  // Can view memberships page with project.view permission
  const canView = user?.is_superuser || rbac.canViewProject();
  // Can manage memberships (add/remove) with membership.manage permission
  const canManage = user?.is_superuser || rbac.canManageMembership();

  useEffect(() => {
    if (projectId) {
      fetchData();
    }
  }, [projectId]);

  const fetchData = async () => {
    if (!projectId) return;

    try {
      setLoading(true);
      setError(null);

      const [membershipsRes, usersRes, rolesRes] = await Promise.all([
        apiClient.listProjectMemberships(Number(projectId)),
        apiClient.listUsers(),
        apiClient.listRoles(),
      ]);

      setMemberships(membershipsRes.data);
      setAllUsers(usersRes.data);
      setRoles(rolesRes.data);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to fetch data';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleAddMembership = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!projectId || !selectedUserId || !selectedRoleId) {
      setError('Please select a user and role');
      return;
    }

    setAdding(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await apiClient.createProjectMembership(Number(projectId), {
        user_id: Number(selectedUserId),
        role_id: selectedRoleId,
      });

      setMemberships([...memberships, response.data]);
      setSuccess('User added to project successfully');
      setTimeout(() => setSuccess(null), 3000);
      setSelectedUserId('');
      setSelectedRoleId('');
      setShowAddModal(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to add user to project';
      setError(errorMessage);
    } finally {
      setAdding(false);
    }
  };

  const handleDeleteMembership = async (membershipId: string, username: string) => {
    if (!projectId) return;

    if (!window.confirm(`Are you sure you want to remove "${username}" from this project?`)) {
      return;
    }

    setDeletingId(membershipId);
    setError(null);

    try {
      await apiClient.deleteProjectMembership(Number(projectId), membershipId);
      setMemberships(memberships.filter(m => m.id !== membershipId));
      setSuccess('User removed from project successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to remove user from project';
      setError(errorMessage);
    } finally {
      setDeletingId(null);
    }
  };

  // Filter out users who are already members
  const availableUsers = allUsers.filter(
    user => !memberships.some(m => m.user_id === user.id)
  );

  if (!canView) {
    return <Navigate to={`/tenants/${tenantId}/projects/${projectId}/dashboard`} replace />;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Project Members</h1>
        {canManage && (
          <button
            onClick={() => {
              setShowAddModal(true);
              setError(null);
              setSuccess(null);
            }}
            className="btn btn-primary"
            disabled={availableUsers.length === 0}
          >
            <Plus className="w-5 h-5 inline mr-2" />
            Add Member
          </button>
        )}
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
          <Loader2 className="w-8 h-8 animate-spin" />
        </div>
      ) : (
        <div className="card">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-slate-200 dark:border-[#3e3e42]">
                  <th className="text-left px-4 py-3 text-sm font-semibold">User</th>
                  <th className="text-left px-4 py-3 text-sm font-semibold">Role</th>
                  <th className="text-left px-4 py-3 text-sm font-semibold">Added</th>
                  {canManage && <th className="text-left px-4 py-3 text-sm font-semibold">Actions</th>}
                </tr>
              </thead>
              <tbody>
                {memberships.map((membership) => (
                  <tr key={membership.id} className="border-b border-slate-100 dark:border-[#3e3e42]/50 hover:bg-slate-50 dark:hover:bg-[#2d2d30]">
                    <td className="px-4 py-3">
                      <div>
                        <div className="font-medium text-slate-900 dark:text-[#ff6b35]">{membership.username}</div>
                        <div className="text-sm text-slate-600 dark:text-[#ff4500]">{membership.email}</div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-950/30 dark:text-blue-400">
                        {membership.role_name}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-600 dark:text-[#ff4500]">
                      {new Date(membership.created_at).toLocaleDateString()}
                    </td>
                    {canManage && (
                      <td className="px-4 py-3">
                        <button
                          onClick={() => handleDeleteMembership(membership.id, membership.username)}
                          disabled={deletingId === membership.id}
                          className="p-2 rounded-lg bg-red-50 dark:bg-red-950/30 hover:bg-red-100 dark:hover:bg-red-950/50 text-red-600 dark:text-red-400 disabled:opacity-50"
                          title="Remove member"
                        >
                          {deletingId === membership.id ? (
                            <Loader2 className="w-4 h-4 animate-spin" />
                          ) : (
                            <Trash2 className="w-4 h-4" />
                          )}
                        </button>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {memberships.length === 0 && !loading && (
            <div className="text-center py-12">
              <p className="text-slate-500 dark:text-[#ff4500]">No members found</p>
            </div>
          )}
        </div>
      )}

      {/* Add Member Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-md w-full bg-white dark:bg-[#1e1e1e]">
            <h2 className="text-xl font-semibold mb-4">Add Member to Project</h2>
            <form onSubmit={handleAddMembership}>
              <div className="mb-4">
                <label htmlFor="userId" className="block text-sm font-medium mb-2">
                  User *
                </label>
                <select
                  id="userId"
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value ? Number(e.target.value) : '')}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  disabled={adding}
                  required
                >
                  <option value="">Select a user</option>
                  {availableUsers.map((user) => (
                    <option key={user.id} value={user.id}>
                      {user.username} {user.email && `(${user.email})`}
                    </option>
                  ))}
                </select>
              </div>

              <div className="mb-6">
                <label htmlFor="roleId" className="block text-sm font-medium mb-2">
                  Role *
                </label>
                <select
                  id="roleId"
                  value={selectedRoleId}
                  onChange={(e) => setSelectedRoleId(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  disabled={adding}
                  required
                >
                  <option value="">Select a role</option>
                  {roles.map((role) => (
                    <option key={role.id} value={role.id}>
                      {role.name} {role.description && `- ${role.description}`}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false);
                    setSelectedUserId('');
                    setSelectedRoleId('');
                    setError(null);
                    setSuccess(null);
                  }}
                  disabled={adding}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={adding || !selectedUserId || !selectedRoleId}
                  className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {adding ? 'Adding...' : 'Add'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

