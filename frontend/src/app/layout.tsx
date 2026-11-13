'use client';

import { StoreProvider } from '@/components/providers/StoreProvider';
import { AuthProvider } from '@/components/providers/AuthProvider';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@/lib/context/ThemeContext';
import { AppShell } from '@/components/layout/AppShell';
import './globals.css';
import { useState } from 'react';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  return (
    <html lang="en" data-theme="dark" className="dark" suppressHydrationWarning>
      <head>
        <title>Orbito | Enterprise Work Management</title>
        <meta name="description" content="Enterprise Task, Project & Work Management Platform" />
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="icon" href="/icon.svg" type="image/svg+xml" />
        <link rel="apple-touch-icon" href="/apple-icon.png" />
      </head>
      <body className="bg-[#09090b] text-zinc-100 antialiased min-h-screen">
        <StoreProvider>
          <AuthProvider>
            <QueryClientProvider client={queryClient}>
              <ThemeProvider>
                <AppShell>{children}</AppShell>
              </ThemeProvider>
            </QueryClientProvider>
          </AuthProvider>
        </StoreProvider>
      </body>
    </html>
  );
}
