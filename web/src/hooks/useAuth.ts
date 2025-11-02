import { useState, useEffect } from 'react';

interface User {
  username: string;
  email?: string;
}

export const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return localStorage.getItem('isAuthenticated') === 'true';
  });
  
  const [user, setUser] = useState<User | null>(() => {
    if (typeof window === 'undefined') return null;
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  });

  useEffect(() => {
    // Sync with localStorage on mount
    const checkAuth = () => {
      const authStatus = localStorage.getItem('isAuthenticated') === 'true';
      const userStr = localStorage.getItem('user');
      setIsAuthenticated(authStatus);
      setUser(userStr ? JSON.parse(userStr) : null);
    };

    checkAuth();

    // Listen for storage changes (e.g., from other tabs)
    window.addEventListener('storage', checkAuth);
    return () => window.removeEventListener('storage', checkAuth);
  }, []);

  const logout = () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
    localStorage.removeItem('isAuthenticated');
    setIsAuthenticated(false);
    setUser(null);
  };

  const login = (userData: User, token?: string) => {
    localStorage.setItem('authToken', token || 'mock-token');
    localStorage.setItem('user', JSON.stringify(userData));
    localStorage.setItem('isAuthenticated', 'true');
    setIsAuthenticated(true);
    setUser(userData);
  };

  return {
    isAuthenticated,
    user,
    login,
    logout
  };
};

