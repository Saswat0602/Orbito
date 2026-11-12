'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Keyboard, Command, Sparkles, Navigation, CheckSquare } from 'lucide-react';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { toggleAiChat } from '@/lib/store/slices/uiSlice';

interface ShortcutItem {
  keys: string[];
  description: string;
}

interface ShortcutCategory {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  shortcuts: ShortcutItem[];
}

const SHORTCUT_CATEGORIES: ShortcutCategory[] = [
  {
    title: 'Global & Navigation',
    icon: Command,
    shortcuts: [
      { keys: ['⌘', 'K'], description: 'Open Command Palette & Global Search' },
      { keys: ['⌘', 'J'], description: 'Toggle Orbito AI Copilot Assistant' },
      { keys: ['?'], description: 'Show Keyboard Shortcuts Dialog' },
      { keys: ['Esc'], description: 'Close active modal, drawer, or dialog' },
    ],
  },
  {
    title: 'Work Management',
    icon: CheckSquare,
    shortcuts: [
      { keys: ['C'], description: 'Create new task (when not typing)' },
      { keys: ['D'], description: "Open Today's Agenda briefing" },
      { keys: ['G', 'B'], description: 'Go to active Tasks Board' },
      { keys: ['G', 'L'], description: 'Go to active Tasks List' },
      { keys: ['G', 'C'], description: 'Go to Calendar view' },
      { keys: ['G', 'T'], description: 'Go to Timeline / Gantt view' },
    ],
  },
  {
    title: 'Workspace Navigation',
    icon: Navigation,
    shortcuts: [
      { keys: ['G', 'P'], description: 'Navigate to Projects Catalog' },
      { keys: ['G', 'M'], description: 'Navigate to Members Directory' },
      { keys: ['G', 'N'], description: 'Navigate to Notification Center' },
      { keys: ['G', 'A'], description: 'Navigate to Workspace Activities' },
      { keys: ['G', 'S'], description: 'Navigate to Account Settings' },
    ],
  },
];

interface KeyboardShortcutsDialogProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function KeyboardShortcutsDialog({
  isOpen: controlledIsOpen,
  onClose: controlledOnClose,
}: KeyboardShortcutsDialogProps = {}) {
  const router = useRouter();
  const params = useParams();
  const dispatch = useAppDispatch();
  const activeWorkspaceSlug = useAppSelector((state) => state.auth.activeWorkspaceSlug);
  const activeProjectSlug = useAppSelector((state) => state.auth.activeProjectSlug);

  const [internalIsOpen, setInternalIsOpen] = useState(false);
  const isOpen = controlledIsOpen !== undefined ? controlledIsOpen : internalIsOpen;

  const handleClose = () => {
    if (controlledOnClose) {
      controlledOnClose();
    } else {
      setInternalIsOpen(false);
    }
  };

  // Global key listener
  useEffect(() => {
    let pendingGKey = false;
    let gKeyTimeout: NodeJS.Timeout | null = null;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is currently typing in an input, textarea, or contentEditable
      const target = e.target as HTMLElement | null;
      if (
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable)
      ) {
        return;
      }

      // Check for ? (Shift + /)
      if (e.key === '?' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        setInternalIsOpen((prev) => !prev);
        return;
      }

      // Check for 'G' sequence
      if (e.key.toLowerCase() === 'g' && !e.metaKey && !e.ctrlKey) {
        pendingGKey = true;
        if (gKeyTimeout) clearTimeout(gKeyTimeout);
        gKeyTimeout = setTimeout(() => {
          pendingGKey = false;
        }, 1200);
        return;
      }

      if (pendingGKey && !e.metaKey && !e.ctrlKey) {
        const key = e.key.toLowerCase();
        pendingGKey = false;
        if (gKeyTimeout) clearTimeout(gKeyTimeout);

        const ws = (params?.workspaceSlug as string) || activeWorkspaceSlug;
        const prj = (params?.projectSlug as string) || activeProjectSlug;

        if (key === 'b' && ws && prj) {
          e.preventDefault();
          router.push(`/${ws}/${prj}/tasks`);
        } else if (key === 'l' && ws && prj) {
          e.preventDefault();
          router.push(`/${ws}/${prj}/tasks`);
        } else if (key === 'p' && ws) {
          e.preventDefault();
          router.push(`/${ws}/projects`);
        } else if (key === 'm' && ws) {
          e.preventDefault();
          router.push(`/${ws}/members`);
        } else if (key === 'a' && ws) {
          e.preventDefault();
          router.push(`/${ws}/activities`);
        } else if (key === 'n') {
          e.preventDefault();
          router.push('/notifications');
        } else if (key === 's') {
          e.preventDefault();
          router.push('/settings');
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      if (gKeyTimeout) clearTimeout(gKeyTimeout);
    };
  }, [router, params, activeWorkspaceSlug, activeProjectSlug]);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-xl bg-zinc-950 border-zinc-800 text-zinc-200 p-0 shadow-2xl rounded-xl overflow-hidden">
        <DialogHeader className="p-5 pb-3 border-b border-zinc-850 bg-zinc-900/60">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-md bg-indigo-950/50 border border-indigo-800/40 flex items-center justify-center text-indigo-400 shrink-0">
              <Keyboard className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <span>Keyboard Shortcuts</span>
                <span className="text-[10px] font-mono font-normal px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700">
                  Press ? anywhere
                </span>
              </DialogTitle>
              <DialogDescription className="text-xs text-zinc-400 mt-0.5">
                Quick commands and key combinations to navigate Orbito like a power user.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="max-h-[60vh] overflow-y-auto p-5 space-y-6">
          {SHORTCUT_CATEGORIES.map((category) => {
            const Icon = category.icon;
            return (
              <div key={category.title} className="space-y-2.5">
                <div className="flex items-center gap-2 text-xs font-semibold text-zinc-300 uppercase tracking-wider px-1">
                  <Icon className="h-3.5 w-3.5 text-indigo-400" />
                  <span>{category.title}</span>
                </div>

                <div className="rounded-md border border-zinc-850 bg-zinc-900/40 divide-y divide-zinc-850/60">
                  {category.shortcuts.map((shortcut, idx) => (
                    <div
                      key={idx}
                      className="flex items-center justify-between px-3 py-2 text-xs hover:bg-zinc-850/30 transition-colors"
                    >
                      <span className="text-zinc-300">{shortcut.description}</span>
                      <div className="flex items-center gap-1 shrink-0 ml-3">
                        {shortcut.keys.map((k, kIdx) => (
                          <React.Fragment key={kIdx}>
                            <kbd className="h-5 min-w-[20px] px-1.5 flex items-center justify-center rounded bg-zinc-950 border border-zinc-800 font-mono text-[10px] font-semibold text-zinc-200 shadow-xs">
                              {k}
                            </kbd>
                            {kIdx < shortcut.keys.length - 1 && (
                              <span className="text-[10px] text-zinc-600 font-mono">+</span>
                            )}
                          </React.Fragment>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </DialogContent>
    </Dialog>
  );
}

