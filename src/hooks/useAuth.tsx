import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import axios from 'axios';

export interface User {
  id: string;
  email: string;
  fullName: string;
  createdAt: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

const API_BASE = (import.meta as any).env?.VITE_API_URL || '/api';

axios.defaults.baseURL = API_BASE;

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
  login: (email: string, password: string, rememberMe?: boolean) => Promise<{ success: boolean; user?: User; token?: string; error?: string }>;
  register: (email: string, password: string, fullName: string) => Promise<{ success: boolean; user?: User; token?: string; error?: string }>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(() => {
    const saved = localStorage.getItem('vana-user');
    return saved ? JSON.parse(saved) : null;
  });

  const [token, setToken] = useState<string | null>(() => {
    return localStorage.getItem('vana-token');
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // attach token to axios if present
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
      delete axios.defaults.headers.common['Authorization'];
    }
  }, [token]);

  const login = useCallback(async (email: string, password: string, rememberMe: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<AuthResponse>('/auth/login', { email, password });
      const { token: newToken, user: newUser } = response.data;
      setToken(newToken);
      setUser(newUser);
      localStorage.setItem('vana-token', newToken);
      localStorage.setItem('vana-user', JSON.stringify(newUser));
      if (rememberMe) localStorage.setItem('vana-remember-email', email);
      else localStorage.removeItem('vana-remember-email');
      return { success: true, user: newUser, token: newToken };
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || 'Login failed';
      setError(message);
      return { success: false, error: message };
    } finally {
      setLoading(false);
    }
  }, []);

  const register = useCallback(async (email: string, password: string, fullName: string) => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<AuthResponse>('/auth/register', { email, password, fullName });
      const { token: newToken, user: newUser } = response.data;
      setToken(newToken);
      setUser(newUser);
      localStorage.setItem('vana-token', newToken);
      localStorage.setItem('vana-user', JSON.stringify(newUser));
      return { success: true, user: newUser, token: newToken };
    } catch (err: any) {
      const message = err?.response?.data?.message || err.message || 'Registration failed';
      setError(message);
      return { success: false, error: message };
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('vana-token');
    localStorage.removeItem('vana-user');
    localStorage.removeItem('vana-remember-email');
  }, []);

  const isAuthenticated = !!token && !!user;

  return (
    <AuthContext.Provider value={{ user, token, loading, error, login, register, logout, isAuthenticated }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
};
