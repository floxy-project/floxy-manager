import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { authFetch } from '../utils/api';
import apiClient, { type CreateProjectRequest } from '../utils/api';
import { useRBAC } from '../auth/permissions';
import { useAuth } from '../auth/AuthContext';
import { AlertCircle, Plus, Edit, Trash2, Briefcase, Loader2 } from 'lucide-react';

interface Project {
  ID: number;
  Name: string;
  Description: string;
  CreatedAt: string;
  UpdatedAt: string;
}

export const Projects: React.FC = () => {
  const { tenantId } = useParams<{ tenantId: string }>();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');
  const [newProjectDescription, setNewProjectDescription] = useState('');
  const [creating, setCreating] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState('');
  const [editingDescription, setEditingDescription] = useState('');
  const [updating, setUpdating] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const navigate = useNavigate();
  const rbac = useRBAC();
  const { user } = useAuth();
  
  const isSuperuser = user?.is_superuser || rbac.isSuperuser;
  const canCreate = isSuperuser || rbac.canCreateProject();

  useEffect(() => {
    if (tenantId) {
      fetchProjects();
    }
  }, [tenantId]);

  const fetchProjects = async () => {
    try {
      setLoading(true);
      const response = await authFetch(`/api/v1/projects?tenant_id=${tenantId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch projects');
      }
      const data = await response.json();
      // Projects are already filtered by tenant_id on backend
      setProjects(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectProject = (projectId: number) => {
    navigate(`/tenants/${tenantId}/projects/${projectId}/dashboard`);
  };

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newProjectName.trim() || !tenantId) {
      setError('Project name and tenant ID are required');
      return;
    }

    setCreating(true);
    setError(null);

    try {
      const request: CreateProjectRequest = {
        name: newProjectName.trim(),
        description: newProjectDescription.trim() || undefined,
        tenant_id: parseInt(tenantId, 10),
      };
      const response = await apiClient.createProject(request);
      
      // Backend returns domain.Project which serializes as ID, Name, etc.
      const newProject: Project = {
        ID: response.data.ID || response.data.id,
        Name: response.data.Name || response.data.name,
        Description: response.data.Description || response.data.description || '',
        CreatedAt: response.data.CreatedAt || response.data.created_at || '',
        UpdatedAt: response.data.UpdatedAt || response.data.updated_at || '',
      };
      
      setProjects([...projects, newProject]);
      setNewProjectName('');
      setNewProjectDescription('');
      setShowCreateModal(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to create project';
      setError(errorMessage);
    } finally {
      setCreating(false);
    }
  };

  const handleUpdateProject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingName.trim() || !editingId) {
      setError('Project name is required');
      return;
    }

    setUpdating(true);
    setError(null);

    try {
      const response = await apiClient.updateProject(editingId, {
        name: editingName.trim(),
        description: editingDescription.trim() || undefined,
      });
      
      setProjects(projects.map(p => 
        p.ID === editingId 
          ? {
              ID: response.data.ID || response.data.id || editingId,
              Name: response.data.Name || response.data.name || editingName.trim(),
              Description: response.data.Description || response.data.description || editingDescription.trim(),
              CreatedAt: response.data.CreatedAt || response.data.created_at || p.CreatedAt,
              UpdatedAt: response.data.UpdatedAt || response.data.updated_at || new Date().toISOString(),
            }
          : p
      ));
      setEditingId(null);
      setEditingName('');
      setEditingDescription('');
      setUpdating(false);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to update project';
      setError(errorMessage);
      setUpdating(false);
    }
  };

  const handleDeleteProject = async (projectId: number, e: React.MouseEvent) => {
    e.stopPropagation();
    
    const project = projects.find(p => p.ID === projectId);
    if (!window.confirm(`Are you sure you want to delete project "${project?.Name}"? This action cannot be undone.`)) {
      return;
    }

    setDeletingId(projectId);
    setError(null);

    try {
      await apiClient.deleteProject(projectId);
      setProjects(projects.filter(p => p.ID !== projectId));
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || 'Failed to delete project';
      setError(errorMessage);
    } finally {
      setDeletingId(null);
    }
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <Loader2 className="w-5 h-5 animate-spin" />
          <span>Loading projects...</span>
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

  return (
    <div>
      <div className="mb-6">
        <button
          onClick={() => navigate('/tenants')}
          className="btn btn-nav"
        >
          ← Back to Tenants
        </button>
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1>Select Project</h1>
        {canCreate && (
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn btn-primary"
          >
            <Plus className="w-5 h-5 inline mr-2" />
            Create Project
          </button>
        )}
      </div>

      {error && (
        <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
          <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
        </div>
      )}
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
        {projects.map((project) => {
          const canManage = isSuperuser || rbac.canManageProject(project.ID);
          const canDelete = isSuperuser || rbac.canDeleteProject(project.ID);
          return (
            <div
              key={project.ID}
              className="card transition-all duration-200 hover:shadow-lg hover:scale-105 relative"
              style={{
                background: 'rgba(255, 255, 255, 0.7)',
                backdropFilter: 'blur(12px) saturate(150%)',
                WebkitBackdropFilter: 'blur(12px) saturate(150%)',
                border: '1px solid rgba(255, 255, 255, 0.5)',
              }}
            >
              {canManage && (
                <div className="absolute top-2 right-2 flex gap-2">
                  <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditingId(project.ID);
                        setEditingName(project.Name);
                        setEditingDescription(project.Description || '');
                        setError(null);
                      }}
                      className="p-1.5 rounded-lg bg-blue-50 dark:bg-blue-950/30 hover:bg-blue-100 dark:hover:bg-blue-950/50 text-blue-600 dark:text-blue-400"
                      title="Edit project"
                    >
                      <Edit className="w-4 h-4" />
                    </button>
                  {canDelete && (
                    <button
                      onClick={(e) => handleDeleteProject(project.ID, e)}
                      disabled={deletingId === project.ID}
                      className="p-1.5 rounded-lg bg-red-50 dark:bg-red-950/30 hover:bg-red-100 dark:hover:bg-red-950/50 text-red-600 dark:text-red-400 disabled:opacity-50"
                      title="Delete project"
                    >
                      {deletingId === project.ID ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <Trash2 className="w-4 h-4" />
                      )}
                    </button>
                  )}
                </div>
              )}
              <div
                onClick={() => handleSelectProject(project.ID)}
                className="p-6 cursor-pointer"
              >
                <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-green-500 to-teal-600 flex items-center justify-center mb-4">
                  <Briefcase className="w-6 h-6 text-white" />
                </div>
                <h2 className="text-xl font-semibold mb-2">{project.Name}</h2>
                {project.Description && (
                  <p className="text-sm text-slate-600 dark:text-[#ff4500]500 mb-3 line-clamp-2">
                    {project.Description}
                  </p>
                )}
                <p className="text-xs text-slate-400">
                  Updated {new Date(project.UpdatedAt).toLocaleDateString()}
                </p>
              </div>
            </div>
          );
        })}
      </div>

      {projects.length === 0 && (
        <div className="text-center py-12">
          <p className="text-slate-500 dark:text-[#ff4500]500">No projects found</p>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="card max-w-md w-full bg-white dark:bg-[#1e1e1e]">
            <h2 className="text-xl font-semibold mb-4">Create New Project</h2>
            <form onSubmit={handleCreateProject}>
              <div className="mb-4">
                <label htmlFor="projectName" className="block text-sm font-medium mb-2">
                  Project Name *
                </label>
                <input
                  type="text"
                  id="projectName"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter project name"
                  autoFocus
                  disabled={creating}
                  required
                />
              </div>
              <div className="mb-4">
                <label htmlFor="projectDescription" className="block text-sm font-medium mb-2">
                  Description (optional)
                </label>
                <textarea
                  id="projectDescription"
                  value={newProjectDescription}
                  onChange={(e) => setNewProjectDescription(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter project description"
                  rows={3}
                  disabled={creating}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setNewProjectName('');
                    setNewProjectDescription('');
                    setError(null);
                  }}
                  disabled={creating}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating || !newProjectName.trim()}
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
            <h2 className="text-xl font-semibold mb-4">Edit Project</h2>
            <form onSubmit={handleUpdateProject}>
              <div className="mb-4">
                <label htmlFor="editingProjectName" className="block text-sm font-medium mb-2">
                  Project Name *
                </label>
                <input
                  type="text"
                  id="editingProjectName"
                  value={editingName}
                  onChange={(e) => setEditingName(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter project name"
                  autoFocus
                  disabled={updating}
                  required
                />
              </div>
              <div className="mb-4">
                <label htmlFor="editingProjectDescription" className="block text-sm font-medium mb-2">
                  Description (optional)
                </label>
                <textarea
                  id="editingProjectDescription"
                  value={editingDescription}
                  onChange={(e) => setEditingDescription(e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  placeholder="Enter project description"
                  rows={3}
                  disabled={updating}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setEditingId(null);
                    setEditingName('');
                    setEditingDescription('');
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

