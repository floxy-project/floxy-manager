import React, { useState, useEffect } from 'react';
import { authFetch } from '../utils/api';
import { useAuth } from '../auth/AuthContext';

interface UserInfo {
  id: number;
  username: string;
  email: string;
  is_superuser: boolean;
  is_active: boolean;
  is_external: boolean;
  two_fa_enabled: boolean;
  two_fa_confirmed_at: string | null;
  created_at: string;
  updated_at: string;
  last_login: string | null;
  license_accepted: boolean;
}

interface ProjectInfo {
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
  } | null;
  permissions: string[];
}

interface ProjectsResponse {
  is_superuser: boolean;
  projects: ProjectInfo[];
}

interface ChangePasswordData {
  old_password: string;
  new_password: string;
  confirm_password: string;
}

interface Setup2FAResponse {
  secret: string;
  qr_url: string;
  qr_image: string;
}

export const Account: React.FC = () => {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<'profile' | '2fa' | 'projects'>('profile');
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [projects, setProjects] = useState<ProjectInfo[]>([]);
  const [isSuperuser, setIsSuperuser] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Password change form
  const [passwordData, setPasswordData] = useState<ChangePasswordData>({
    old_password: '',
    new_password: '',
    confirm_password: '',
  });
  const [passwordErrors, setPasswordErrors] = useState<Partial<ChangePasswordData>>({});
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null);
  const [changingPassword, setChangingPassword] = useState(false);

  // 2FA state
  const [twoFASetup, setTwoFASetup] = useState<Setup2FAResponse | null>(null);
  const [twoFACode, setTwoFACode] = useState('');
  const [confirming2FA, setConfirming2FA] = useState(false);
  const [disabling2FA, setDisabling2FA] = useState(false);
  const [twoFAEmailCode, setTwoFAEmailCode] = useState('');

  useEffect(() => {
    fetchUserData();
  }, []);

  const fetchUserData = async () => {
    try {
      setLoading(true);
      const [userRes, projectsRes] = await Promise.all([
        authFetch('/api/v1/users/me'),
        authFetch('/api/v1/users/me/projects'),
      ]);

      if (!userRes.ok) {
        throw new Error('Failed to fetch user information');
      }

      const userData: UserInfo = await userRes.json();
      setUserInfo(userData);

      if (projectsRes.ok) {
        const projectsData: ProjectsResponse = await projectsRes.json();
        console.log('Projects data received:', projectsData);
        setProjects(projectsData.projects);
        setIsSuperuser(projectsData.is_superuser);
      } else {
        const errorText = await projectsRes.text();
        console.error('Failed to fetch projects:', projectsRes.status, errorText);
        throw new Error('Failed to fetch projects information');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load account data');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setPasswordData(prev => ({ ...prev, [name]: value }));
    setPasswordErrors(prev => ({ ...prev, [name]: undefined }));
    setPasswordSuccess(null);
  };

  const validatePasswordForm = (): boolean => {
    const errors: Partial<ChangePasswordData> = {};

    if (!passwordData.old_password) {
      errors.old_password = 'Current password is required';
    }

    if (!passwordData.new_password) {
      errors.new_password = 'New password is required';
    } else if (passwordData.new_password.length < 6) {
      errors.new_password = 'Password must be at least 6 characters';
    }

    if (!passwordData.confirm_password) {
      errors.confirm_password = 'Please confirm your password';
    } else if (passwordData.new_password !== passwordData.confirm_password) {
      errors.confirm_password = 'Passwords do not match';
    }

    setPasswordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validatePasswordForm()) {
      return;
    }

    setChangingPassword(true);
    setPasswordSuccess(null);

    try {
      const response = await authFetch('/api/v1/users/me/password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          old_password: passwordData.old_password,
          new_password: passwordData.new_password,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Failed to change password' }));
        setPasswordErrors({ general: errorData.error || 'Failed to change password' });
        setChangingPassword(false);
        return;
      }

      setPasswordSuccess('Password changed successfully');
      setPasswordData({
        old_password: '',
        new_password: '',
        confirm_password: '',
      });
    } catch (err) {
      setPasswordErrors({ general: 'An error occurred. Please try again.' });
    } finally {
      setChangingPassword(false);
    }
  };

  const handleSetup2FA = async () => {
    try {
      const response = await authFetch('/api/v1/auth/2fa/setup', {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Failed to setup 2FA');
      }

      const data: Setup2FAResponse = await response.json();
      setTwoFASetup(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to setup 2FA');
    }
  };

  const handleConfirm2FA = async (e: React.FormEvent) => {
    e.preventDefault();
    setConfirming2FA(true);

    try {
      const response = await authFetch('/api/v1/auth/2fa/confirm', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: twoFACode,
        }),
      });

      if (!response.ok) {
        throw new Error('Invalid code');
      }

      await fetchUserData();
      setTwoFASetup(null);
      setTwoFACode('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid 2FA code');
    } finally {
      setConfirming2FA(false);
    }
  };

  const handleDisable2FA = async (e: React.FormEvent) => {
    e.preventDefault();
    setDisabling2FA(true);

    try {
      // First, send email code
      if (!twoFAEmailCode) {
        await authFetch('/api/v1/auth/2fa/send-code', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            action: 'disable',
          }),
        });
        setError(null);
        return;
      }

      // Then disable with code
      const response = await authFetch('/api/v1/auth/2fa/disable', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email_code: twoFAEmailCode,
        }),
      });

      if (!response.ok) {
        throw new Error('Invalid email code');
      }

      await fetchUserData();
      setTwoFAEmailCode('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid email code');
    } finally {
      setDisabling2FA(false);
    }
  };

  const formatPermission = (perm: string): string => {
    const permMap: Record<string, string> = {
      'project.view': 'View Project',
      'project.manage': 'Manage Project',
      'project.create': 'Create Project',
      'audit.view': 'View Audit',
      'membership.manage': 'Manage Memberships',
    };
    return permMap[perm] || perm;
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading account information...</span>
        </div>
      </div>
    );
  }

  if (error && !userInfo) {
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
      <h1>Account Settings</h1>

      <div className="card">
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem', borderBottom: '1px solid #e5e7eb', paddingBottom: '1rem' }}>
          <button 
            className={`btn ${activeTab === 'profile' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('profile')}
          >
            Profile
          </button>
          <button 
            className={`btn ${activeTab === '2fa' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('2fa')}
          >
            2FA Security
          </button>
          <button 
            className={`btn ${activeTab === 'projects' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('projects')}
          >
            Projects & Permissions
          </button>
        </div>

        {error && (
          <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
            <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {activeTab === 'profile' && userInfo && (
          <div>
            <h2>Profile Information</h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
              <div>
                <strong>Username:</strong>
                <div className="mt-1">{userInfo.username}</div>
              </div>
              <div>
                <strong>Email:</strong>
                <div className="mt-1">{userInfo.email}</div>
              </div>
              <div>
                <strong>Account Type:</strong>
                <div className="mt-1">
                  {userInfo.is_superuser ? (
                    <span className="status completed">Superuser</span>
                  ) : (
                    <span className="status pending">Regular User</span>
                  )}
                </div>
              </div>
              <div>
                <strong>Status:</strong>
                <div className="mt-1">
                  {userInfo.is_active ? (
                    <span className="status completed">Active</span>
                  ) : (
                    <span className="status failed">Inactive</span>
                  )}
                </div>
              </div>
              <div>
                <strong>Created:</strong>
                <div className="mt-1">{new Date(userInfo.created_at).toLocaleString()}</div>
              </div>
              {userInfo.last_login && (
                <div>
                  <strong>Last Login:</strong>
                  <div className="mt-1">{new Date(userInfo.last_login).toLocaleString()}</div>
                </div>
              )}
            </div>

            <h2>Change Password</h2>
            <form onSubmit={handlePasswordSubmit} style={{ maxWidth: '500px' }}>
              {passwordSuccess && (
                <div className="mb-4 p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50">
                  <p className="text-sm text-emerald-700 dark:text-emerald-400">{passwordSuccess}</p>
                </div>
              )}

              {passwordErrors.general && (
                <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
                  <p className="text-sm text-red-700 dark:text-red-400">{passwordErrors.general}</p>
                </div>
              )}

              <div style={{ marginBottom: '1rem' }}>
                <label htmlFor="old_password" className="block text-sm font-medium mb-2">
                  Current Password
                </label>
                <input
                  type="password"
                  id="old_password"
                  name="old_password"
                  value={passwordData.old_password}
                  onChange={handlePasswordChange}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  required
                />
                {passwordErrors.old_password && (
                  <p className="mt-1 text-sm text-red-600">{passwordErrors.old_password}</p>
                )}
              </div>

              <div style={{ marginBottom: '1rem' }}>
                <label htmlFor="new_password" className="block text-sm font-medium mb-2">
                  New Password
                </label>
                <input
                  type="password"
                  id="new_password"
                  name="new_password"
                  value={passwordData.new_password}
                  onChange={handlePasswordChange}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  required
                />
                {passwordErrors.new_password && (
                  <p className="mt-1 text-sm text-red-600">{passwordErrors.new_password}</p>
                )}
              </div>

              <div style={{ marginBottom: '1rem' }}>
                <label htmlFor="confirm_password" className="block text-sm font-medium mb-2">
                  Confirm New Password
                </label>
                <input
                  type="password"
                  id="confirm_password"
                  name="confirm_password"
                  value={passwordData.confirm_password}
                  onChange={handlePasswordChange}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                  required
                />
                {passwordErrors.confirm_password && (
                  <p className="mt-1 text-sm text-red-600">{passwordErrors.confirm_password}</p>
                )}
              </div>

              <button
                type="submit"
                disabled={changingPassword}
                className="btn btn-primary"
              >
                {changingPassword ? 'Changing Password...' : 'Change Password'}
              </button>
            </form>
          </div>
        )}

        {activeTab === '2fa' && userInfo && (
          <div>
            <h2>Two-Factor Authentication (2FA)</h2>
            
            {userInfo.two_fa_enabled ? (
              <div>
                <div className="mb-4 p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50">
                  <p className="text-sm text-emerald-700 dark:text-emerald-400">
                    <strong>2FA is enabled</strong>
                    {userInfo.two_fa_confirmed_at && (
                      <span className="ml-2">
                        (confirmed on {new Date(userInfo.two_fa_confirmed_at).toLocaleDateString()})
                      </span>
                    )}
                  </p>
                </div>

                <h3>Disable 2FA</h3>
                <p className="mb-4 text-sm text-slate-600 dark:text-[#ff4500]">
                  To disable 2FA, a code will be sent to your email. Enter the code below to confirm.
                </p>
                <form onSubmit={handleDisable2FA}>
                  <div style={{ marginBottom: '1rem', maxWidth: '400px' }}>
                    <label htmlFor="twoFAEmailCode" className="block text-sm font-medium mb-2">
                      Email Verification Code
                    </label>
                    <input
                      type="text"
                      id="twoFAEmailCode"
                      value={twoFAEmailCode}
                      onChange={(e) => setTwoFAEmailCode(e.target.value)}
                      placeholder="Enter code from email"
                      className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20"
                    />
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    {!twoFAEmailCode && (
                      <button
                        type="button"
                        onClick={handleDisable2FA}
                        className="btn btn-secondary"
                      >
                        Send Code
                      </button>
                    )}
                    {twoFAEmailCode && (
                      <>
                        <button
                          type="submit"
                          disabled={disabling2FA}
                          className="btn btn-danger"
                        >
                          {disabling2FA ? 'Disabling...' : 'Disable 2FA'}
                        </button>
                        <button
                          type="button"
                          onClick={() => setTwoFAEmailCode('')}
                          className="btn btn-secondary"
                        >
                          Cancel
                        </button>
                      </>
                    )}
                  </div>
                </form>
              </div>
            ) : (
              <div>
                <p className="mb-4 text-slate-600 dark:text-[#ff4500]">
                  Two-factor authentication adds an extra layer of security to your account.
                  After enabling 2FA, you'll need to enter a code from your authenticator app when logging in.
                </p>

                {!twoFASetup ? (
                  <button
                    onClick={handleSetup2FA}
                    className="btn btn-primary"
                  >
                    Enable 2FA
                  </button>
                ) : (
                  <div>
                    <h3>Setup 2FA</h3>
                    <div className="mb-4">
                      <p className="mb-2 text-sm">Scan this QR code with your authenticator app:</p>
                      <div style={{ display: 'inline-block', padding: '1rem', background: 'white', borderRadius: '8px' }}>
                        <img 
                          src={`data:image/png;base64,${twoFASetup.qr_image}`} 
                          alt="2FA QR Code"
                          style={{ maxWidth: '200px' }}
                        />
                      </div>
                    </div>
                    <div className="mb-4">
                      <p className="mb-2 text-sm">Or enter this secret manually:</p>
                      <code className="px-3 py-2 bg-slate-100 dark:bg-[#2d2d30] rounded text-sm font-mono">
                        {twoFASetup.secret}
                      </code>
                    </div>
                    <form onSubmit={handleConfirm2FA}>
                      <div style={{ marginBottom: '1rem', maxWidth: '400px' }}>
                        <label htmlFor="twoFACode" className="block text-sm font-medium mb-2">
                          Enter 6-digit code from your authenticator app
                        </label>
                        <input
                          type="text"
                          id="twoFACode"
                          value={twoFACode}
                          onChange={(e) => setTwoFACode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                          placeholder="000000"
                          maxLength={6}
                          className="w-full px-4 py-2.5 rounded-lg border border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20 text-center text-2xl tracking-widest font-mono"
                        />
                      </div>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          type="submit"
                          disabled={confirming2FA || twoFACode.length !== 6}
                          className="btn btn-primary"
                        >
                          {confirming2FA ? 'Confirming...' : 'Confirm & Enable 2FA'}
                        </button>
                        <button
                          type="button"
                          onClick={() => {
                            setTwoFASetup(null);
                            setTwoFACode('');
                          }}
                          className="btn btn-secondary"
                        >
                          Cancel
                        </button>
                      </div>
                    </form>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {activeTab === 'projects' && (
          <div>
            <h2>Projects & Permissions</h2>
            
            {isSuperuser && (
              <div className="mb-4 p-4 rounded-lg bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800/50">
                <p className="text-sm text-blue-700 dark:text-blue-400">
                  <strong>Superuser Access:</strong> You have full access to all projects and all permissions.
                </p>
              </div>
            )}

            {projects.length === 0 ? (
              <p>You don't have access to any projects.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Project ID</th>
                      <th>Project Name</th>
                      <th>Role</th>
                      <th>Permissions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {projects.map((projInfo) => (
                      <tr key={projInfo.project.id}>
                        <td className="font-mono text-xs">{projInfo.project.id}</td>
                        <td>
                          <div className="font-medium">{projInfo.project.name}</div>
                          {projInfo.project.description && (
                            <div className="text-xs text-slate-500 dark:text-[#ff4500]500 mt-0.5">
                              {projInfo.project.description}
                            </div>
                          )}
                        </td>
                        <td>
                          {projInfo.role ? (
                            <div>
                              <span className="status completed">{projInfo.role.name}</span>
                              {projInfo.role.description && (
                                <div className="text-xs text-slate-500 dark:text-[#ff4500]500 mt-0.5">
                                  {projInfo.role.description}
                                </div>
                              )}
                            </div>
                          ) : (
                            <span className="text-slate-400">-</span>
                          )}
                        </td>
                        <td>
                          {projInfo.permissions.length > 0 ? (
                            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.25rem' }}>
                              {projInfo.permissions.map((perm) => (
                                <span
                                  key={perm}
                                  className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 dark:bg-[#3e3e42] text-slate-700 dark:text-[#ff4500]400"
                                >
                                  {formatPermission(perm)}
                                </span>
                              ))}
                            </div>
                          ) : (
                            <span className="text-slate-400">No permissions</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

