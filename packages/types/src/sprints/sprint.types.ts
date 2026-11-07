import type { SprintStatus } from '../enums';

export interface SprintSummary {
  id: string;
  name: string;
  slug?: string;
  status: SprintStatus | string;
  projectId: string;
}

export interface Sprint extends SprintSummary {
  goal?: string | null;
  startDate?: string | null;
  endDate?: string | null;
  completedAt?: string | null;
  totalPoints?: number;
  completedPoints?: number;
  createdAt: string;
  updatedAt: string;
}

export interface SprintBurndownPoint {
  date: string;
  idealRemaining: number;
  actualRemaining: number;
}

export interface StartSprintPayload {
  startDate?: string;
  endDate?: string;
  goal?: string;
}

export interface CompleteSprintPayload {
  moveToSprintId?: string;
}
