import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';

export type Theme = 'dark' | 'greening' | 'light';

interface ThemeContextValue {
  theme: Theme;
  setTheme: (t: Theme) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

const STORAGE_KEY = 'vana_theme_v1';

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>('greening');

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
    const initial = stored && ['dark', 'greening', 'light'].includes(stored) ? stored : 'greening';
    applyTheme(initial);
    setThemeState(initial);
  }, []);

  const setTheme = useCallback((t: Theme) => {
    applyTheme(t);
    setThemeState(t);
    localStorage.setItem(STORAGE_KEY, t);
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function applyTheme(t: Theme) {
  const root = document.documentElement;
  root.classList.remove('dark', 'greening', 'light');
  root.classList.add(t);
  if (t === 'greening' || t === 'dark') {
    root.classList.add('dark');
  }
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used inside ThemeProvider');
  return ctx;
}
