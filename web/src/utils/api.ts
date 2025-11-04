import axios, { AxiosInstance, AxiosResponse } from 'axios';

/**
 * Get the authorization token from localStorage
 */
export const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('accessToken') || localStorage.getItem('authToken');
};

/**
 * Create fetch options with authorization header
 */
export const createAuthHeaders = (): HeadersInit => {
  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  return headers;
};

/**
 * Fetch with automatic authorization header
 */
export const authFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const headers = {
    ...createAuthHeaders(),
    ...(options.headers || {}),
  };

  return fetch(url, {
    ...options,
    headers,
  });
};

// Types for API requests/responses
export interface LoginRequest {
  username_or_email?: string;
  username?: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  is_tmp_password?: boolean;
  requires_2fa?: boolean;
  session_id?: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  is_superuser: boolean;
  is_active: boolean;
  is_external: boolean;
  is_tmp_password?: boolean;
  two_fa_enabled: boolean;
  two_fa_confirmed_at: string | null;
  created_at: string;
  updated_at: string;
  last_login: string | null;
  license_accepted: boolean;
  project_permissions?: Record<string, string[]>;
}

export interface ProjectInfo {
  project: {
    id: number;
    name: string;
    description: string;
    created_at: string;
    updated_at: string;
    archived_at: string | null;
  };
  role: {
    id: string;
    key: string;
    name: string;
    description: string;
    created_at: string;
  } | null;
  permissions: string[];
}

export interface MyProjectsResponse {
  is_superuser: boolean;
  projects: ProjectInfo[];
}

export interface Verify2FARequest {
  code: string;
  session_id: string;
}

export interface Verify2FAResponse {
  access_token: string;
  refresh_token: string;
  is_tmp_password?: boolean;
}

export interface ChangeUserPasswordRequest {
  old_password: string;
  new_password: string;
}

export interface SSOCallbackRequest {
  provider: string;
  response: string;
  state: string;
}

export interface SSOCallbackResponse {
  access_token: string;
  refresh_token: string;
  is_tmp_password?: boolean;
}

export interface SSOProvider {
  name: string;
  display_name: string;
  icon_url: string;
  type: string;
}

export interface CreateTenantRequest {
  name: string;
}

