import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import logoImage from '../assets/floxy_logo.png';

interface LoginFormData {
  usernameOrEmail: string;
  password: string;
}

interface FormErrors {
  usernameOrEmail?: string;
  password?: string;
  general?: string;
}

export const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, is2FARequired, hasTmpPassword, error: authError } = useAuth();
  const [formData, setFormData] = useState<LoginFormData>({
    usernameOrEmail: '',
    password: ''
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    // Check for success message from location state
    const state = location.state as { message?: string } | null;
    if (state?.message) {
      setSuccessMessage(state.message);
      // Clear the state to avoid showing message on refresh
      window.history.replaceState({}, document.title);
    }
  }, [location]);

  useEffect(() => {
    // Handle navigation after login
    if (is2FARequired) {
      navigate('/2fa', { 
        state: { 
          username: formData.usernameOrEmail.includes('@') 
            ? formData.usernameOrEmail.split('@')[0] 
            : formData.usernameOrEmail
        } 
      });
      return;
    }
    
    if (hasTmpPassword) {
      navigate('/change-password');
      return;
    }
  }, [is2FARequired, hasTmpPassword, navigate, formData.usernameOrEmail]);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    // Validate username/email
    if (!formData.usernameOrEmail.trim()) {
      newErrors.usernameOrEmail = 'Username or email is required';
    } else if (formData.usernameOrEmail.includes('@')) {
      // If contains @, validate as email
      if (!validateEmail(formData.usernameOrEmail)) {
        newErrors.usernameOrEmail = 'Please enter a valid email address';
      }
    } else {
      // Validate as username (minimum 3 characters)
      if (formData.usernameOrEmail.trim().length < 3) {
        newErrors.usernameOrEmail = 'Username must be at least 3 characters';
      }
    }

    // Validate password
    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      await login(formData.usernameOrEmail, formData.password);
      
      // Navigation will be handled by useEffect watching is2FARequired and hasTmpPassword
    } catch (error) {
      console.error('Login error:', error);
      if (authError) {
        setErrors({ general: authError });
      } else {
        setErrors({ general: 'An error occurred. Please try again later.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Clear error for this field when user starts typing
    if (errors[name as keyof FormErrors]) {
      setErrors(prev => ({
        ...prev,
        [name]: undefined
      }));
    }
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
            <h1 className="text-2xl font-bold mb-2 text-slate-600" >Log In</h1>
            <p className="text-slate-600 dark:text-[#ff4500]">
              Enter your credentials to access Floxy Manager
            </p>
          </div>

          {successMessage && (
            <div className="mb-4 p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50">
              <p className="text-sm text-emerald-700 dark:text-emerald-400">{successMessage}</p>
            </div>
          )}

          {errors.general && (
            <div className="mb-4 p-4 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800/50">
              <p className="text-sm text-red-700 dark:text-red-400">{errors.general}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label 
                htmlFor="usernameOrEmail" 
                className="block text-sm font-medium mb-2 text-slate-700 dark:text-[#ff6b35]"
              >
                Username or Email
              </label>
              <input
                type="text"
                id="usernameOrEmail"
                name="usernameOrEmail"
                value={formData.usernameOrEmail}
                onChange={handleChange}
                className={`w-full px-4 py-2.5 rounded-lg border transition-colors ${
                  errors.usernameOrEmail
                    ? 'border-red-300 dark:border-red-700 focus:border-red-500 dark:focus:border-red-500'
                    : 'border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35]'
                } bg-white dark:bg-[#252526] text-slate-900 dark:text-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20`}
                placeholder="Enter username or email"
                disabled={isLoading}
              />
              {errors.usernameOrEmail && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                  {errors.usernameOrEmail}
                </p>
              )}
            </div>

            <div>
              <label 
                htmlFor="password" 
                className="block text-sm font-medium mb-2 text-slate-700 dark:text-[#ff6b35]"
              >
                Password
              </label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                className={`w-full px-4 py-2.5 rounded-lg border transition-colors ${
                  errors.password
                    ? 'border-red-300 dark:border-red-700 focus:border-red-500 dark:focus:border-red-500'
                    : 'border-slate-300 dark:border-[#3e3e42] focus:border-slate-500 dark:focus:border-[#ff6b35]'
                } bg-white dark:bg-[#252526] text-slate-900 dark:text-[#ff6b35] focus:outline-none focus:ring-2 focus:ring-slate-200 dark:focus:ring-[#ff6b35]/20`}
                placeholder="Enter password"
                disabled={isLoading}
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                  {errors.password}
                </p>
              )}
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full btn btn-primary py-3 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <span className="flex items-center justify-center gap-2">
                  <div className="w-5 h-5 border-2 border-white dark:border-slate-900 border-t-transparent rounded-full animate-spin"></div>
                  Logging in...
                </span>
              ) : (
                'Log In'
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

