import { Task, TaskStatus } from '@/types';
import { storage, STORAGE_KEYS } from '../utils/storage';

// Mock data for initial state
const MOCK_TASKS: Task[] = [
    {
        id: '1',
        text: 'Finalize UI Mockups',
        description: 'Complete the final design mockups for the dashboard',
        status: 'in-progress',
        priority: 'high',
        type: 'feature',
        week: 1,
        date: '2026-01-03',
        tags: ['design', 'ui'],
        comments: [],
        createdAt: new Date('2026-01-01').toISOString(),
        updatedAt: new Date('2026-01-02').toISOString(),
    },
    {
        id: '2',
        text: 'Monthly Budget Review',
        description: 'Review and analyze monthly budget allocations',
        status: 'completed',
        priority: 'medium',
        type: 'other',
        week: 1,
        date: '2026-01-05',
        tags: ['finance'],
        comments: [],
        createdAt: new Date('2026-01-01').toISOString(),
        updatedAt: new Date('2026-01-05').toISOString(),
    },
];

// Initialize storage with mock data if empty
const initializeTasks = (): Task[] => {
    const stored = storage.get<Task[]>(STORAGE_KEYS.TASKS);
    if (!stored || stored.length === 0) {
        storage.set(STORAGE_KEYS.TASKS, MOCK_TASKS);
        return MOCK_TASKS;
    }
    return stored;
};

// API Functions (currently using localStorage, ready for backend integration)
export const getTasks = async (): Promise<Task[]> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 300));
    return initializeTasks();
};

export const getTaskById = async (id: string): Promise<Task | null> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const tasks = initializeTasks();
    return tasks.find(task => task.id === id) || null;
};

export const createTask = async (taskData: Omit<Task, 'id' | 'createdAt' | 'updatedAt'>): Promise<Task> => {
    await new Promise(resolve => setTimeout(resolve, 300));

    const newTask: Task = {
        ...taskData,
        id: Date.now().toString(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    };

    const tasks = initializeTasks();
    const updatedTasks = [newTask, ...tasks];
    storage.set(STORAGE_KEYS.TASKS, updatedTasks);

    return newTask;
};

export const updateTask = async (id: string, updates: Partial<Task>): Promise<Task> => {
    await new Promise(resolve => setTimeout(resolve, 300));

    const tasks = initializeTasks();
    const taskIndex = tasks.findIndex(task => task.id === id);

    if (taskIndex === -1) {
        throw new Error(`Task with id ${id} not found`);
    }

    const updatedTask: Task = {
        ...tasks[taskIndex],
        ...updates,
        updatedAt: new Date().toISOString(),
    };

    tasks[taskIndex] = updatedTask;
    storage.set(STORAGE_KEYS.TASKS, tasks);

    return updatedTask;
};

export const updateTaskStatus = async (id: string, status: TaskStatus): Promise<Task> => {
    return updateTask(id, { status });
};

export const deleteTask = async (id: string): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 300));

    const tasks = initializeTasks();
    const filteredTasks = tasks.filter(task => task.id !== id);
    storage.set(STORAGE_KEYS.TASKS, filteredTasks);
};

export const getTasksByWeek = async (week: number): Promise<Task[]> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const tasks = initializeTasks();
    return tasks.filter(task => task.week === week);
};

export const getTasksByStatus = async (status: TaskStatus): Promise<Task[]> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const tasks = initializeTasks();
    return tasks.filter(task => task.status === status);
};
