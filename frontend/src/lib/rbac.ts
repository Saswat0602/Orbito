import { Role } from '@orbito/types';

export { Role };

export const ROLE_HIERARCHY: Record<Role, number> = {
  [Role.SUPER_ADMIN]: 100,
  [Role.OWNER]: 80,
  [Role.MANAGER]: 60,
  [Role.MEMBER]: 40,
  [Role.VIEWER]: 20,
};

export type AppAction =
  | 'view'
  | 'create'
  | 'edit'
  | 'delete'
  | 'manage'
  | 'admin';

export type AppResource =
  | 'tasks'
  | 'sprints'
  | 'projects'
  | 'workspaces'
  | 'organizations'
  | 'members'
  | 'settings'
  | 'billing'
  | 'automations'
  | 'integrations'
  | 'inbox'
  | 'admin_panel';

/**
 * Checks whether userRole is at or above the requiredRole in the hierarchy.
 */
export function hasRole(
  userRole: Role | string | undefined | null,
  requiredRole: Role
): boolean {
  if (!userRole) return false;
  const userRank = ROLE_HIERARCHY[userRole as Role] || 0;
  const requiredRank = ROLE_HIERARCHY[requiredRole] || 0;
  return userRank >= requiredRank;
}

/**
 * Checks granular permissions for a given role, action, and resource.
 */
export function can(
  userRole: Role | string | undefined | null,
  action: AppAction,
  resource: AppResource
): boolean {
  if (!userRole) return false;
  const role = userRole as Role;

  // Super admin has unrestricted permission
  if (role === Role.SUPER_ADMIN) return true;

  switch (resource) {
    case 'admin_panel':
      return false;

    case 'organizations':
    case 'billing':
      return role === Role.OWNER;

    case 'settings':
    case 'integrations':
      if (action === 'view') return true;
      return role === Role.OWNER || role === Role.MANAGER;

    case 'automations':
    case 'members':
    case 'workspaces':
      if (action === 'view') return true;
      return role === Role.OWNER || role === Role.MANAGER;

    case 'projects':
      if (action === 'view') return true;
      if (action === 'create' || action === 'edit')
        return role === Role.OWNER || role === Role.MANAGER;
      if (action === 'delete') return role === Role.OWNER;
      return false;

    case 'sprints':
      if (action === 'view') return true;
      return (
        role === Role.OWNER || role === Role.MANAGER || role === Role.MEMBER
      );

    case 'inbox':
      if (action === 'view') return true;
      return (
        role === Role.OWNER || role === Role.MANAGER || role === Role.MEMBER
      );

    case 'tasks':
      if (action === 'view') return true;
      if (action === 'create' || action === 'edit') {
        return (
          role === Role.OWNER || role === Role.MANAGER || role === Role.MEMBER
        );
      }
      if (action === 'delete') {
        return role === Role.OWNER || role === Role.MANAGER;
      }
      return false;

    default:
      return false;
  }
}

/**
 * Returns a styling token for role badges.
 */
export function getRoleBadgeStyle(role: Role | string | undefined | null): {
  label: string;
  className: string;
} {
  switch (role) {
    case Role.SUPER_ADMIN:
      return {
        label: 'Super Admin',
        className: 'bg-zinc-800/90 border-zinc-700 text-zinc-100 font-mono tracking-wide',
      };
    case Role.OWNER:
      return {
        label: 'Owner',
        className: 'bg-zinc-850 border-zinc-750 text-zinc-200 font-medium',
      };
    case Role.MANAGER:
      return {
        label: 'Manager',
        className: 'bg-zinc-900 border-zinc-800 text-zinc-300 font-medium',
      };
    case Role.MEMBER:
      return {
        label: 'Member',
        className: 'bg-zinc-900/80 border-zinc-800 text-zinc-400 font-medium',
      };
    case Role.VIEWER:
    default:
      return {
        label: 'Viewer',
        className: 'bg-zinc-950 border-zinc-850 text-zinc-500 font-medium',
      };
  }
}

