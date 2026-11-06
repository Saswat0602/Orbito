export type TaskStatus = 'todo' | 'in-progress' | 'needs-improvement' | 'completed';

export type TaskPriority = 'low' | 'medium' | 'high';

export type TaskType = 'feature' | 'bug' | 'improvement' | 'research' | 'meeting' | 'other';

export interface Task {
    id: string;
    text: string;
    description?: string;
    status: TaskStatus;
    priority: TaskPriority;
    type?: TaskType;
    week: number;
    date: string;
    tags: string[];
    comments: Comment[];
    createdAt: string;
    updatedAt: string;
}

export interface Comment {
    id: string;
    taskId: string;
    user: string;
    text: string;
    time: string;
    createdAt: string;
}

export interface StatusConfig {
    id: TaskStatus;
    label: string;
    icon: string;
    color: string;
    bg: string;
}

export interface Week {
    id: number;
    label: string;
    range: string;
}