export interface Tenant {
  ID?: number;
  Name?: string;
  CreatedAt?: string;
  id?: number;  // Fallback for lowercase
  name?: string;
  created_at?: string;
}

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    const token = getAuthToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// API client methods
const apiClient = {
  login: async (data: LoginRequest): Promise<AxiosResponse<LoginResponse>> => {
    const payload = data.username ? { username_or_email: data.username, password: data.password } : data;
    return api.post('/api/v1/auth/login', payload);
  },

  refreshToken: async (data: RefreshTokenRequest): Promise<AxiosResponse<RefreshTokenResponse>> => {
    return api.post('/api/v1/auth/refresh', data);
  },

  getCurrentUser: async (): Promise<AxiosResponse<User>> => {
    return api.get('/api/v1/users/me');
  },

  getMyProjects: async (): Promise<AxiosResponse<MyProjectsResponse>> => {
    return api.get('/api/v1/users/me/projects');
  },

  verify2FA: async (data: Verify2FARequest): Promise<AxiosResponse<Verify2FAResponse>> => {
    return api.post('/api/v1/auth/2fa/verify', data);
  },

  userChangeMyPassword: async (data: ChangeUserPasswordRequest): Promise<AxiosResponse<void>> => {
    return api.post('/api/v1/users/me/password', data);
  },

  sSOCallback: async (data: SSOCallbackRequest): Promise<AxiosResponse<SSOCallbackResponse>> => {
    return api.post('/api/v1/auth/sso/callback', data);
  },

  getSSOProviders: async (): Promise<AxiosResponse<{ providers: SSOProvider[] }>> => {
    return api.get('/api/v1/auth/sso/providers');
  },

  sSOInitiate: async (providerName: string): Promise<AxiosResponse<{ redirect_url: string }>> => {
    return api.post('/api/v1/auth/sso/initiate', { provider_name: providerName });
  },

  createTenant: async (data: CreateTenantRequest): Promise<AxiosResponse<Tenant>> => {
    return api.post('/api/v1/tenants', data);
  },

  updateTenant: async (id: number, data: { name: string }): Promise<AxiosResponse<Tenant>> => {
    return api.put(`/api/v1/tenants/${id}`, data);
  },

  deleteTenant: async (id: number): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/tenants/${id}`);
  },

  createProject: async (data: CreateProjectRequest): Promise<AxiosResponse<Project>> => {
    return api.post('/api/v1/projects', data);
  },

  updateProject: async (id: number, data: { name: string; description?: string }): Promise<AxiosResponse<Project>> => {
    return api.put(`/api/v1/projects/${id}`, data);
  },

  deleteProject: async (id: number): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/projects/${id}`);
  },

  createUser: async (data: CreateUserRequest): Promise<AxiosResponse<CreateUserResponse>> => {
    return api.post('/api/v1/users', data);
  },

  listUsers: async (): Promise<AxiosResponse<UserListItem[]>> => {
    return api.get('/api/v1/users');
  },

  updateUserStatus: async (id: number, data: { is_active?: boolean; is_superuser?: boolean }): Promise<AxiosResponse<UserListItem>> => {
    return api.put(`/api/v1/users/${id}/status`, data);
  },

  deleteUser: async (id: number): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/users/${id}`);
  },

  listProjectMemberships: async (projectId: number): Promise<AxiosResponse<Membership[]>> => {
    return api.get(`/api/v1/projects/${projectId}/memberships`);
  },

  createProjectMembership: async (projectId: number, data: { user_id: number; role_id: string }): Promise<AxiosResponse<Membership>> => {
    return api.post(`/api/v1/projects/${projectId}/memberships`, data);
  },

  deleteProjectMembership: async (projectId: number, membershipId: string): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/projects/${projectId}/memberships/${membershipId}`);
  },

  listRoles: async (): Promise<AxiosResponse<Role[]>> => {
    return api.get('/api/v1/roles');
  },

  // LDAP endpoints
  getLDAPConfig: async (): Promise<AxiosResponse<LDAPConfig>> => {
    return api.get('/api/v1/ldap/config');
  },

  updateLDAPConfig: async (config: LDAPConfig): Promise<AxiosResponse<LDAPConfigResponse>> => {
    return api.post('/api/v1/ldap/config', config);
  },

  deleteLDAPConfig: async (): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete('/api/v1/ldap/config');
  },

  testLDAPConnection: async (config: LDAPConnectionTest): Promise<AxiosResponse<LDAPConnectionTestResponse>> => {
    return api.post('/api/v1/ldap/test-connection', config);
  },

  syncLDAPUsers: async (): Promise<AxiosResponse<LDAPSyncStartResponse>> => {
    return api.post('/api/v1/ldap/sync/users');
  },

  cancelLDAPSync: async (): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete('/api/v1/ldap/sync/cancel');
  },

  getLDAPSyncStatus: async (): Promise<AxiosResponse<LDAPSyncStatus>> => {
    return api.get('/api/v1/ldap/sync/status');
  },

  getLDAPSyncProgress: async (): Promise<AxiosResponse<LDAPSyncProgress>> => {
    return api.get('/api/v1/ldap/sync/progress');
  },

  getLDAPSyncLogs: async (
    limit?: number,
    level?: string,
    syncId?: string,
    username?: string,
    from?: string,
    to?: string
  ): Promise<AxiosResponse<LDAPSyncLogsResult>> => {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    if (level) params.append('level', level);
    if (syncId) params.append('sync_id', syncId);
    if (username) params.append('username', username);
    if (from) params.append('from', from);
    if (to) params.append('to', to);
    return api.get(`/api/v1/ldap/sync/logs?${params.toString()}`);
  },

  getLDAPSyncLogDetails: async (id: number): Promise<AxiosResponse<LDAPSyncLog>> => {
    return api.get(`/api/v1/ldap/sync/logs/${id}`);
  },

  getLDAPStatistics: async (): Promise<AxiosResponse<LDAPStatistics>> => {
    return api.get('/api/v1/ldap/statistics');
  },

  // Workflow assignment endpoints
  listUnassignedWorkflows: async (
    page?: number,
    pageSize?: number
  ): Promise<AxiosResponse<WorkflowsResponse>> => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());
    const query = params.toString() ? `?${params.toString()}` : '';
    return api.get(`/api/v1/unassigned-workflows${query}`);
  },

  assignWorkflowsToProject: async (
    projectId: number,
    workflowIds?: string[]
  ): Promise<AxiosResponse<AssignWorkflowsResponse>> => {
    return api.post(`/api/v1/projects/${projectId}/workflows/assign`, {
      workflow_ids: workflowIds || [],
    });
  },
};

