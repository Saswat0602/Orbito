# Task List: Complete Taskosaur Parity (Task Edit Design, Page Styles & Advanced Work Management)

- [x] **Slice 1: Task Edit Design & Dedicated Full-Page Task View**
  - [x] Create 2-column `TaskDetailWorkspace.tsx` component with inline title/desc editing, Save/Cancel, and rich Markdown toolbar
  - [x] Implement full-page route `frontend/src/app/[workspaceSlug]/[projectSlug]/tasks/[taskId]/page.tsx`
  - [x] Add "Expand to Fullscreen" (`Maximize2`) button in `TaskDetailDrawer.tsx`
  - [x] Add Metadata Inspector: Type selector, Sprint selector, Story Points selector (Fibonacci), Start Date, Due Date, Reporter
  - [x] Add Task Hierarchy & Subtask Progress Bar (`TaskHierarchy.tsx`)

- [x] **Slice 2: Workspace Projects Catalog Page Styling**
  - [x] Upgrade `frontend/src/app/[workspaceSlug]/projects/page.tsx` with search filter (debounced)
  - [x] Add Project Status (`ALL`, `ACTIVE`, `COMPLETED`, `ARCHIVED`) and Priority filter pills
  - [x] Create `ProjectHealthIndicator.tsx` circular SVG gauge (0–100%) and throughput heatmap
  - [x] Enhance project cards with health score, active sprint, member avatar stack, and progress bar

- [x] **Slice 3: Workspace Members & Pending Invitations Management**
  - [x] Add dual-tab switcher in `frontend/src/app/[workspaceSlug]/members/page.tsx`: `Active Members` vs `Pending Invitations`
  - [x] Add Member search filter with debounce
  - [x] Add Role selector dropdown (`Select`) to update member roles with confirmation
  - [x] Create `PendingInvitationsList.tsx` showing active tokens, copy link, resend, and revoke actions

- [x] **Slice 4: Dedicated Full-Page Notification Center (`/notifications`)**
  - [x] Create `frontend/src/app/notifications/page.tsx`
  - [x] Create `NotificationCenter.tsx` with category filter tabs (`All`, `Tasks`, `Mentions`, `Status Updates`, `Invitations`)
  - [x] Add search input, "Mark All as Read", and "Delete Read" actions
  - [x] Connect footer link in `NotificationPopover.tsx` to `/notifications`

- [x] **Slice 5: Advanced Tasks Organization & Productivity Tools**
  - [x] Create `GroupByManager.tsx` (Status, Priority, Assignee, Type, Due Date)
  - [x] Create `SortingManager.tsx` (multi-field sort with ASC/DESC)
  - [x] Create `ColumnManager.tsx` to toggle table column visibility
  - [x] Create `taskExport.ts` for 1-click export to CSV and Excel (XLSX)
  - [x] Create `BulkActionsBar.tsx` floating multi-select bar in List view
  - [x] Create `TodayAgendaDialog.tsx` daily briefing modal with header shortcut

- [x] **Checkpoint: Taskosaur Parity Complete**
  - [x] `npm run build --workspace=@orbito/frontend` passes with 0 errors
  - [x] `npm run build --workspace=@orbito/backend` passes with 0 errors
  - [x] Full end-to-end verification of task editing, full-page view, projects, members, notifications, and board tools

---

# Phase 6: Workspace Analytics, Profile Customization, AI Settings & Shortcuts

- [x] **Task 1: Workspace Overview & Aggregated Analytics**
  - [x] Update `frontend/src/app/[workspaceSlug]/page.tsx` to aggregate metrics across all projects in the workspace
  - [x] Add Completion Rate %, Workload Health, Status & Priority distribution progress bars
  - [x] Add Quick Action buttons (New Project, Invite Member, Open Agenda, Projects Overview)

- [x] **Task 2: User Profile Avatar & Preferences**
  - [x] Update `frontend/src/lib/store/api/usersApi.ts` `UpdateProfileRequest` to support `avatar?: string` and `preferences?: Record<string, unknown>`
  - [x] Update `frontend/src/app/settings/page.tsx` with avatar upload dropzone, image preview, removal, and Department/Job Title fields

