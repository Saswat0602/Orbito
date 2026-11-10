# Implementation Plan: Phase 4 — Rich Editor, Recurrence & Collaboration

## Overview
Implement **Phase 4: Rich Editor, Recurrence & Collaboration** across the `@orbito/platform` monorepo. This provides dual-mode Markdown editing (`Write` / `Preview`), interactive `@mention` autocomplete of workspace collaborators, task recurrence lifecycle management, visual recurring indicators, and a dual-tab Task Activity History and Workspace Activity Stream.

---

## Architecture Decisions
- **Markdown & Syntax Architecture**: Build `DualModeEditor` with standard Markdown shortcuts (Bold, Italic, Headings, Lists, Quotes, Code Blocks) and live split or tabbed Preview rendered via `react-markdown`.
- **Zero Third-Party Mention Lib**: Implement lightweight `@mention` detection using native textarea selection and a floating Radix popover to avoid heavy bundle overhead and ensure 100% Obsidian monochrome styling.
- **Native Backend Recurrence Leveraging**: Orbito backend already provides Prisma `RecurringTask` model and endpoints `POST/PUT/DELETE /api/tasks/:id/recurrence`. Extend `tasksApi.ts` with these endpoints and connect them to `TaskRecurrenceSelector`.
- **Task & Workspace Activity Integration**: Leverage backend `ActivityLogController` (`GET /api/activity-logs/task/:taskId/activities` and `/organization/:orgId/recent`) to power the Task Activity tab in `TaskDetailDrawer` and enhance `WorkspaceActivitiesPage`.

---

## Tasks Breakdown

- [ ] **Task 1: Recurrence & Activity RTK Query Endpoints**
  - Extend `tasksApi.ts` with `getTaskRecurrence`, `addTaskRecurrence`, `updateTaskRecurrence`, `deleteTaskRecurrence`.
  - Extend `activityLogApi.ts` with `getTaskActivities` querying `GET /api/activity-logs/task/:taskId/activities`.
  - Declare strict TypeScript types for `RecurringTaskConfig`, `RecurrenceType`, and `RecurrenceEndType`.

- [ ] **Task 2: Dual-Mode Markdown Editor (`DualModeEditor`)**
  - Create `frontend/src/components/common/DualModeEditor.tsx` with Markdown toolbar and `[Write]` / `[Preview]` tabs.
  - Support formatted code blocks, blockquotes, lists, links, and headers.
  - Style with Obsidian dark theme and `react-markdown`.

- [ ] **Task 3: Team Member `@mention` Autocomplete**
  - Implement `@mention` detection and auto-complete dropdown in `DualModeEditor` and comment inputs.
  - Query workspace members via `useGetWorkspaceMembersQuery`.
  - Render `@mentions` as stylized violet badges (`@FirstName LastName`) in comments and descriptions.

- [ ] **Task 4: Task Recurrence UI & Visual Badges**
  - Create `frontend/src/features/tasks/components/TaskRecurrenceSelector.tsx` with interval, day-of-week, and end condition controls.
  - Embed recurrence selector into `TaskDetailDrawer.tsx` and `CreateTaskModal.tsx`.
  - Add visual `Repeat` indicator badge to `TaskCard.tsx`, Gantt view, and List view.

- [ ] **Task 5: Task Activity History & Timeline**
  - In `TaskDetailDrawer.tsx`, add tab switcher: `[Comments]` and `[Activity History]`.
  - Render chronological audit timeline of field updates, reassignments, and comments with user avatars and relative timestamps.
  - Enhance `WorkspaceActivitiesPage` with entity filtering (`Task`, `Project`, `Workspace`).

- [ ] **Checkpoint: Complete Phase 4 Verification**
  - `npm run build --workspace=@orbito/frontend` (exit code 0)
  - `npm run build --workspace=@orbito/backend` (exit code 0)
  - Manual verification of Markdown editing, `@mention` picker, recurring task configuration, and activity logs.
