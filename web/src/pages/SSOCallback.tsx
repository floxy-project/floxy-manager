import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import logoImage from '../assets/floxy_logo.png';

const SSOCallback: React.FC = () => {
  const [searchParams] = useSearchParams();
  const { handleSSOCallback, loginWithTokens, isAuthenticated, error: authError } = useAuth();
  const [isProcessing, setIsProcessing] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const processCallback = async () => {
      // Check if we have tokens from backend redirect (OAuth-like flow)
      const accessToken = searchParams.get('access_token');
      const refreshToken = searchParams.get('refresh_token');

      if (accessToken && refreshToken) {
        try {
          await loginWithTokens(accessToken, refreshToken);
          navigate('/', { replace: true });
          return;
        } catch (err) {
          console.error('Failed to login with tokens:', err);
          setError('Failed to complete authentication');
          setIsProcessing(false);
          return;
        }
      }

      // Check if we have SAML response (SAML flow)
      const samlResponse = searchParams.get('SAMLResponse');
      const relayState = searchParams.get('RelayState');

      if (samlResponse && relayState) {
        try {
          await handleSSOCallback(samlResponse, relayState);
          navigate('/', { replace: true });
          return;
        } catch (err) {
          console.error('SSO callback processing failed:', err);
          setError(authError || 'SSO authentication failed');
          setIsProcessing(false);
          return;
        }
      }

      // No valid parameters found
      setError('Invalid callback. Missing required parameters.');
      setIsProcessing(false);
    };

    processCallback();
  }, [searchParams, handleSSOCallback, loginWithTokens, navigate, authError]);

  // If already authenticated, redirect to dashboard
  if (isAuthenticated) {
    navigate('/', { replace: true });
    return null;
  }

  // Show error if processing is complete and there's an error
  if (!isProcessing && error) {
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
            </div>
            
            <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
              <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
            </div>
            
            <button
              onClick={() => navigate('/login')}
              className="w-full btn btn-primary py-3"
            >
              Return to Login
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Show loading state
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
            <h2 className="text-xl font-semibold mb-2 text-slate-900 dark:text-[#ff6b35]">
              Completing SSO Authentication
            </h2>
            <p className="text-sm text-slate-600 dark:text-[#ff4500]">
              Please wait while we complete your sign-in...
            </p>
          </div>

          {authError && (
            <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
              <p className="text-sm text-red-700 dark:text-red-400">{authError}</p>
            </div>
          )}

          <div className="flex justify-center py-8">
            <div className="w-12 h-12 border-4 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SSOCallback;