- [x] **Task 3: AI Copilot Preferences & Settings**
  - [x] Create `frontend/src/features/settings/components/AISettings.tsx`
  - [x] Persona selector (Agile Scrum Master, Staff Architect, Product Manager, Custom)
  - [x] Creativity/Temperature slider (0.0 to 1.0) and voice auto-submit toggle
  - [x] Integrate into `frontend/src/app/settings/page.tsx` as a new "AI Assistant" tab

- [x] **Task 4: Global Keyboard Shortcuts Dialog**
  - [x] Create `frontend/src/components/layout/KeyboardShortcutsDialog.tsx`
  - [x] Implement global `?` key listener (ignoring inputs/textareas)
  - [x] Hotkeys display: Navigation (`⌘K`, `⌘J`, `G+B`, `G+L`, `G+P`, `G+M`), Tasks (`C`/`N`, `D`, `Esc`), Shortcuts (`?`)
  - [x] Mount in `frontend/src/components/layout/AppShell.tsx` with trigger in sidebar/footer

- [x] **Checkpoint: Build & Verification**
  - [x] `npm run build --workspace=@orbito/frontend` passes with 0 errors
  - [x] `npm run build --workspace=@orbito/backend` passes with 0 errors
  - [x] Dev servers healthy on ports 3000 and 3001

---

# Phase 7: Project Operations & Productivity

- [x] **Task 1: Dedicated Sprint Execution Board**
  - [x] Implement `frontend/src/app/[workspaceSlug]/[projectSlug]/sprints/[sprintId]/page.tsx`
  - [x] Scoped sprint view: Kanban board, List view, and Gantt view filtered strictly to that sprint
  - [x] Sprint header with name, goal, status badge, dates, burndown metric badge, story points progress bar, and "+ Add Task to Sprint" button
  - [x] Link sprint cards in `SprintPlanningView.tsx` to `/[workspaceSlug]/[projectSlug]/sprints/[sprintId]`

- [x] **Task 2: Dedicated Project Members Portal**
  - [x] Add RTK Query endpoints for `project-members` in `frontend/src/lib/store/api/projectsApi.ts` (`getProjectMembers`, `updateProjectMemberRole`, `removeProjectMember`, `inviteProjectMember`)
  - [x] Create `frontend/src/features/projects/components/ProjectInviteMemberModal.tsx`
  - [x] Implement `frontend/src/app/[workspaceSlug]/[projectSlug]/members/page.tsx` with search, role dropdown, remove member, and pending invitations
  - [x] Update project navigation in `Sidebar.tsx` to include "Members" link

- [x] **Task 3: CSV Task Bulk Import Modal**
  - [x] Enhanced `frontend/src/features/tasks/components/CsvImportModal.tsx` with `sprintId?: string` support
  - [x] Connected to `useBulkCreateTasksMutation`
  - [x] Added `Import CSV` action button to tasks board, list toolbars, and sprint execution board

- [x] **Task 4: Live Stopwatch Time Tracker**
  - [x] Updated `frontend/src/features/tasks/components/TaskTimeTracking.tsx` with start/pause/reset stopwatch controls
  - [x] Animated ticking display with live duration calculation
  - [x] 1-click "Stop & Log" auto-filling duration into the time log dialog

- [x] **Task 5: Quick Share Button in Task Detail Drawer**
  - [x] Verified `<ShareTaskDialog>` trigger in `TaskDetailWorkspace.tsx` and `TaskDetailDrawer.tsx`
  - [x] Accessible from drawer and fullpage task workspaces

---

# Phase 8: Organization Intelligence & Super Admin Console

- [x] **Task 1: Executive Organization Analytics Dashboard**
  - [x] Implement `frontend/src/app/organization/page.tsx` and `frontend/src/app/dashboard/page.tsx`
  - [x] Query `GET /organizations/:id/charts` for KPI metrics, project portfolio, team utilization, and task distribution
  - [x] Render portfolio health cards, completion rates, work breakdown bars, and workspace quick-jump grid
  - [x] Export endpoint `getOrganizationCharts` in `frontend/src/lib/store/api/organizationsApi.ts`

