import { STORAGE_KEYS } from './constants';

export const storage = {
    get: <T>(key: string): T | null => {
        if (typeof window === 'undefined') return null;

        try {
            const item = window.localStorage.getItem(key);
            return item ? JSON.parse(item) : null;
        } catch (error) {
            console.error(`Error reading from localStorage:`, error);
            return null;
        }
    },

    set: <T>(key: string, value: T): void => {
        if (typeof window === 'undefined') return;

        try {
            window.localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error(`Error writing to localStorage:`, error);
        }
    },

    remove: (key: string): void => {
        if (typeof window === 'undefined') return;

        try {
            window.localStorage.removeItem(key);
        } catch (error) {
            console.error(`Error removing from localStorage:`, error);
        }
    },

    clear: (): void => {
        if (typeof window === 'undefined') return;

        try {
            window.localStorage.clear();
        } catch (error) {
            console.error(`Error clearing localStorage:`, error);
        }
    },
};

// Helper functions for specific storage operations
export const getTasks = () => storage.get(STORAGE_KEYS.TASKS);
export const setTasks = (tasks: any) => storage.set(STORAGE_KEYS.TASKS, tasks);

export const getTags = () => storage.get(STORAGE_KEYS.TAGS);
export const setTags = (tags: any) => storage.set(STORAGE_KEYS.TAGS, tags);

export const getTheme = () => storage.get(STORAGE_KEYS.THEME);
export const setTheme = (theme: string) => storage.set(STORAGE_KEYS.THEME, theme);

// Re-export STORAGE_KEYS for convenience
export { STORAGE_KEYS };
