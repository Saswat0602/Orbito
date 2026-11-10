export interface ProjectKPIMetrics {
  totalTasks: number;
  completedTasks: number;
  activeSprints: number;
  totalBugs: number;
  resolvedBugs: number;
  completionRate: number;
  bugResolutionRate: number;
}

export interface TaskStatusFlow {
  statusId: string;
  count: number;
  status: {
    id: string;
    name: string;
    color: string;
    category: string;
    position: number;
  };
}

export interface SprintVelocity {
  id: string;
  name: string;
  startDate: string | null;
  endDate: string | null;
  velocity: number;
}

export interface TaskTypeCount {
  type: string;
  _count: {
    type: number;
  };
}

export interface TaskPriorityCount {
  priority: string;
  _count: {
    priority: number;
  };
}

export interface ProjectChartDataResponse {
  'kpi-metrics'?: ProjectKPIMetrics;
  'task-status'?: TaskStatusFlow[];
  'sprint-velocity'?: SprintVelocity[];
  'task-type'?: TaskTypeCount[];
  'task-priority'?: TaskPriorityCount[];
  [key: string]: unknown;
}

export interface SprintBurndownTask {
  completedAt: string | null;
  storyPoints: number | null;
  createdAt: string;
}
