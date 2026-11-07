'use client';

import * as React from 'react';
import { usePathname } from 'next/navigation';
import { Sidebar } from './Sidebar';
import { Toaster } from 'sonner';

const AUTH_ROUTES = ['/login', '/register', '/forgot-password', '/reset-password', '/setup'];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isAuthRoute = AUTH_ROUTES.some((route) => pathname?.startsWith(route));

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
    <div className="flex h-screen w-full overflow-hidden bg-[#09090b] text-zinc-100 antialiased">
      <Sidebar />
      <main className="flex-1 overflow-hidden flex flex-col bg-[#09090b]">
        {children}
      </main>
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

