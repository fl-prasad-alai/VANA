/**
 * Theme System - VANA Emerald Moss
 * Three modes: Greening, Dark, Light
 */

export type ThemeMode = 'greening' | 'dark' | 'light';

export interface ThemeColors {
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  surface: string;
  text: string;
  textSecondary: string;
  border: string;
  success: string;
  warning: string;
  error: string;
  critical: string;
}

export const THEMES: Record<ThemeMode, ThemeColors> = {
  greening: {
    // Nature-inspired: Lush green, calming moss tones
    primary: '#2D5A27', // Deep moss
    secondary: '#7AA973', // Sage green
    accent: '#A8D5BA', // Mint
    background: '#F5FAF7', // Off-white with green tint
    surface: '#FFFFFF', // White
    text: '#1A1A1A', // Almost black
    textSecondary: '#666666', // Gray
    border: '#E0E8E3', // Light green border
    success: '#22C55E', // Green
    warning: '#F59E0B', // Amber
    error: '#EF4444', // Red
    critical: '#DC2626', // Dark red
  },
  dark: {
    // Dark mode: Deep grays with emerald accents
    primary: '#1E3A2F', // Deep forest
    secondary: '#4A7C59', // Muted green
    accent: '#6FA86F', // Medium green
    background: '#0F1419', // Almost black
    surface: '#1A1F26', // Dark gray
    text: '#E5E7EB', // Light gray
    textSecondary: '#9CA3AF', // Medium gray
    border: '#374151', // Dark border
    success: '#10B981', // Emerald
    warning: '#F59E0B', // Amber
    error: '#EF4444', // Red
    critical: '#DC2626', // Dark red
  },
  light: {
    // Light mode: Clean, bright with emerald touches
    primary: '#047857', // Emerald
    secondary: '#06B6D4', // Cyan
    accent: '#8B5CF6', // Purple
    background: '#FFFFFF', // White
    surface: '#F3F4F6', // Light gray
    text: '#1F2937', // Dark gray
    textSecondary: '#6B7280', // Medium gray
    border: '#E5E7EB', // Light border
    success: '#10B981', // Green
    warning: '#F59E0B', // Amber
    error: '#EF4444', // Red
    critical: '#DC2626', // Dark red
  },
};

export const getCSSVariables = (theme: ThemeColors): Record<string, string> => {
  return {
    '--color-primary': theme.primary,
    '--color-secondary': theme.secondary,
    '--color-accent': theme.accent,
    '--color-background': theme.background,
    '--color-surface': theme.surface,
    '--color-text': theme.text,
    '--color-text-secondary': theme.textSecondary,
    '--color-border': theme.border,
    '--color-success': theme.success,
    '--color-warning': theme.warning,
    '--color-error': theme.error,
    '--color-critical': theme.critical,
  };
};
