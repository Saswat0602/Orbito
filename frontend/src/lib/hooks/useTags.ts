'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Tag } from '@/types';
import * as tagsApi from '../api/tags';

export const useTags = () => {
    return useQuery({
        queryKey: ['tags'],
        queryFn: tagsApi.getTags,
    });
};

export const useCreateTag = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: tagsApi.createTag,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tags'] });
        },
    });
};

export const useDeleteTag = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: tagsApi.deleteTag,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tags'] });
        },
    });
};
