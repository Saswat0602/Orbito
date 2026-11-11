# Task List: Complete Taskosaur Parity (Task Edit Design, Page Styles & Advanced Work Management)

- [x] **Slice 1: Task Edit Design & Dedicated Full-Page Task View**
  - [x] Create 2-column `TaskDetailWorkspace.tsx` component with inline title/desc editing, Save/Cancel, and rich Markdown toolbar
  - [x] Implement full-page route `frontend/src/app/[workspaceSlug]/[projectSlug]/tasks/[taskId]/page.tsx`
  - [x] Add "Expand to Fullscreen" (`Maximize2`) button in `TaskDetailDrawer.tsx`
  - [x] Add Metadata Inspector: Type selector, Sprint selector, Story Points selector (Fibonacci), Start Date, Due Date, Reporter
  - [x] Add Task Hierarchy & Subtask Progress Bar (`TaskHierarchy.tsx`)

- [ ] **Slice 2: Workspace Projects Catalog Page Styling**
  - [ ] Upgrade `frontend/src/app/[workspaceSlug]/projects/page.tsx` with search filter (debounced)
  - [ ] Add Project Status (`ALL`, `ACTIVE`, `COMPLETED`, `ARCHIVED`) and Priority filter pills
  - [ ] Create `ProjectHealthIndicator.tsx` circular SVG gauge (0–100%) and throughput heatmap
  - [ ] Enhance project cards with health score, active sprint, member avatar stack, and progress bar

- [ ] **Slice 3: Workspace Members & Pending Invitations Management**
  - [ ] Add dual-tab switcher in `frontend/src/app/[workspaceSlug]/members/page.tsx`: `Active Members` vs `Pending Invitations`
  - [ ] Add Member search filter with debounce
  - [ ] Add Role selector dropdown (`Select`) to update member roles with confirmation
  - [ ] Create `PendingInvitationsList.tsx` showing active tokens, copy link, resend, and revoke actions

- [ ] **Slice 4: Dedicated Full-Page Notification Center (`/notifications`)**
  - [ ] Create `frontend/src/app/notifications/page.tsx`
  - [ ] Create `NotificationCenter.tsx` with category filter tabs (`All`, `Tasks`, `Mentions`, `Status Updates`, `Invitations`)
  - [ ] Add search input, "Mark All as Read", and "Delete Read" actions
  - [ ] Connect footer link in `NotificationPopover.tsx` to `/notifications`

- [ ] **Slice 5: Advanced Tasks Organization & Productivity Tools**
  - [ ] Create `GroupByManager.tsx` (Status, Priority, Assignee, Type, Due Date)
  - [ ] Create `SortingManager.tsx` (multi-field sort with ASC/DESC)
  - [ ] Create `ColumnManager.tsx` to toggle table column visibility
  - [ ] Create `taskExport.ts` for 1-click export to CSV and Excel (XLSX)
  - [ ] Create `BulkActionsBar.tsx` floating multi-select bar in List view
  - [ ] Create `TodayAgendaDialog.tsx` daily briefing modal with header shortcut

- [ ] **Checkpoint: Build & Verification**
  - [ ] `npm run build --workspace=@orbito/frontend` passes with 0 errors
  - [ ] `npm run build --workspace=@orbito/backend` passes with 0 errors
  - [ ] Full end-to-end verification of task editing, full-page view, projects, members, notifications, and board tools
