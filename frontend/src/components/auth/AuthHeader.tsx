'use client';

import * as React from 'react';
import Link from 'next/link';

interface AuthHeaderProps {
  title: string;
  subtitle: string;
}

export function AuthHeader({ title, subtitle }: AuthHeaderProps) {
  return (
    <div className="flex flex-col items-center text-center space-y-2 mb-8">
      {/* Orbito Logo Mark */}
      <Link
        href="/"
        className="group mb-2 flex items-center justify-center h-12 w-12 rounded-xl bg-zinc-900 border border-zinc-800 hover:border-zinc-700 transition-all duration-200 shadow-md"
        aria-label="Orbito Home"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="h-6 w-6 text-zinc-100 group-hover:scale-105 transition-transform"
        >
          {/* Planetary Orbit geometric icon */}
          <circle cx="12" cy="12" r="4" />
          <path d="M21.17 8A10.06 10.06 0 0 0 12 2a10 10 0 0 0-8.94 5.5" />
          <path d="M2.83 16A10.06 10.06 0 0 0 12 22a10 10 0 0 0 8.94-5.5" />
        </svg>
      </Link>

      <span className="text-xs font-semibold uppercase tracking-widest text-zinc-400">
        Orbito Platform
      </span>

      <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-zinc-50">
        {title}
      </h1>

      <p className="text-sm text-zinc-400 max-w-sm">
        {subtitle}
      </p>
    </div>
  );
}

