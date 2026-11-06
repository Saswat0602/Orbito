'use client';

import React, { useState } from 'react';
import {
    DndContext,
    DragEndEvent,
    DragOverlay,
    DragStartEvent,
    PointerSensor,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { Task, TaskStatus } from '@/types';
import { STATUSES } from '@/lib/utils/constants';
import { StatusColumn } from './StatusColumn';
import { TaskCard } from './TaskCard';

interface TaskBoardProps {
    tasks: Task[];
    onTaskClick: (task: Task) => void;
    onStatusChange: (taskId: string, status: TaskStatus) => void;
    onAddTask: (status: TaskStatus) => void;
}

export const TaskBoard: React.FC<TaskBoardProps> = ({
    tasks,
    onTaskClick,
    onStatusChange,
    onAddTask,
}) => {
    const [activeTask, setActiveTask] = useState<Task | null>(null);

    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: {
                distance: 8,
            },
        })
    );

    const handleDragStart = (event: DragStartEvent) => {
        const task = tasks.find((t) => t.id === event.active.id);
        if (task) {
            setActiveTask(task);
        }
    };

    const handleDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;

        if (over && active.id !== over.id) {
            const taskId = active.id as string;
            const newStatus = over.id as TaskStatus;

            // Check if dropped over a status column
            if (STATUSES.some((s) => s.id === newStatus)) {
                onStatusChange(taskId, newStatus);
            }
        }

        setActiveTask(null);
    };

    const getTasksByStatus = (status: TaskStatus) => {
        return tasks.filter((task) => task.status === status);
    };

    return (
        <DndContext
            sensors={sensors}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
        >
            <div className="flex gap-4 md:gap-6 overflow-x-auto pb-6 md:pb-10 scrollbar-thin min-h-0">
                {STATUSES.map((status) => (
                    <StatusColumn
                        key={status.id}
                        status={status}
                        tasks={getTasksByStatus(status.id)}
                        onTaskClick={onTaskClick}
                        onAddTask={() => onAddTask(status.id)}
                    />
                ))}
            </div>

            <DragOverlay>
                {activeTask ? <TaskCard task={activeTask} isDragging /> : null}
            </DragOverlay>
        </DndContext>
    );
};
