import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Workflows } from './pages/Workflows';
import { WorkflowDetail } from './pages/WorkflowDetail';
import { Instances } from './pages/Instances';
import { InstanceDetail } from './pages/InstanceDetail';
import { Stats } from './pages/Stats';
import { DLQ } from './pages/DLQ';
import { DLQDetail } from './pages/DLQDetail';

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
        <Route path="/dlq" element={<DLQ />} />
        <Route path="/dlq/:id" element={<DLQDetail />} />
      </Routes>
    </Layout>
  );
}

export default App;
