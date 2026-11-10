---
name: orbito-custom-hooks
description: >-
  Guides creation, typing, and testing of reusable custom React hooks in the Orbito frontend, focusing on state separation, performance, lifecycle safety, and clean interfaces.
---

# Orbito Custom Hooks Engineering Guide

This skill governs the architecture, contracts, and lifecycle patterns for custom React hooks located in `frontend/src/lib/hooks/`.

---

## 1. Core Principles for Custom Hooks

1. **Separation of Presentation & Behavior**: Extract complex side effects, polling, window listeners, and Redux selections into custom hooks. Keep UI components pure and declarative.
2. **Explicit Return Contracts**: Always declare typed return interfaces or tuple contracts with `as const`. Never return untyped anonymous objects.
3. **Zero `any` Policy**: Use generics `<T>` or TypeScript discrimination.
4. **Lifecycle & Cleanup Hygiene**: Always unregister event listeners, clear intervals/timeouts, and abort pending fetches in the `useEffect` cleanup return.

---

## 2. Standard Hook Templates in Orbito

### 2.1 RBAC & Identity Hook (`useAuthRbac`)
```typescript
import { useAppSelector } from '@/lib/store/hooks';
import { Role } from '@orbito/types';

export interface AuthRbacResult {
  user: any;
  isAuthenticated: boolean;
  role: Role | null;
  isSuperAdmin: boolean;
  isOwner: boolean;
  isManager: boolean;
  isMember: boolean;
  isViewer: boolean;
  canManageWorkspace: boolean;
  canManageProject: boolean;
  canWriteTasks: boolean;
  roleBadge: { label: string; variant: string };
}

export function useAuthRbac(): AuthRbacResult {
  const { user, isAuthenticated } = useAppSelector((state) => state.auth);
  const role = user?.role as Role | null;

  const isSuperAdmin = role === Role.SUPER_ADMIN;
  const isOwner = role === Role.OWNER || isSuperAdmin;
  const isManager = role === Role.MANAGER || isOwner;
  const isMember = role === Role.MEMBER || isManager;
  const isViewer = role === Role.VIEWER;

  return {
    user,
    isAuthenticated,
    role,
    isSuperAdmin,
    isOwner,
    isManager,
    isMember,
    isViewer,
    canManageWorkspace: isOwner || isSuperAdmin,
    canManageProject: isManager,
    canWriteTasks: !isViewer,
    roleBadge: {
      label: role ? role.replace('_', ' ') : 'Guest',
      variant: isSuperAdmin ? 'destructive' : isOwner ? 'default' : 'secondary',
    },
  };
}
```

### 2.2 Debounce Hook (`useDebounce`)
Used for fast search filters, command palettes, and auto-saving task titles:
```typescript
import { useState, useEffect } from 'react';

export function useDebounce<T>(value: T, delayMs: number = 300): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delayMs);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delayMs]);

  return debouncedValue;
}
```

### 2.3 Keydown & Shortcut Listener (`useKeyShortcut`)
Used for global shortcuts like `⌘K` command palette or `Escape` dismissal:
```typescript
import { useEffect } from 'react';

export function useKeyShortcut(
  targetKey: string,
  callback: (e: KeyboardEvent) => void,
  metaKeyRequired: boolean = false
) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isTargetKey = e.key.toLowerCase() === targetKey.toLowerCase();
      const isMetaSatisfied = metaKeyRequired ? e.metaKey || e.ctrlKey : true;

      if (isTargetKey && isMetaSatisfied) {
        callback(e);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [targetKey, callback, metaKeyRequired]);
}
```

---

## 3. Anti-Patterns to Avoid

- **No Duplicated State**: Do not mirror RTK Query cache in `useState` unless editing a temporary draft.
- **No Unstable Dependency Arrays**: Memoize callbacks with `useCallback` or primitive IDs when passing them into `useEffect` dependencies.
- **No Catch-All File Dumping**: Every hook should reside in its own dedicated file (e.g., `src/lib/hooks/useDebounce.ts`) and be exported cleanly from `src/lib/hooks/index.ts`.

