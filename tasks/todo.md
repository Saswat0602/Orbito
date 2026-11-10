# Task List: Phase 5 — Custom Workflows & Inbound Email

- [x] **Task 1: Types & API Endpoint Contracts**
  - [x] Add task status positions, fromProject, byProject, byId, restore endpoints in `frontend/src/lib/api/endpoint.ts`
  - [x] Add `workflowId` to `Project` and `_count` to `TaskStatus` in `@orbito/types`
  - [x] Re-export typed inbox contracts in `frontend/src/types/inbox.ts` and `frontend/src/types/index.ts`

- [x] **Task 2: Redux Store Integration (`inboxApi.ts` & `tasksApi.ts`)**
  - [x] Register `'TaskStatus'` and `'Workflow'` tags in `apiSlice.ts`
  - [x] Add status CRUD and batch position update mutations in `tasksApi.ts`
  - [x] Create dedicated `inboxApi.ts` with IMAP setup, connection test, sync, rules, messages, and task conversion endpoints

- [x] **Task 3: Interactive Kanban Workflow & Status Editor (`WorkflowEditor.tsx`)**
  - [x] Create `frontend/src/features/projects/components/WorkflowEditor.tsx`
  - [x] Render horizontal status timeline and stage cards with category badges and hex color accents
  - [x] Add atomic position swap buttons (`Move Left`, `Move Right`) persisting to `PATCH /api/task-statuses/positions`
  - [x] Implement "+ Add Stage Column" dialog with category selector and color presets
  - [x] Implement "Edit Stage" and "Delete Stage" dialogs with task protection checks

- [x] **Task 4: Inbound Email Ingestion & Rules Management (`ProjectInboxTab.tsx`)**
  - [x] Create `frontend/src/features/projects/components/ProjectInboxTab.tsx`
  - [x] Build Mailbox Connection setup card with quick presets (Gmail, Outlook, Custom IMAP)
  - [x] Add task auto-creation preferences (type, priority, assignee, sync interval)
  - [x] Add IMAP connection test button and manual "Sync Now" trigger
  - [x] Build Inbound Automation Rules table and "+ Add Rule" modal (conditions + actions)
  - [x] Build Incoming Message review queue with 1-click "Convert to Task" and "Ignore" actions

- [x] **Task 5: Project Settings Integration & Verification**
  - [x] Mount `<WorkflowEditor />` in `frontend/src/app/[workspaceSlug]/[projectSlug]/settings/page.tsx`
  - [x] Add 5th tab `inbox` in Project Settings hosting `<ProjectInboxTab />`
  - [x] Role-gate editing actions with `useAuthRbac` (`canEdit: isManager || isOwner || isSuperAdmin`)
  - [x] Compile check: `@orbito/types`, `@orbito/backend`, and `@orbito/frontend` build with 0 errors
