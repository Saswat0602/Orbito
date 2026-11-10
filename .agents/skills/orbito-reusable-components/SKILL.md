---
name: orbito-reusable-components
description: >-
  Guides construction, styling, and composition of reusable UI components in the Orbito frontend using Radix UI primitives, Lucide icons, and the Obsidian monochrome design system.
---

# Orbito Reusable UI Components & Design System Runbook

This skill defines the technical contracts, styling tokens, and composition patterns for authoring and reusing UI components in `frontend/src/components/ui/` and feature views.

---

## 1. Design System Tokens & Obsidian Aesthetic

All UI elements in Orbito adhere strictly to an enterprise-grade Obsidian monochrome aesthetic:

| Token Category | CSS / Tailwind Class | Intended Surface |
| :--- | :--- | :--- |
| **Canvas Background** | `bg-[#09090b]` | Top-level viewport, app shell body |
| **Card / Panel** | `bg-zinc-900/90` or `bg-zinc-900` | Floating surfaces, cards, modals, popovers |
| **Subtle Hover** | `hover:bg-zinc-800/80` or `hover:bg-zinc-850` | Interactive list items, table rows, secondary buttons |
| **Borders** | `border-zinc-800` or `border-zinc-800/80` | Clean 1px geometric borders |
| **Radii** | `rounded-md` (6px) or `rounded-sm` (4px) | **Strictly 4-6px**. Never use bloated squircles |
| **Indicator Dots** | `rounded-full` | Restricted solely to avatars and status indicator dots |
| **Focus Rings** | `focus-visible:ring-1 focus-visible:ring-zinc-600` | Subtle, accessible keyboard focus rings |

> [!IMPORTANT]
> **Anti-Cartoon Rule**: Do not introduce bright pastel badges, oversized border radii, playful bouncing animations, or bloated squircle cards. Keep surfaces sharp, geometric, and functional.

---

## 2. Component Composition Checklist

Every reusable component authored in Orbito must satisfy the following four states:

1. **Idle / Default**: Crisp typography, high contrast, clean 1px border.
2. **Loading**: Use `LoadingSkeleton` or `Skeleton` with pulse animation matching the exact dimensions of the target content. Primary buttons must render an inline spinner and set `disabled={true}`.
3. **Empty**: Use `EmptyState.tsx` (`frontend/src/components/ui/EmptyState.tsx`) or an inline icon + description whenever arrays or lists have 0 items.
4. **Error**: Display a contextual banner or card with an explicit action trigger (e.g., `<Button onClick={refetch}>Retry</Button>`).

---

## 3. Standard Component Patterns

### 3.1 Radix UI + Tailwind Merge (`cn`)
All primitives must accept `className` and merge classes safely using `cn(...)`:
```tsx
import * as React from 'react';
import { cn } from '@/lib/utils'; // or standard classnames utility

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
  size?: 'sm' | 'md' | 'lg' | 'icon';
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center font-medium transition-colors cursor-pointer',
          'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-600 disabled:pointer-events-none disabled:opacity-50',
          'rounded-md text-xs',
          variant === 'primary' && 'bg-zinc-100 text-zinc-900 hover:bg-white shadow-xs',
          variant === 'secondary' && 'bg-zinc-800 text-zinc-200 hover:bg-zinc-750 border border-zinc-750',
          variant === 'outline' && 'border border-zinc-800 bg-zinc-900/60 text-zinc-300 hover:bg-zinc-800 hover:text-white',
          variant === 'ghost' && 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60',
          variant === 'destructive' && 'bg-rose-950/80 text-rose-300 border border-rose-900/60 hover:bg-rose-900/90',
          size === 'sm' && 'h-7 px-2.5 text-[11px]',
          size === 'md' && 'h-8 px-3 text-xs',
          size === 'lg' && 'h-10 px-4 text-sm',
          size === 'icon' && 'h-8 w-8 p-0',
          className
        )}
        {...props}
      />
    );
  }
);
Button.displayName = 'Button';
```

### 3.2 Controlled Modal Template
When building modals under `frontend/src/components/modals/`:
```tsx
'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

interface ActionModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: () => Promise<void>;
  title: string;
}

export function ActionModal({ open, onOpenChange, onSubmit, title }: ActionModalProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit();
      onOpenChange(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-md rounded-md p-6">
        <DialogHeader>
          <DialogTitle className="text-sm font-semibold text-zinc-100">{title}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          {/* Form fields with proper labels and inputs */}
          <DialogFooter className="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800/80">
            <Button type="button" variant="ghost" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Confirm'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

---

## 4. UI Quality Verification Steps

1. Verify keyboard accessibility: Tab navigation, Escape key closes overlays, Enter triggers primary actions.
2. Verify screen sizes: Ensure mobile (`<640px`), tablet (`768px`), and desktop (`1024px+`) viewports display zero horizontal overflow.
3. Test dark/light mode compatibility via `ThemeToggle`.
