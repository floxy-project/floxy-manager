import React, { useEffect, useState, useRef } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import logoImage from '../assets/floxy_logo.png';
import { useAuth } from '../hooks/useAuth';
import { Breadcrumbs } from './Breadcrumbs';

interface TenantProjectLayoutProps {
  children: React.ReactNode;
}

export const TenantProjectLayout: React.FC<TenantProjectLayoutProps> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const { isAuthenticated, user, logout } = useAuth();
  const headerRef = useRef<HTMLElement>(null);
  const [headerHeight, setHeaderHeight] = useState(49);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(() => {
    if (typeof window === 'undefined') return true;
    const stored = localStorage.getItem('sidebarOpen');
    return stored === null ? true : stored === 'true';
  });

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

  useEffect(() => {
    localStorage.setItem('sidebarOpen', String(sidebarOpen));
  }, [sidebarOpen]);

  const NavLink: React.FC<{ to: string; label: string; icon: React.ReactNode }> = ({ to, label, icon }) => {
    const active = location.pathname === to || location.pathname.startsWith(to + '/');
    const fullPath = `/tenants/${tenantId}/projects/${projectId}${to}`;
    
    return (
      <Link
        to={fullPath}
        className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 text-sm font-medium group ${
          active 
            ? 'glass-strong text-slate-900 dark:text-[#ff6b35] shadow-sm font-semibold' 
            : 'text-slate-700 dark:text-[#ff4500] hover:text-slate-900 dark:hover:text-[#ff6b35]'
        } ${!sidebarOpen ? 'justify-center' : ''}`}
        title={!sidebarOpen ? label : undefined}
        style={!active ? {
          transition: 'all 0.2s ease',
        } : {}}
        onMouseEnter={(e) => {
          if (!active) {
            e.currentTarget.style.background = 'rgba(255, 255, 255, 0.5)';
            e.currentTarget.style.backdropFilter = 'blur(12px) saturate(150%)';
          }
        }}
        onMouseLeave={(e) => {
          if (!active) {
            e.currentTarget.style.background = 'transparent';
            e.currentTarget.style.backdropFilter = 'none';
          }
        }}
        onClick={() => {
          if (window.innerWidth < 768) {
            setSidebarOpen(false);
          }
        }}
      >
        <span className={`flex-shrink-0 ${active ? 'text-slate-900 dark:text-[#ff6b35]' : 'text-slate-600 dark:text-[#ff4500]'}`}>
          {icon}
        </span>
        {sidebarOpen && <span className="whitespace-nowrap">{label}</span>}
      </Link>
    );
  };

  return (
    <div className="min-h-screen flex flex-col">
      <header ref={headerRef} className="fixed top-0 left-0 right-0 z-50 glass-strong border-b border-slate-200/50 dark:border-slate-700/50">
        <div className="container">
          <div className="flex items-center justify-between h-12">
            <div className="flex items-center gap-3">
              <button 
                className="md:hidden btn btn-outline p-2" 
                onClick={() => setSidebarOpen(s => !s)} 
                aria-label="Toggle menu"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
              
              <div className="flex items-center" style={{ marginLeft: '-100px' }}>
                <img 
                  src={logoImage} 
                  alt="Floxy Manager" 
                  className="h-12 w-auto object-contain"
                />
              </div>
            </div>

            <div className="flex items-center gap-2">
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
        <aside className={`glass-strong transition-all duration-300 flex flex-col fixed z-30 ${
          sidebarOpen ? 'w-64' : 'w-16'
        }`}
        style={{
          top: `${headerHeight}px`,
          height: `calc(100vh - ${headerHeight}px)`,
          left: 0
        }}>
          <nav className="flex-1 overflow-y-auto p-2 space-y-1">
            <NavLink 
              to="/dashboard" 
              label="Dashboard"
              icon={
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              }
            />
            <NavLink 
              to="/workflows" 
              label="Workflows"
              icon={
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              }
            />
            <NavLink 
              to="/instances" 
              label="Instances"
              icon={
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              }
            />
            <NavLink 
              to="/dlq" 
              label="Dead Letter Queue"
              icon={
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                </svg>
              }
            />
            <NavLink 
              to="/stats" 
              label="Statistics"
              icon={
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              }
            />
          </nav>

          <div className="p-2 border-t border-slate-200/50 dark:border-slate-700/50">
            <button
              className={`btn btn-outline w-full justify-center ${!sidebarOpen ? 'px-2' : ''}`}
              onClick={() => setSidebarOpen(!sidebarOpen)}
              aria-label={sidebarOpen ? "Collapse sidebar" : "Expand sidebar"}
              title={!sidebarOpen ? 'Expand menu' : undefined}
            >
              {sidebarOpen ? (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
                </svg>
              ) : (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
                </svg>
              )}
              {sidebarOpen && <span className="ml-2"></span>}
            </button>
          </div>
        </aside>

        <div className={`flex-1 transition-all duration-300 min-w-0 ${
          sidebarOpen ? 'ml-64' : 'ml-16'
        }`}>
          <main className="container py-6 h-full overflow-auto">
            <Breadcrumbs />
            {children}
          </main>
        </div>
      </div>

      {sidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  );
};

