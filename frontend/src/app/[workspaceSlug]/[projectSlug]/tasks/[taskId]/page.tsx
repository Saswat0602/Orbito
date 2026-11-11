'use client';

import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { TaskDetailWorkspace } from '@/features/tasks/components/TaskDetailWorkspace';

export default function FullPageTaskView() {
  const params = useParams<{
    workspaceSlug: string;
    projectSlug: string;
    taskId: string;
  }>();
  const router = useRouter();

  const workspaceSlug = params?.workspaceSlug || '';
  const projectSlug = params?.projectSlug || '';
  const taskId = params?.taskId || '';

  if (!taskId) return null;

  return (
    <div className="flex-1 overflow-y-auto bg-[#09090b]">
      <TaskDetailWorkspace
        taskId={taskId}
        workspaceSlug={workspaceSlug}
        projectSlug={projectSlug}
        mode="fullpage"
        onClose={() => router.push(`/${workspaceSlug}/${projectSlug}/tasks`)}
      />
    </div>
  );
}

