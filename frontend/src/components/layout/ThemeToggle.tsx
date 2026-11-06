'use client';

import React from 'react';
import { useTheme } from '@/lib/context/ThemeContext';
import { Moon, Sun } from 'lucide-react';

export const ThemeToggle: React.FC = () => {
    const { theme, toggleTheme } = useTheme();

    return (
        <button
            onClick={toggleTheme}
            className="p-2.5 rounded-xl bg-[rgb(var(--color-bg-tertiary))] hover:bg-[rgb(var(--color-primary))] hover:text-white transition-all duration-200 group"
            aria-label="Toggle theme"
        >
            {theme === 'light' ? (
                <Moon size={18} className="transition-transform group-hover:rotate-12" />
            ) : (
                <Sun size={18} className="transition-transform group-hover:rotate-12" />
            )}
        </button>
    );
};
