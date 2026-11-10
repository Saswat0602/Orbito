import type { LucideIcon } from 'lucide-react';

export interface SidebarNavItem {
  name: string;
  href: string;
  icon: LucideIcon;
  exact?: boolean;
  badge?: string | number;
}

export interface WorkspaceScope {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  projectsCount?: number;
  membersCount?: number;
}

export interface ProjectScope {
  id: string;
  name: string;
  slug: string;
  workspaceId: string;
  tasksCount?: number;
  openSprintsCount?: number;
}

