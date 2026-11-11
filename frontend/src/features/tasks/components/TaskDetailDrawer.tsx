'use client';

import * as React from 'react';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { TaskDetailWorkspace } from './TaskDetailWorkspace';

export function TaskDetailDrawer() {
  const dispatch = useAppDispatch();
  const taskDrawerId = useAppSelector((state) => state.ui.taskDrawerId);
  const activeWorkspaceSlug = useAppSelector((state) => state.auth.activeWorkspaceSlug);
  const activeProjectSlug = useAppSelector((state) => state.auth.activeProjectSlug);

  if (!taskDrawerId) return null;

  const handleClose = () => {
    dispatch(setTaskDrawerId(null));
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity"
        onClick={handleClose}
      />

      {/* Drawer Container */}
      <div className="relative w-full max-w-4xl bg-zinc-900 border-l border-zinc-800 h-full flex flex-col shadow-2xl z-10 overflow-hidden">
        <TaskDetailWorkspace
          taskId={taskDrawerId}
          workspaceSlug={activeWorkspaceSlug || undefined}
          projectSlug={activeProjectSlug || undefined}
          mode="drawer"
          onClose={handleClose}
        />
      </div>
    </div>
  );
}
