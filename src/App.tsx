import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider } from './hooks/useAuth';
import { LoginPage } from './pages/Login';
import { SignupPage } from './pages/Signup';
import { DashboardPage } from './pages/Dashboard';

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/login"     element={<LoginPage />} />
            <Route path="/signup"    element={<SignupPage />} />
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/"          element={<Navigate to="/login" replace />} />
            <Route path="*"          element={<Navigate to="/login" replace />} />
          </Routes>
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
