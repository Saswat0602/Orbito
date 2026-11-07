'use client';

import { StoreProvider } from '@/components/providers/StoreProvider';
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
      <body className="bg-[#09090b] text-zinc-100 antialiased min-h-screen">
        <StoreProvider>
          <QueryClientProvider client={queryClient}>
            <ThemeProvider>
              <AppShell>{children}</AppShell>
            </ThemeProvider>
          </QueryClientProvider>
        </StoreProvider>
      </body>
    </html>
  );
}
