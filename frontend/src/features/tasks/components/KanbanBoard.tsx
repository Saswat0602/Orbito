'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragStartEvent,
  DragEndEvent,
  DragOverEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Plus, MoreHorizontal } from 'lucide-react';
import type { Task, TaskPriority } from '@orbito/types';
import { TaskCard } from './TaskCard';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  useUpdateTaskMutation,
  useReorderTaskMutation,
} from '@/lib/store/api/tasksApi';
import { toast } from 'sonner';

export const KANBAN_COLUMNS = [
  { key: 'TODO', label: 'To Do', color: 'border-zinc-700 text-zinc-300' },
  { key: 'IN_PROGRESS', label: 'In Progress', color: 'border-amber-500/60 text-amber-400' },
  { key: 'IN_REVIEW', label: 'In Review', color: 'border-sky-500/60 text-sky-400' },
  { key: 'DONE', label: 'Done', color: 'border-emerald-500/60 text-emerald-400' },
] as const;

interface KanbanColumnProps {
  id: string;
  label: string;
  color: string;
  tasks: Task[];
  onAddTask?: (columnKey: string) => void;
}

function KanbanColumn({ id, label, color, tasks, onAddTask }: KanbanColumnProps) {
  const { setNodeRef } = useSortable({
    id,
    data: {
      type: 'Column',
      columnKey: id,
    },
  });

  const taskIds = useMemo(() => tasks.map((t) => t.id), [tasks]);

  return (
    <div
      ref={setNodeRef}
      className="flex flex-col flex-1 min-w-[280px] max-w-[360px] rounded-xl border border-zinc-850 bg-zinc-950/40 p-3 h-full"
    >
      {/* Column Header */}
      <div className="flex items-center justify-between pb-3 px-1 border-b border-zinc-850 mb-3">
        <div className="flex items-center gap-2">
          <span className={`h-2 w-2 rounded-full border ${color}`} />
          <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-300">
            {label}
          </h3>
          <Badge variant="outline" className="text-[10px] px-1.5 py-0 font-mono text-zinc-400 border-zinc-800">
            {tasks.length}
          </Badge>
        </div>

        {onAddTask && (
          <button
            type="button"
            onClick={() => onAddTask(id)}
            className="p-1 rounded text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800 transition-colors cursor-pointer"
            title={`Add task to ${label}`}
          >
            <Plus className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Column Task List */}
      <SortableContext items={taskIds} strategy={verticalListSortingStrategy}>
        <div className="flex-1 overflow-y-auto space-y-2.5 pr-1 min-h-[150px]">
          {tasks.map((task) => (
            <TaskCard key={task.id} task={task} />
          ))}

          {tasks.length === 0 && (
            <div className="h-28 flex flex-col items-center justify-center rounded-lg border border-dashed border-zinc-850 p-4 text-center">
              <span className="text-[11px] text-zinc-600">No tasks in {label}</span>
              {onAddTask && (
                <button
                  type="button"
                  onClick={() => onAddTask(id)}
                  className="mt-2 text-[10px] text-zinc-400 hover:text-zinc-200 underline cursor-pointer"
                >
                  Create one
                </button>
              )}
            </div>
          )}
        </div>
      </SortableContext>
    </div>
  );
}

interface KanbanBoardProps {
  tasks: Task[];
  projectId: string;
  onAddTask?: (columnKey: string) => void;
}

export function KanbanBoard({ tasks, projectId, onAddTask }: KanbanBoardProps) {
  const [activeTask, setActiveTask] = useState<Task | null>(null);

  const [updateTask] = useUpdateTaskMutation();
  const [reorderTask] = useReorderTaskMutation();

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5, // 5px movement required before drag begins
      },
    }),
    useSensor(KeyboardSensor)
  );

  // Group tasks by category
  const columns = useMemo(() => {
    const map: Record<string, Task[]> = {
      TODO: [],
      IN_PROGRESS: [],
      IN_REVIEW: [],
      DONE: [],
    };

    tasks.forEach((t) => {
      const cat = t.status?.category || 'TODO';
      if (map[cat]) {
        map[cat].push(t);
      } else {
        map.TODO.push(t);
      }
    });

    return map;
  }, [tasks]);

  const handleDragStart = (event: DragStartEvent) => {
    const { active } = event;
    const task = tasks.find((t) => t.id === active.id);
    if (task) {
      setActiveTask(task);
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveTask(null);

    if (!over) return;

    const activeId = active.id as string;
    const overId = over.id as string;

    const currentTask = tasks.find((t) => t.id === activeId);
    if (!currentTask) return;

    // Check if dropped onto a column directly or onto another task
    let targetCategory = '';
    const isOverColumn = KANBAN_COLUMNS.some((col) => col.key === overId);

    if (isOverColumn) {
      targetCategory = overId;
    } else {
      const overTask = tasks.find((t) => t.id === overId);
      if (overTask) {
        targetCategory = overTask.status?.category || 'TODO';
      }
    }

    const sourceCategory = currentTask.status?.category || 'TODO';

    // 1. If moved to a different column/status:
    if (targetCategory && targetCategory !== sourceCategory) {
      try {
        await updateTask({
          id: activeId,
          updates: {
            // Updating status category directly
            status: { ...currentTask.status, category: targetCategory as any } as any,
          },
        }).unwrap();
        toast.success(`Task moved to ${targetCategory}`);
      } catch {
        toast.error('Failed to update task status');
      }
    }

    // 2. Reorder LexoRank positioning:
    const targetColumnTasks = columns[targetCategory || sourceCategory] || [];
    const overIndex = targetColumnTasks.findIndex((t) => t.id === overId);

    if (overIndex !== -1) {
      const beforeTaskId = overIndex > 0 ? targetColumnTasks[overIndex - 1].id : null;
      const afterTaskId = overIndex < targetColumnTasks.length - 1 ? targetColumnTasks[overIndex + 1].id : null;

      try {
        await reorderTask({
          taskId: activeId,
          payload: {
            scopeType: 'PROJECT' as any,
            scopeId: projectId,
            viewType: 'BOARD' as any,
            afterTaskId,
            beforeTaskId,
          },
        }).unwrap();
      } catch {
        // Fallback silently if server recalculates
      }
    }
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex-1 flex gap-4 overflow-x-auto pb-4 items-start select-none">
        {KANBAN_COLUMNS.map((col) => (
          <KanbanColumn
            key={col.key}
            id={col.key}
            label={col.label}
            color={col.color}
            tasks={columns[col.key] || []}
            onAddTask={onAddTask}
          />
        ))}
      </div>

      <DragOverlay>
        {activeTask ? <TaskCard task={activeTask} isOverlay /> : null}
      </DragOverlay>
    </DndContext>
  );
}