export interface AssignWorkflowsResponse {
  message: string;
  assigned_count: number;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  is_superuser: boolean;
}

export interface CreateUserResponse {
  id: number;
  username: string;
  email: string;
  is_superuser: boolean;
  is_active: boolean;
  is_external: boolean;
  is_tmp_password: boolean;
  created_at: string;
}

export interface UserListItem {
  id: number;
  username: string;
  email: string;
  is_superuser: boolean;
  is_active: boolean;
  is_external: boolean;
  two_fa_enabled: boolean;
  two_fa_confirmed_at?: string | null;
  created_at: string;
  updated_at: string;
  last_login?: string | null;
  license_accepted: boolean;
  is_tmp_password: boolean;
}

export interface Membership {
  id: string;
  project_id: number;
  user_id: number;
  username: string;
  email: string;
  role_id: string;
  role_key: string;
  role_name: string;
  created_at: string;
}

export interface Role {
  id: string;
  key: string;
  name: string;
  description: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  tenant_id: number;
}

export interface Project {
  ID?: number;
  Name?: string;
  Description?: string;
  CreatedAt?: string;
  UpdatedAt?: string;
  ArchivedAt?: string | null;
  id?: number;  // Fallback for lowercase
  name?: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
  archived_at?: string | null;
}

// LDAP Types
export interface LDAPConfig {
  enabled: boolean;
  url: string;
  bind_dn: string;
  bind_password: string;
  user_base_dn: string;
  user_filter: string;
  user_name_attr: string;
  user_email_attr: string;
  start_tls: boolean;
  insecure_tls: boolean;
  timeout: string;
  sync_interval: number;
}

export interface LDAPConfigResponse {
  message: string;
  config: LDAPConfig;
}

export interface LDAPConnectionTest {
  url: string;
  bind_dn: string;
  bind_password: string;
  user_base_dn?: string;
  user_filter?: string;
  user_name_attr?: string;
  start_tls?: boolean;
  insecure_tls?: boolean;
  timeout?: string;
}

export interface LDAPConnectionTestResponse {
  success: boolean;
  message: string;
  details?: {
    server_info?: string;
    user_count?: number;
    test_user?: string;
  };
}

export interface LDAPSyncStartResponse {
  message: string;
  sync_id: string;
  estimated_duration?: string;
}

export interface LDAPSyncStatus {
  status: string;
  is_running: boolean;
  last_sync_time?: string;
  total_users: number;
  synced_users: number;
  errors: number;
  warnings: number;
  last_sync_duration?: string;
}

export interface LDAPSyncProgress {
  is_running: boolean;
  progress: number;
  current_step?: string;
  processed_items: number;
  total_items: number;
  estimated_time?: string;
  start_time?: string;
  sync_id?: string;
}

export interface LDAPSyncLog {
  id: number;
  timestamp: string;
  level: string;
  message: string;
  username?: string | null;
  details?: string | null;
  sync_session_id: string;
  stack_trace?: string | null;
  ldap_error_code?: number | null;
  ldap_error_message?: string | null;
}

export interface LDAPSyncLogsResult {
  logs: LDAPSyncLog[];
  total: number;
  has_more: boolean;
}

export interface LDAPStatistics {
  ldap_users: number;
  local_users: number;
  active_users: number;
  inactive_users: number;
  sync_history: Array<{
    date: string;
    users_synced: number;
    errors: number;
    duration_minutes: number;
  }>;
  sync_success_rate: number;
}

export default apiClient;

