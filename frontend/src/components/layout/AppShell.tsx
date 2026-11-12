'use client';

import * as React from 'react';
import { usePathname } from 'next/navigation';
import { Sparkles } from 'lucide-react';
import { Sidebar } from './Sidebar';
import { Navbar } from './Navbar';
import { KeyboardShortcutsDialog } from './KeyboardShortcutsDialog';
import { TaskDetailDrawer } from '@/features/tasks';
import { AiChatDrawer } from '@/features/ai-chat';
import { useAppDispatch } from '@/lib/store/hooks';
import { toggleAiChat } from '@/lib/store/slices/uiSlice';
import { Toaster } from 'sonner';

const AUTH_ROUTES = ['/login', '/register', '/forgot-password', '/reset-password', '/setup'];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const isAuthRoute = AUTH_ROUTES.some((route) => pathname?.startsWith(route));

  // Global ⌘J / Ctrl+J listener for AI Copilot
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'j') {
        e.preventDefault();
        dispatch(toggleAiChat());
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [dispatch]);

  if (isAuthRoute) {
    return (
      <div className="min-h-screen w-full bg-[#09090b] text-zinc-100 antialiased selection:bg-zinc-800 selection:text-white">
        {children}
        <Toaster
          position="bottom-right"
          theme="dark"
          toastOptions={{
            style: {
              background: '#18181b',
              border: '1px solid #27272a',
              color: '#fafafa',
            },
          }}
        />
      </div>
    );
  }

  return (
    <div className="flex h-screen w-full overflow-hidden bg-[#09090b] text-zinc-100 antialiased relative">
      <Sidebar />
      <div className="flex-1 overflow-hidden flex flex-col bg-[#09090b]">
        <Navbar />
        <main className="flex-1 overflow-y-auto flex flex-col bg-[#09090b]">
          {children}
        </main>
      </div>
      <TaskDetailDrawer />
      <AiChatDrawer />
      <KeyboardShortcutsDialog />

      {/* Floating Copilot Launcher */}
      <button
        type="button"
        onClick={() => dispatch(toggleAiChat())}
        className="fixed bottom-6 right-6 z-30 p-2.5 rounded-full bg-violet-600 hover:bg-violet-500 text-white shadow-lg shadow-violet-950/60 border border-violet-400/40 transition-all hover:scale-105 active:scale-95 flex items-center justify-center cursor-pointer"
        title="Open Orbito Copilot (⌘J)"
        aria-label="Open Orbito AI Copilot"
      >
        <Sparkles className="w-4 h-4 text-white" />
      </button>

      <Toaster
        position="bottom-right"
        theme="dark"
        toastOptions={{
          style: {
            background: '#18181b',
            border: '1px solid #27272a',
            color: '#fafafa',
          },
        }}
      />
    </div>
  );
}

