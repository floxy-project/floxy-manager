import React, { useEffect, useState, useRef } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { Menu, Settings, User, LogOut, Home, RefreshCw, Zap, Inbox, BarChart3, ChevronLeft, ChevronRight, Users, FileText } from 'lucide-react';
import logoImage from '../assets/floxy_logo.png';
import { useAuth } from '../auth/AuthContext';
import { useRBAC } from '../auth/permissions';
import { Breadcrumbs } from './Breadcrumbs';

interface TenantProjectLayoutProps {
  children: React.ReactNode;
}

export const TenantProjectLayout: React.FC<TenantProjectLayoutProps> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { tenantId, projectId } = useParams<{ tenantId: string; projectId: string }>();
  const { isAuthenticated, user, logout } = useAuth();
  const rbac = useRBAC(projectId);
  const headerRef = useRef<HTMLElement>(null);
  const [headerHeight, setHeaderHeight] = useState(36);
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
          <div className="flex items-center justify-between h-9">
            <div className="flex items-center gap-3">
              <button 
                className="md:hidden btn btn-outline p-2" 
                onClick={() => setSidebarOpen(s => !s)} 
                aria-label="Toggle menu"
              >
                <Menu className="w-5 h-5" />
              </button>
              
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
                  <Settings className="w-5 h-5" />
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
                    <User className="w-5 h-5" />
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
              icon={<Home className="w-5 h-5" />}
            />
            <NavLink 
              to="/workflows" 
              label="Workflows"
              icon={<RefreshCw className="w-5 h-5" />}
            />
            <NavLink 
              to="/instances" 
              label="Instances"
              icon={<Zap className="w-5 h-5" />}
            />
            <NavLink 
              to="/dlq" 
              label="Dead Letter Queue"
              icon={<Inbox className="w-5 h-5" />}
            />
            <NavLink 
              to="/stats" 
              label="Statistics"
              icon={<BarChart3 className="w-5 h-5" />}
            />
            <NavLink 
              to="/memberships" 
              label="Members"
              icon={<Users className="w-5 h-5" />}
            />
            {(rbac.isSuperuser || rbac.canViewAudit()) && (
              <NavLink 
                to="/audit-log" 
                label="Audit Log"
                icon={<FileText className="w-5 h-5" />}
              />
            )}
          </nav>

          <div className="p-2 border-t border-slate-200/50 dark:border-slate-700/50">
            <button
              className={`btn btn-outline w-full justify-center ${!sidebarOpen ? 'px-2' : ''}`}
              onClick={() => setSidebarOpen(!sidebarOpen)}
              aria-label={sidebarOpen ? "Collapse sidebar" : "Expand sidebar"}
              title={!sidebarOpen ? 'Expand menu' : undefined}
            >
              {sidebarOpen ? (
                <ChevronLeft className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
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

