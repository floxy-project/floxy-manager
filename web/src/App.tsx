import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Workflows } from './pages/Workflows';
import { WorkflowDetail } from './pages/WorkflowDetail';
import { Instances } from './pages/Instances';
import { InstanceDetail } from './pages/InstanceDetail';
import { Stats } from './pages/Stats';

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/workflows" element={<Workflows />} />
        <Route path="/workflows/:id" element={<WorkflowDetail />} />
        <Route path="/instances" element={<Instances />} />
        <Route path="/instances/:id" element={<InstanceDetail />} />
        <Route path="/stats" element={<Stats />} />
      </Routes>
    </Layout>
  );
}

export default App;
