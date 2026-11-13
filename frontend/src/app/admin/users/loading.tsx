import { AdminTableSkeleton } from '@/components/skeletons';

export default function AdminUsersLoading() {
  return (
    <AdminTableSkeleton
      title="Users Directory"
      description="Manage enterprise users, roles, statuses, and password credentials."
    />
  );
}

