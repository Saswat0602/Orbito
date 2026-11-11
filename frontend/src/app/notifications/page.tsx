'use client';

import * as React from 'react';
import { NotificationCenter } from '@/features/notifications';

export default function NotificationsPage() {
  return (
    <div className="flex-1 overflow-y-auto bg-[#09090b]">
      <NotificationCenter />
    </div>
  );
}

