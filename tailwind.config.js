/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        heading: ['Outfit', 'Inter', 'system-ui', 'sans-serif'],
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'ui-monospace', 'monospace'],
      },
      colors: {
        obsidian: {
          DEFAULT: '#000000',
          50:  '#09090b',
          100: '#0c0c0f',
          200: '#111115',
          300: '#18181b',
          400: '#27272a',
          500: '#3f3f46',
          600: '#52525b',
          700: '#71717a',
          800: '#a1a1aa',
          900: '#d4d4d8',
          950: '#fafafa',
        },
        alabaster: {
          DEFAULT: '#fafafa',
          50:  '#ffffff',
          100: '#fafafa',
          200: '#f4f4f5',
          300: '#e4e4e7',
          400: '#d4d4d8',
          500: '#a1a1aa',
          600: '#71717a',
          700: '#52525b',
          800: '#3f3f46',
          900: '#27272a',
          950: '#09090b',
        },
      },
      boxShadow: {
        ambient:    '0 1px 2px rgba(0,0,0,0.7), 0 4px 8px rgba(0,0,0,0.4), 0 16px 32px rgba(0,0,0,0.25), 0 32px 64px rgba(0,0,0,0.12)',
        glass:      '0 8px 32px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.05)',
        lift:       '0 16px 64px rgba(0,0,0,0.6), 0 0 0 1px rgba(139,92,246,0.25), 0 0 40px rgba(139,92,246,0.08)',
        'glow-violet':  '0 0 12px rgba(139,92,246,0.6), 0 0 32px rgba(139,92,246,0.2)',
        'glow-emerald': '0 0 12px rgba(52,211,153,0.5),  0 0 32px rgba(52,211,153,0.15)',
        'glow-slate':   '0 0 10px rgba(148,163,184,0.35)',
        'card-light':   '0 1px 3px rgba(0,0,0,0.06), 0 4px 16px rgba(0,0,0,0.04)',
        'lift-light':   '0 8px 32px rgba(0,0,0,0.1), 0 0 0 1px rgba(124,58,237,0.15)',
      },
      animation: {
        shimmer:      'shimmer 2.4s linear infinite',
        'pulse-glow': 'pulse-glow 3s ease-in-out infinite',
        'fade-up':    'fade-up 0.4s ease forwards',
        float:        'float 6s ease-in-out infinite',
        breath:       'breath 4s ease-in-out infinite',
      },
      keyframes: {
        shimmer: {
          '0%':   { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(200%)' },
        },
        'pulse-glow': {
          '0%, 100%': { opacity: '1',   boxShadow: '0 0 12px rgba(139,92,246,0.6)' },
          '50%':      { opacity: '0.7', boxShadow: '0 0 24px rgba(139,92,246,0.9)' },
        },
        'fade-up': {
          from: { opacity: '0', transform: 'translateY(10px)' },
          to:   { opacity: '1', transform: 'translateY(0)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%':      { transform: 'translateY(-10px)' },
        },
        breath: {
          '0%, 100%': { transform: 'scale(1)', opacity: '0.8' },
          '50%':      { transform: 'scale(1.08)', opacity: '1' },
        },
      },
      backgroundImage: {
        'glow-radial': 'radial-gradient(ellipse 60% 40% at 50% 0%, rgba(139,92,246,0.15) 0%, transparent 70%)',
        'glow-radial-light': 'radial-gradient(ellipse 60% 40% at 50% 0%, rgba(124,58,237,0.07) 0%, transparent 70%)',
      },
      transitionTimingFunction: {
        'spring-out': 'cubic-bezier(0.22, 1, 0.36, 1)',
      },
    },
  },
  plugins: [],
};
