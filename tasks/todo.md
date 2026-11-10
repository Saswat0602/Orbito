# Task List: Phase 4 — Rich Editor, Recurrence & Collaboration

- [x] **Task 1: Recurrence & Activity RTK Query Endpoints**
  - [x] Add `RecurringTaskConfig`, `RecurrenceType`, `RecurrenceEndType` contracts in `types/tasks.ts`
  - [x] Extend `tasksApi.ts` with recurrence CRUD mutations (`POST/PUT/DELETE /api/tasks/:id/recurrence`)
  - [x] Extend `activityLogApi.ts` with `getTaskActivities` querying `GET /api/activity-logs/task/:taskId/activities`

- [x] **Task 2: Dual-Mode Markdown Editor (`DualModeEditor`)**
  - [x] Create `frontend/src/components/common/DualModeEditor.tsx`
  - [x] Implement toolbar actions (Bold, Italic, Headings, Lists, Quotes, Code Blocks)
  - [x] Add `[Write]` / `[Preview]` tabs with `react-markdown` prose styling
  - [x] Replace standard textareas in `TaskDetailDrawer` and `CreateTaskModal`

- [x] **Task 3: Team Member `@mention` Autocomplete**
  - [x] Build cursor `@mention` detection and dropdown in `DualModeEditor`
  - [x] Connect with workspace members via `useGetWorkspaceMembersQuery`
  - [x] Render styled `@mention` badges in comment threads and descriptions

- [x] **Task 4: Task Recurrence UI & Visual Badges**
  - [x] Create `frontend/src/features/tasks/components/TaskRecurrenceSelector.tsx`
  - [x] Support Daily, Weekly (day checkboxes), Monthly, and Custom end conditions
  - [x] Embed recurrence selector into `TaskDetailDrawer.tsx` metadata and `CreateTaskModal.tsx`
  - [x] Add `Repeat` badge indicator to `TaskCard.tsx`, Gantt view, and List view

- [x] **Task 5: Task Activity History & Timeline**
  - [x] Add `[Comments]` and `[Activity History]` tabs in `TaskDetailDrawer.tsx`
  - [x] Render chronological task audit trail with user avatars and relative timestamps
  - [x] Add entity type filters (`Task`, `Project`, `Workspace`) in `WorkspaceActivitiesPage`

- [x] **Checkpoint: Complete Phase 4 Verification**
  - [x] Frontend build succeeds (`npm run build --workspace=@orbito/frontend`)
  - [x] Backend build succeeds (`npm run build --workspace=@orbito/backend`)
  - [x] Verification of Markdown toolbar, mention picker, recurring config, and activity tabs
