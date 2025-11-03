import React, { useState, useEffect } from 'react';
import apiClient, { SSOProvider } from '../utils/api';
import { useAuth } from '../auth/AuthContext';

interface SSOButtonProps {
  fullWidth?: boolean;
  disabled?: boolean;
  onProvidersLoaded?: (hasProviders: boolean) => void;
}

const SSOButton: React.FC<SSOButtonProps> = ({
  fullWidth = false,
  disabled = false,
  onProvidersLoaded
}) => {
  const { isLoading } = useAuth();
  const [providers, setProviders] = useState<SSOProvider[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(false);
  const [initiatingSSO, setInitiatingSSO] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  useEffect(() => {
    const fetchProviders = async () => {
      setLoadingProviders(true);
      try {
        const response = await apiClient.getSSOProviders();
        const providersList = response.data.providers || [];
        setProviders(providersList);
        if (onProvidersLoaded) {
          onProvidersLoaded(providersList.length > 0);
        }
      } catch (error) {
        console.error('Failed to fetch SSO providers:', error);
        setProviders([]);
        if (onProvidersLoaded) {
          onProvidersLoaded(false);
        }
      } finally {
        setLoadingProviders(false);
      }
    };

    fetchProviders();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSSOInitiate = async (providerName: string) => {
    setInitiatingSSO(true);
    setShowMenu(false);
    
    try {
      const response = await apiClient.sSOInitiate(providerName);
      if (response.data && response.data.redirect_url) {
        window.location.href = response.data.redirect_url;
      }
    } catch (error) {
      console.error('SSO initiation failed:', error);
      setInitiatingSSO(false);
    }
  };

  const handleClick = () => {
    if (providers.length === 1) {
      // If there is only one provider, initiate SSO immediately
      handleSSOInitiate(providers[0].name);
    } else if (providers.length > 1) {
      // If there are multiple providers, show the menu
      setShowMenu(!showMenu);
    }
  };

  // If there are no available providers, don't show the button
  if (providers.length === 0 && !loadingProviders) {
    return null;
  }

  return (
    <div className={`relative ${fullWidth ? 'w-full' : ''}`}>
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled || isLoading || loadingProviders || initiatingSSO}
        className={`${fullWidth ? 'w-full' : ''} btn btn-secondary py-3 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2`}
      >
        {loadingProviders || initiatingSSO ? (
          <div className="w-5 h-5 border-2 border-slate-600 dark:border-slate-400 border-t-transparent rounded-full animate-spin"></div>
        ) : (
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
            />
          </svg>
        )}
        <span>
          {initiatingSSO ? 'Redirecting...' : 'SSO'}
        </span>
      </button>

      {showMenu && providers.length > 1 && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setShowMenu(false)}
          />
          <div className="absolute top-full mt-2 right-0 w-56 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-[#3e3e42] rounded-lg shadow-lg z-20">
            <div className="py-1">
              {providers.map((provider) => (
                <button
                  key={provider.name}
                  onClick={() => handleSSOInitiate(provider.name)}
                  disabled={initiatingSSO}
                  className="w-full px-4 py-2 text-left text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-[#2d2d30] flex items-center gap-3 disabled:opacity-50"
                >
                  {provider.icon_url ? (
                    <img
                      src={provider.icon_url}
                      alt={provider.display_name}
                      className="w-5 h-5 object-contain"
                    />
                  ) : (
                    <div className="w-5 h-5 rounded-full bg-slate-300 dark:bg-[#3e3e42]"></div>
                  )}
                  <span>{provider.display_name || provider.name}</span>
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default SSOButton;

