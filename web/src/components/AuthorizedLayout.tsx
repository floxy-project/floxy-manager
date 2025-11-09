import React, { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Settings, User, LogOut } from 'lucide-react';
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
  const [headerHeight, setHeaderHeight] = useState(36);
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
          <div className="flex items-center justify-between h-9">
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
                  className="h-9 w-auto object-contain"
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
                  <Settings className="w-4 h-4" />
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
                    <User className="w-4 h-4" />
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
                            <User className="w-4 h-4" />
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
                            <LogOut className="w-4 h-4" />
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

