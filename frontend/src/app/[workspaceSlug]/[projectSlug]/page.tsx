'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';

export default function ProjectIndexPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const router = useRouter();

  useEffect(() => {
    if (params?.workspaceSlug && params?.projectSlug) {
      router.replace(`/${params.workspaceSlug}/${params.projectSlug}/tasks`);
    }
  }, [params, router]);

  return null;
}
