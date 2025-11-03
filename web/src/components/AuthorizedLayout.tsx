import React, { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import logoImage from '../assets/floxy_logo.png';
import { useAuth } from '../auth/AuthContext';
import { Breadcrumbs } from './Breadcrumbs';

interface AuthorizedLayoutProps {
  children: React.ReactNode;
}

export const AuthorizedLayout: React.FC<AuthorizedLayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const { isAuthenticated, user, logout } = useAuth();
  const headerRef = useRef<HTMLElement>(null);
  const [headerHeight, setHeaderHeight] = useState(49);
  const [showUserMenu, setShowUserMenu] = useState(false);

  useEffect(() => {
    const updateHeaderHeight = () => {
      if (headerRef.current) {
        const height = headerRef.current.offsetHeight;
        setHeaderHeight(height);
      }
    };

    requestAnimationFrame(() => {
      updateHeaderHeight();
      setTimeout(updateHeaderHeight, 0);
    });

    window.addEventListener('resize', updateHeaderHeight);
    return () => window.removeEventListener('resize', updateHeaderHeight);
  }, []);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.remove('dark');
  }, []);

  return (
    <div className="min-h-screen flex flex-col">
      <header ref={headerRef} className="fixed top-0 left-0 right-0 z-50 glass-strong border-b border-slate-200/50 dark:border-slate-700/50">
        <div className="container">
          <div className="flex items-center justify-between h-12">
            <div className="flex items-center gap-3">
              <div 
                className="flex items-center cursor-pointer hover:opacity-80 transition-opacity" 
                style={{ marginLeft: '-100px' }}
                onClick={() => navigate('/tenants')}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    navigate('/tenants');
                  }
                }}
                aria-label="Go to home page"
              >
                <img 
                  src={logoImage} 
                  alt="Floxy Manager" 
                  className="h-12 w-auto object-contain"
                />
              </div>
            </div>

            <div className="flex items-center gap-2">
              {isAuthenticated && user && user.is_superuser && (
                <button
                  className="btn btn-outline p-2"
                  onClick={() => navigate('/admin')}
                  aria-label="Admin panel"
                  title="Admin panel"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </button>
              )}
              {isAuthenticated && user && (
                <div className="relative">
                  <button
                    className="btn btn-outline p-2"
                    onClick={() => setShowUserMenu(!showUserMenu)}
                    aria-label="User menu"
                    title={user.username}
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </button>

                  {showUserMenu && (
                    <>
                      <div 
                        className="fixed inset-0 z-40" 
                        onClick={() => setShowUserMenu(false)}
                      />
                      <div className="absolute right-0 mt-2 w-48 glass-strong rounded-lg shadow-lg z-50 border border-slate-200/50 dark:border-[#3e3e42] overflow-hidden">
                        <div className="px-4 py-3 border-b border-slate-200/50 dark:border-[#3e3e42]">
                          <p className="text-sm font-semibold text-slate-900 dark:text-[#ff6b35]">
                            {user.username}
                          </p>
                          {user.email && (
                            <p className="text-xs text-slate-600 dark:text-[#ff4500] mt-1">
                              {user.email}
                            </p>
                          )}
                        </div>
                        <div className="py-1">
                          <button
                            onClick={() => {
                              navigate('/account');
                              setShowUserMenu(false);
                            }}
                            className="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-[#ff6b35] hover:bg-slate-100 dark:hover:bg-[#2d2d30] transition-colors flex items-center gap-2"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                            Account
                          </button>
                          <button
                            onClick={() => {
                              logout();
                              setShowUserMenu(false);
                              navigate('/login');
                            }}
                            className="w-full text-left px-4 py-2 text-sm text-slate-700 dark:text-[#ff6b35] hover:bg-slate-100 dark:hover:bg-[#2d2d30] transition-colors flex items-center gap-2"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                            </svg>
                            Logout
                          </button>
                        </div>
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      <div className="flex flex-1 min-h-screen" style={{ paddingTop: `${headerHeight}px` }}>
        <main className="flex-1 container py-6 h-full overflow-auto">
          <Breadcrumbs />
          {children}
        </main>
      </div>
    </div>
  );
};

