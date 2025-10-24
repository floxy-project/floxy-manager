import React from 'react';
import { Link, useLocation } from 'react-router-dom';

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();

  return (
    <div>
      <header className="header">
        <div className="container">
          <nav className="nav">
            <Link to="/" className={location.pathname === '/' ? 'active' : ''}>
              Dashboard
            </Link>
            <Link to="/workflows" className={location.pathname.startsWith('/workflows') ? 'active' : ''}>
              Workflows
            </Link>
            <Link to="/instances" className={location.pathname.startsWith('/instances') ? 'active' : ''}>
              Instances
            </Link>
            <Link to="/stats" className={location.pathname === '/stats' ? 'active' : ''}>
              Statistics
            </Link>
          </nav>
        </div>
      </header>
      <main className="container">
        {children}
      </main>
    </div>
  );
};
