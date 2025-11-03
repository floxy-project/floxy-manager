import React from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import logoImage from '../assets/floxy_logo.png';

const SSOError: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const error = searchParams.get('error') || 'SSO authentication failed';
  const errorDetails = searchParams.get('details') || '';

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
            <h1 className="text-2xl font-bold mb-2 text-slate-600 dark:text-slate-300">
              SSO Authentication Failed
            </h1>
          </div>

          <div className="mb-6 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
            <div className="flex items-start gap-3">
              <svg
                className="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <div className="flex-1">
                <p className="text-sm font-medium text-red-700 dark:text-red-400 mb-1">
                  {error}
                </p>
                {errorDetails && (
                  <p className="text-xs text-red-600 dark:text-red-500 mt-2">
                    {errorDetails}
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="space-y-3">
            <button
              onClick={() => navigate('/login')}
              className="w-full btn btn-primary py-3"
            >
              Return to Login
            </button>
            <button
              onClick={() => window.location.reload()}
              className="w-full btn btn-secondary py-3"
            >
              Try Again
            </button>
          </div>

          <div className="mt-6 pt-6 border-t border-slate-200 dark:border-[#3e3e42]">
            <p className="text-xs text-center text-slate-500 dark:text-slate-400">
              If this problem persists, please contact your administrator.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SSOError;