- [x] **Task 2: Organization Workspaces Directory & Archive Management**
  - [x] Implement `frontend/src/app/workspaces/page.tsx`
  - [x] Tabbed views for Active vs Archived workspaces
  - [x] Endpoints for `getArchivedWorkspaces`, `archiveWorkspace`, and `unarchiveWorkspace` in `frontend/src/lib/store/api/workspacesApi.ts`
  - [x] Dialog to create new workspace with name, slug preview, and description

- [x] **Task 3: Super Admin User Drill-Down Inspector**
  - [x] Implement `frontend/src/app/admin/users/[id]/page.tsx`
  - [x] Status toggle (`ACTIVE` / `INACTIVE`), Role elevation (`SUPER_ADMIN` / `MEMBER`), Password reset email dispatch, and Account deletion
  - [x] Link to user inspector from admin user directory table and dropdowns

- [x] **Task 4: Super Admin Organization Drill-Down Inspector**
  - [x] Implement `frontend/src/app/admin/organizations/[id]/page.tsx`
  - [x] Organization details, primary owner, member list, and workspace directory
  - [x] Actions: Suspend/Activate tenant, Transfer ownership dialog, and Permanent deletion dialog
  - [x] Export `getAdminOrganizationDetail` and `deleteAdminOrganization` in `frontend/src/lib/store/api/adminApi.ts`
  - [x] Link to organization inspector from admin organizations table and dropdowns

- [x] **Checkpoint: Phase 8 Build & Verification**
  - [x] `npm run build --workspace=@orbito/frontend` passes with 0 errors
  - [x] `npm run build --workspace=@orbito/backend` passes with 0 errors
  - [x] Live servers healthy: API on port 3000 (`/api/health`), Web on port 3001

---

# Phase 9: Onboarding, Rich Mentions & Compliance

- [x] **Task 1: First-Run Super Admin Setup Wizard**
  - [x] Created `frontend/src/app/setup/page.tsx`
  - [x] Integrated `getSetupRequired` and `setupSuperAdmin` endpoints in `frontend/src/lib/store/api/authApi.ts`
  - [x] Added public route exception in `frontend/src/middleware.ts`
  - [x] Handles both pending setup initialization and "already configured" guard state

- [x] **Task 2: Editor `@mention` Autocomplete Popover**
  - [x] Created `frontend/src/features/tasks/components/MentionAutocomplete.tsx`
  - [x] Exported from `frontend/src/features/tasks/index.ts`
  - [x] Integrated into `TaskDetailWorkspace.tsx` comments input with real-time `@` detection
  - [x] Keyboard navigation (`ArrowUp`/`ArrowDown`/`Enter`/`Escape`) and automatic formatting into `@[Name]`

- [x] **Task 3: Member Public Profile Page**
  - [x] Created `frontend/src/app/[workspaceSlug]/members/[id]/page.tsx`
  - [x] Header profile card with avatar, role badge, job title, department, joined date
  - [x] Contact information card with timezone and email
  - [x] Assigned active tasks list in current workspace with priority, status, and direct task links
  - [x] Linked member list rows in `/[workspaceSlug]/members/page.tsx` directly to profile

- [x] **Task 4: Public Legal Compliance Pages**
  - [x] Created `frontend/src/app/terms-of-service/page.tsx`
  - [x] Created `frontend/src/app/privacy-policy/page.tsx`
  - [x] Added public route exceptions in `frontend/src/middleware.ts`
  - [x] Styled in Obsidian monochrome design system with responsive typography and navigation

- [x] **Checkpoint: Phase 9 Build & Verification**
  - [x] `npm run build --workspace=@orbito/frontend` passes with 0 errors (all 22 static + dynamic routes generated)
  - [x] `npm run build --workspace=@orbito/backend` passes with 0 errors
  - [x] Dev servers healthy on port 3000 (`/api/health`) and port 3001 (`/setup`, `/terms-of-service`, `/privacy-policy` return 200 OK)
