import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authFetch } from '../utils/api';

interface Tenant {
  ID: number;
  Name: string;
  CreatedAt: string;
}

export const Tenants: React.FC = () => {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchTenants = async () => {
      try {
        const response = await authFetch('/api/v1/tenants');
        if (!response.ok) {
          throw new Error('Failed to fetch tenants');
        }
        const data = await response.json();
        setTenants(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchTenants();
  }, []);

  const handleSelectTenant = (tenantId: number) => {
    navigate(`/tenants/${tenantId}/projects`);
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="flex items-center justify-center gap-2">
          <div className="w-5 h-5 border-2 border-slate-300 dark:border-slate-600 border-t-slate-600 dark:border-t-slate-400 rounded-full animate-spin"></div>
          <span>Loading tenants...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error">
        <div className="flex items-center justify-center gap-2">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>Error: {error}</span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <h1>Select Tenant</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
        {tenants.map((tenant) => (
          <div
            key={tenant.ID}
            onClick={() => handleSelectTenant(tenant.ID)}
            className="card cursor-pointer transition-all duration-200 hover:shadow-lg hover:scale-105"
            style={{
              background: 'rgba(255, 255, 255, 0.7)',
              backdropFilter: 'blur(12px) saturate(150%)',
              WebkitBackdropFilter: 'blur(12px) saturate(150%)',
              border: '1px solid rgba(255, 255, 255, 0.5)',
            }}
          >
            <div className="flex flex-col items-center text-center p-6">
              <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-4">
                <span className="text-2xl font-bold text-white">
                  {tenant.Name?.charAt(0).toUpperCase() || 'T'}
                </span>
              </div>
              <h2 className="text-xl font-semibold mb-2">{tenant.Name}</h2>
              <p className="text-sm text-slate-600 dark:text-[#ff4500]500">
                Created {new Date(tenant.CreatedAt).toLocaleDateString()}
              </p>
            </div>
          </div>
        ))}
      </div>

      {tenants.length === 0 && (
        <div className="text-center py-12">
          <p className="text-slate-500 dark:text-[#ff4500]500">No tenants found</p>
        </div>
      )}
    </div>
  );
};

