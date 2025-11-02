import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './tw.css';

// Ensure light theme is always active (remove dark class)
if (typeof document !== 'undefined') {
  document.documentElement.classList.remove('dark');
}

const container = document.getElementById('root');
const root = createRoot(container!);

root.render(
  <BrowserRouter>
    <App />
  </BrowserRouter>
);
