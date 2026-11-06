'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Task, TaskStatus } from '@/types';
import * as tasksApi from '../api/tasks';

export const useTasks = () => {
    return useQuery({
        queryKey: ['tasks'],
        queryFn: tasksApi.getTasks,
    });
};

export const useTask = (id: string) => {
    return useQuery({
        queryKey: ['tasks', id],
        queryFn: () => tasksApi.getTaskById(id),
        enabled: !!id,
    });
};

export const useTasksByWeek = (week: number) => {
    return useQuery({
        queryKey: ['tasks', 'week', week],
        queryFn: () => tasksApi.getTasksByWeek(week),
    });
};

export const useTasksByStatus = (status: TaskStatus) => {
    return useQuery({
        queryKey: ['tasks', 'status', status],
        queryFn: () => tasksApi.getTasksByStatus(status),
    });
};

export const useCreateTask = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: tasksApi.createTask,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tasks'] });
        },
    });
};

export const useUpdateTask = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ id, updates }: { id: string; updates: Partial<Task> }) =>
            tasksApi.updateTask(id, updates),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tasks'] });
        },
    });
};

export const useUpdateTaskStatus = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ id, status }: { id: string; status: TaskStatus }) =>
            tasksApi.updateTaskStatus(id, status),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tasks'] });
        },
    });
};

export const useDeleteTask = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: tasksApi.deleteTask,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tasks'] });
        },
    });
};
