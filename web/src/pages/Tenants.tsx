import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authFetch } from '../utils/api';
import apiClient from '../utils/api';
import { useAuth } from '../auth/AuthContext';
import { useRBAC } from '../auth/permissions';

interface Tenant {
  ID: number;
  Name: string;
  CreatedAt: string;
}

export const Tenants: React.FC = () => {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newTenantName, setNewTenantName] = useState('');
  const [creating, setCreating] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState('');
  const [updating, setUpdating] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const navigate = useNavigate();
  const { user } = useAuth();
  const rbac = useRBAC();
  
  const isSuperuser = user?.is_superuser || rbac.isSuperuser;

  useEffect(() => {
    const fetchTenants = async () => {
      try {
        const response = await authFetch('/api/v1/tenants');
        if (!response.ok) {
          throw new Error('Failed to fetch tenants');
        }
        const data = await response.json();
        setTenants(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchTenants();
  }, []);

  const handleSelectTenant = (tenantId: number) => {
    navigate(`/tenants/${tenantId}/projects`);
  };

  const handleCreateTenant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTenantName.trim()) {
      setError('Tenant name is required');
      return;
    }

    setCreating(true);
    setError(null);

    try {
      const response = await apiClient.createTenant({ name: newTenantName.trim() });
      // Backend returns domain.Tenant which serializes as ID, Name, CreatedAt (capitalized)
      setTenants([...tenants, {
        ID: response.data.ID || response.data.id,
        Name: response.data.Name || response.data.name,
        CreatedAt: response.data.CreatedAt || response.data.created_at,
      }]);
      setNewTenantName('');
      setShowCreateModal(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to create tenant';
      setError(errorMessage);
    } finally {
      setCreating(false);
    }
  };

  const handleUpdateTenant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingName.trim() || !editingId) {
      setError('Tenant name is required');
      return;
    }

    setUpdating(true);
    setError(null);

    try {
      const response = await apiClient.updateTenant(editingId, { name: editingName.trim() });
      setTenants(tenants.map(t => 
        t.ID === editingId 
          ? {
              ID: response.data.ID || response.data.id || editingId,
              Name: response.data.Name || response.data.name || editingName.trim(),
              CreatedAt: response.data.CreatedAt || response.data.created_at || t.CreatedAt,
            }
          : t
      ));
      setEditingId(null);
      setEditingName('');
      setUpdating(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to update tenant';
      setError(errorMessage);
      setUpdating(false);
    }
  };

  const handleDeleteTenant = async (tenantId: number, e: React.MouseEvent) => {
    e.stopPropagation();
    
    if (!window.confirm(`Are you sure you want to delete tenant "${tenants.find(t => t.ID === tenantId)?.Name}"? This action cannot be undone.`)) {
      return;
    }

    setDeletingId(tenantId);
    setError(null);

    try {
      await apiClient.deleteTenant(tenantId);
      setTenants(tenants.filter(t => t.ID !== tenantId));
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to delete tenant';
      setError(errorMessage);
    } finally {
      setDeletingId(null);
    }
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading tenants...</span>
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
        <h1>Select Tenant</h1>
        {isSuperuser && (
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn btn-primary"
          >
            <svg className="w-5 h-5 inline mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Tenant
          </button>
        )}
      </div>

      {error && (
        <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
        </div>
      )}
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
        {tenants.map((tenant) => (
          <div
            key={tenant.ID}
            className="card transition-all duration-200 hover:shadow-lg hover:scale-105 relative"
            style={{
              background: 'rgba(255, 255, 255, 0.7)',
              backdropFilter: 'blur(12px) saturate(150%)',
              WebkitBackdropFilter: 'blur(12px) saturate(150%)',
              border: '1px solid rgba(255, 255, 255, 0.5)',
            }}
          >
            {isSuperuser && (
              <div className="absolute top-2 right-2 flex gap-2">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setEditingId(tenant.ID);
                    setEditingName(tenant.Name);
                    setError(null);
                  }}
                  className="p-1.5 rounded-lg bg-blue-50 dark:bg-blue-950/30 hover:bg-blue-100 dark:hover:bg-blue-950/50 text-blue-600 dark:text-blue-400"
                  title="Edit tenant"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  onClick={(e) => handleDeleteTenant(tenant.ID, e)}
                  disabled={deletingId === tenant.ID}
                  className="p-1.5 rounded-lg bg-red-50 dark:bg-red-950/30 hover:bg-red-100 dark:hover:bg-red-950/50 text-red-600 dark:text-red-400 disabled:opacity-50"
                  title="Delete tenant"
                >
                  {deletingId === tenant.ID ? (
                    <div className="w-4 h-4 border-2 border-red-600 dark:border-red-400 border-t-transparent rounded-full animate-spin"></div>
                  ) : (
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  )}
                </button>
              </div>
            )}
            <div
              onClick={() => handleSelectTenant(tenant.ID)}
              className="flex flex-col items-center text-center p-6 cursor-pointer"
            >
              <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-4">
                <span className="text-2xl font-bold text-white">
                  {tenant.Name?.charAt(0).toUpperCase() || 'T'}
                </span>
              </div>
              <h2 className="text-xl font-semibold mb-2">{tenant.Name}</h2>
              <p className="text-sm text-slate-600 dark:text-[#ff4500]500">
                Created {new Date(tenant.CreatedAt).toLocaleDateString()}
              </p>
            </div>
          </div>
        ))}
      </div>

      {tenants.length === 0 && (
        <div className="text-center py-12">
          <p className="text-slate-500 dark:text-[#ff4500]500">No tenants found</p>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-md w-full bg-white dark:bg-[#1e1e1e]">
            <h2 className="text-xl font-semibold mb-4">Create New Tenant</h2>
            <form onSubmit={handleCreateTenant}>
              <div className="mb-4">
                <label htmlFor="tenantName" className="block text-sm font-medium mb-2">
                  Tenant Name
                </label>
                <input
                  type="text"
                  id="tenantName"
                  value={newTenantName}
                  onChange={(e) => setNewTenantName(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter tenant name"
                  autoFocus
                  disabled={creating}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setNewTenantName('');
                    setError(null);
                  }}
                  disabled={creating}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating || !newTenantName.trim()}
                  className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {creating ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingId !== null && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-md w-full bg-white dark:bg-[#1e1e1e]">
            <h2 className="text-xl font-semibold mb-4">Edit Tenant</h2>
            <form onSubmit={handleUpdateTenant}>
              <div className="mb-4">
                <label htmlFor="editingTenantName" className="block text-sm font-medium mb-2">
                  Tenant Name
                </label>
                <input
                  type="text"
                  id="editingTenantName"
                  value={editingName}
                  onChange={(e) => setEditingName(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter tenant name"
                  autoFocus
                  disabled={updating}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setEditingId(null);
                    setEditingName('');
                    setError(null);
                  }}
                  disabled={updating}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={updating || !editingName.trim()}
                  className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {updating ? 'Updating...' : 'Update'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

