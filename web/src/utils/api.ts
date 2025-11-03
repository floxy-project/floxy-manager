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

  createTenant: async (data: CreateTenantRequest): Promise<AxiosResponse<Tenant>> => {
    return api.post('/api/v1/tenants', data);
  },

  deleteTenant: async (id: number): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/tenants/${id}`);
  },

  createProject: async (data: CreateProjectRequest): Promise<AxiosResponse<Project>> => {
    return api.post('/api/v1/projects', data);
  },

  deleteProject: async (id: number): Promise<AxiosResponse<{ message: string }>> => {
    return api.delete(`/api/v1/projects/${id}`);
  },
};

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

export default apiClient;

