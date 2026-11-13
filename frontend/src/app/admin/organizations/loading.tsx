import { AdminTableSkeleton } from '@/components/skeletons';

export default function AdminOrganizationsLoading() {
  return (
    <AdminTableSkeleton
      title="Organizations"
      description="Multi-tenant boundary management, suspensions, and ownership transfers."
    />
  );
}

