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

const TOKEN_KEY = 'vana-token';
const USER_KEY = 'vana-user';

// Attach the token on every request. Reading it per request (rather than in an
// effect) means child components can call the API on their very first render.
axios.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    config.headers = config.headers ?? {};
    (config.headers as any).Authorization = `Bearer ${token}`;
  }
  return config;
});

// Older builds stored the user with snake_case keys
const normalizeUser = (raw: any): User | null => {
  if (!raw || !raw.id) return null;
  return {
    id: raw.id,
    email: raw.email,
    fullName: raw.fullName ?? raw.full_name ?? '',
    createdAt: raw.createdAt ?? raw.created_at ?? '',
  };
};

const readStoredUser = (): User | null => {
  try {
    const saved = localStorage.getItem(USER_KEY);
    return saved ? normalizeUser(JSON.parse(saved)) : null;
  } catch {
    return null;
  }
};

const errorMessage = (err: any, fallback: string): string =>
  err?.response?.data?.message ||
  (err?.response ? fallback : 'Unable to reach VANA. Please check your connection.');

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
  const [user, setUser] = useState<User | null>(readStoredUser);

  const [token, setToken] = useState<string | null>(() => {
    return localStorage.getItem(TOKEN_KEY);
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const saveSession = (newToken: string, rawUser: any): User | null => {
    const newUser = normalizeUser(rawUser);
    localStorage.setItem(TOKEN_KEY, newToken);
    localStorage.setItem(USER_KEY, JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
    return newUser;
  };

  const login = useCallback(async (email: string, password: string, rememberMe: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<AuthResponse>('/auth/login', { email, password });
      const { token: newToken } = response.data;
      const newUser = saveSession(newToken, response.data.user);
      if (rememberMe) localStorage.setItem('vana-remember-email', email);
      else localStorage.removeItem('vana-remember-email');
      return { success: true, user: newUser ?? undefined, token: newToken };
    } catch (err: any) {
      const message = errorMessage(err, 'Invalid email or password.');
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
      const { token: newToken } = response.data;
      const newUser = saveSession(newToken, response.data.user);
      return { success: true, user: newUser ?? undefined, token: newToken };
    } catch (err: any) {
      const message = errorMessage(err, 'Registration failed. Please try again.');
      setError(message);
      return { success: false, error: message };
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    setUser(null);
    setToken(null);
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem('vana-remember-email');
  }, []);

  // An expired or revoked session on any API call signs the user out
  useEffect(() => {
    const id = axios.interceptors.response.use(undefined, (err) => {
      const url: string = err?.config?.url ?? '';
      if (err?.response?.status === 401 && !url.startsWith('/auth/')) logout();
      return Promise.reject(err);
    });
    return () => axios.interceptors.response.eject(id);
  }, [logout]);

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
