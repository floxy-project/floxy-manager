import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthorizedLayout } from './components/AuthorizedLayout';
import { TenantProjectLayout } from './components/TenantProjectLayout';
import { Tenants } from './pages/Tenants';
import { Projects } from './pages/Projects';
import { Dashboard } from './pages/Dashboard';
import { Workflows } from './pages/Workflows';
import { WorkflowDetail } from './pages/WorkflowDetail';
import { Instances } from './pages/Instances';
import { InstanceDetail } from './pages/InstanceDetail';
import { Stats } from './pages/Stats';
import { DLQ } from './pages/DLQ';
import { DLQDetail } from './pages/DLQDetail';
import { Login } from './pages/Login';
import { ChangePassword } from './pages/ChangePassword';
import { TwoFAVerify } from './pages/TwoFAVerify';
import { Account } from './pages/Account';
import { Admin } from './pages/Admin';
import { Memberships } from './pages/Memberships';
import { useAuth } from './auth/AuthContext';

// Protected Route Component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

// Public Route Component (redirect to dashboard if already logged in)
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  
  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }
  
  return <>{children}</>;
};

function App() {
  return (
    <Routes>
      <Route path="/login" element={
        <PublicRoute>
          <Login />
        </PublicRoute>
      } />
      <Route path="/change-password" element={
        <ProtectedRoute>
          <ChangePassword />
        </ProtectedRoute>
      } />
      <Route path="/2fa" element={<TwoFAVerify />} />
      <Route path="/account" element={
        <ProtectedRoute>
          <AuthorizedLayout>
            <Account />
          </AuthorizedLayout>
        </ProtectedRoute>
      } />
      <Route path="/admin" element={
        <ProtectedRoute>
          <AuthorizedLayout>
            <Admin />
          </AuthorizedLayout>
        </ProtectedRoute>
      } />
      <Route path="/tenants" element={
        <ProtectedRoute>
          <AuthorizedLayout>
            <Tenants />
          </AuthorizedLayout>
        </ProtectedRoute>
      } />
      <Route path="/tenants/:tenantId/projects" element={
        <ProtectedRoute>
          <AuthorizedLayout>
            <Projects />
          </AuthorizedLayout>
        </ProtectedRoute>
      } />
      <Route path="/tenants/:tenantId/projects/:projectId/*" element={
        <ProtectedRoute>
          <TenantProjectLayout>
            <Routes>
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="workflows" element={<Workflows />} />
              <Route path="workflows/:id" element={<WorkflowDetail />} />
              <Route path="instances" element={<Instances />} />
              <Route path="instances/:id" element={<InstanceDetail />} />
              <Route path="stats" element={<Stats />} />
              <Route path="dlq" element={<DLQ />} />
              <Route path="dlq/:id" element={<DLQDetail />} />
              <Route path="memberships" element={<Memberships />} />
              <Route path="*" element={<Navigate to="dashboard" replace />} />
            </Routes>
          </TenantProjectLayout>
        </ProtectedRoute>
      } />
      <Route path="/" element={
        <ProtectedRoute>
          <Navigate to="/tenants" replace />
        </ProtectedRoute>
      } />
    </Routes>
  );
}

export default App;
