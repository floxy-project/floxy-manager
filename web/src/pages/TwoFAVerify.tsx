import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import logoImage from '../assets/floxy_logo.png';

interface LocationState {
  sessionId?: string;
  username?: string;
}

export const TwoFAVerify: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string | null>(null);

  useEffect(() => {
    const state = location.state as LocationState | null;
    if (state?.sessionId) {
      setSessionId(state.sessionId);
    } else {
      // If no sessionId, redirect back to login
      navigate('/login', { 
        state: { message: 'Please log in first to verify 2FA' } 
      });
    }
  }, [location, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!sessionId) {
      setError('Session ID is missing. Please log in again.');
      return;
    }

    if (code.length !== 6) {
      setError('Please enter a 6-digit code');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/auth/2fa/verify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: code,
          session_id: sessionId,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Invalid 2FA code' }));
        setError(errorData.error || 'Invalid 2FA code. Please try again.');
        setIsLoading(false);
        return;
      }

      const data = await response.json();

      // Get username from location state
      const state = location.state as LocationState | null;
      const username = state?.username || 'user';

      // Save refresh token if provided
      if (data.refresh_token) {
        localStorage.setItem('refreshToken', data.refresh_token);
      }

      // Fetch user info to get complete user data
      let userEmail: string | undefined;
      try {
        const userResponse = await fetch('/api/v1/users/me', {
          headers: {
            'Authorization': `Bearer ${data.access_token}`,
          },
        });

        if (userResponse.ok) {
          const userData = await userResponse.json();
          userEmail = userData.email;
          // Check if user has temporary password
          // If so, we'll handle redirect after login
        }
      } catch (err) {
        console.error('Failed to fetch user info:', err);
        // Continue with basic user info
      }

      // Save authentication tokens and user info
      login(
        { 
          username: username,
          email: userEmail,
        },
        data.access_token || ''
      );

      // Redirect to dashboard
      // Note: If user has temporary password, they will be redirected from protected routes
      navigate('/');
    } catch (error) {
      console.error('2FA verification error:', error);
      setError('An error occurred. Please try again later.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, '').slice(0, 6);
    setCode(value);
    setError(null);
  };

  const handleBackToLogin = () => {
    navigate('/login');
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="card">
          <div className="text-center mb-8">
            <div className="flex justify-center mb-6">
              <img 
                src={logoImage} 
                alt="Floxy Manager" 
                className="h-16 w-auto object-contain"
              />
            </div>
            <h1 className="text-2xl font-bold mb-2 text-slate-600">Two-Factor Authentication</h1>
            <p className="text-slate-600 dark:text-[#ff4500]">
              Enter the 6-digit code from your authenticator app
            </p>
          </div>

          {error && (
            <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
              <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label 
                htmlFor="code" 
                className="block text-sm font-medium mb-2 text-slate-700 dark:text-[#ff6b35]"
              >
                Authentication Code
              </label>
              <input
                type="text"
                id="code"
                name="code"
                value={code}
                onChange={handleCodeChange}
                className={`w-full px-4 py-2.5 rounded-lg border transition-colors text-center text-2xl tracking-widest font-mono focus:outline-none focus:ring-2 ${
                  error
                    ? 'border-red-300 dark:border-red-700 focus:border-red-500 dark:focus:border-red-500'
                    : 'border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35]'
                } bg-white dark:bg-[#252526] text-slate-900 dark:text-[#ff6b35] focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20`}
                placeholder="000000"
                maxLength={6}
                disabled={isLoading || !sessionId}
                autoComplete="off"
                autoFocus
              />
              <p className="mt-2 text-xs text-slate-500 dark:text-[#ff4500] text-center">
                Enter the 6-digit code from your authenticator app
              </p>
            </div>

            <div className="flex gap-3">
              <button
                type="button"
                onClick={handleBackToLogin}
                disabled={isLoading}
                className="flex-1 btn btn-secondary py-3 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Back to Login
              </button>
              <button
                type="submit"
                disabled={isLoading || code.length !== 6 || !sessionId}
                className="flex-1 btn btn-primary py-3 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? (
                  <span className="flex items-center justify-center gap-2">
                    <div className="w-5 h-5 border-2 border-white dark:border-slate-900 border-t-transparent rounded-full animate-spin"></div>
                    Verifying...
                  </span>
                ) : (
                  'Verify'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

