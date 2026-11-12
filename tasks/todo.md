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
