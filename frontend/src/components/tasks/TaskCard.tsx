'use client';

import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Task } from '@/types';
import { MessageSquare, GripVertical } from 'lucide-react';
import { PriorityDot } from '../ui/PriorityBadge';

interface TaskCardProps {
    task: Task;
    onClick?: () => void;
    isDragging?: boolean;
}

export const TaskCard: React.FC<TaskCardProps> = ({
    task,
    onClick,
    isDragging: isDraggingProp,
}) => {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id: task.id });

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging || isDraggingProp ? 0.5 : 1,
    };

    return (
        <div
            ref={setNodeRef}
            style={style}
            className="group bg-[rgb(var(--color-bg-primary))] p-4 md:p-5 rounded-2xl border border-[rgb(var(--color-border-primary))] shadow-custom-sm hover:border-[rgb(var(--color-primary))] hover:shadow-custom-xl transition-all cursor-pointer relative touch-manipulation"
            onClick={onClick}
        >
            {/* Drag Handle */}
            <div
                {...attributes}
                {...listeners}
                className="absolute top-2 right-2 p-2 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab active:cursor-grabbing touch-none"
                onClick={(e) => e.stopPropagation()}
            >
                <GripVertical size={16} className="text-[rgb(var(--color-text-tertiary))]" />
            </div>

            <div className="flex justify-between items-start mb-3 md:mb-4 pr-8">
                <p className="text-sm font-bold text-[rgb(var(--color-text-primary))] leading-snug flex-1">
                    {task.text}
                </p>
                <div className="shrink-0 ml-2">
                    <PriorityDot priority={task.priority} />
                </div>
            </div>

            {task.description && (
                <p className="text-xs text-[rgb(var(--color-text-secondary))] mb-3 line-clamp-2">
                    {task.description}
                </p>
            )}

            {/* Tags */}
            {task.tags && task.tags.length > 0 && (
                <div className="flex flex-wrap gap-1.5 mb-3">
                    {task.tags.slice(0, 3).map((tag, index) => (
                        <span
                            key={index}
                            className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))]"
                        >
                            {tag}
                        </span>
                    ))}
                    {task.tags.length > 3 && (
                        <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-[rgb(var(--color-bg-tertiary))] text-[rgb(var(--color-text-tertiary))]">
                            +{task.tags.length - 3}
                        </span>
                    )}
                </div>
            )}

            <div className="flex items-center justify-between pt-3 border-t border-[rgb(var(--color-border-secondary))]">
                <div className="flex items-center gap-1.5 text-[10px] text-[rgb(var(--color-text-tertiary))] font-bold bg-[rgb(var(--color-bg-tertiary))] px-2 py-1 rounded-lg">
                    <MessageSquare size={12} /> {task.comments?.length || 0}
                </div>

                <span className="text-[10px] font-bold text-[rgb(var(--color-text-tertiary))] uppercase">
                    {task.type || 'Task'}
                </span>
            </div>
        </div>
    );
};
