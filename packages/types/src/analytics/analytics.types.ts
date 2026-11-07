export interface WorkspaceAnalyticsProps {
  workspaceSlug: string;
}

export interface AnalyticsData {
  projectStatus: { status: string; count: number }[];
  taskPriority: { priority: string; count: number }[];
  kpiMetrics: {
    totalTasks: number;
    completionRate: number;
    velocityScore: number;
    activeSprints: number;
  };
  taskType: { type: string; count: number }[];
  sprintStatus: { status: string; count: number }[];
  monthlyCompletion: { month: string; completed: number; created: number }[];
}

export interface Widget {
  id: string;
  title: string;
  dataKey: keyof AnalyticsData;
  visible: boolean;
  gridCols?: string;
  priority?: number;
}
