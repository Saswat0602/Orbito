/**
 * @orbito/ui
 * Reusable monochrome design system components (Linear / Slack / Monday aesthetic)
 */

export const ORBITO_THEME = {
  colors: {
    bgDark: '#09090b',       // Obsidian Black
    bgDarkElevated: '#18181b', // Dark zinc elevation
    borderDark: '#27272a',
    textDark: '#fafafa',
    textMutedDark: '#a1a1aa',

    bgLight: '#ffffff',
    bgLightElevated: '#f4f4f5',
    borderLight: '#e4e4e7',
    textLight: '#09090b',
    textMutedLight: '#71717a',

    accent: '#3b82f6',
    success: '#10b981',
    warning: '#f59e0b',
    danger: '#ef4444',
  },
} as const;

export const FOCUSFLOW_THEME = ORBITO_THEME;
export type OrbitoTheme = typeof ORBITO_THEME;
export type FocusFlowTheme = OrbitoTheme;
