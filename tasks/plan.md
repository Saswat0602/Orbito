# Implementation Plan: Phase 3 — Orbito AI Copilot & Conversational Task Assistant

## Overview
Implement **Phase 3: Orbito AI Copilot & Conversational Task Assistant** across the `@orbito/platform` monorepo. This brings autonomous task assistance, voice-to-text queries, multi-session conversation histories, context-aware project summarization, and AI description generation to Orbito.

---

## Architecture Decisions
- **Unified UI Redux State**: AI Copilot open/close state (`isAiChatOpen`) is managed in `uiSlice.ts` to allow triggering from floating buttons, navbar icons, and global keyboard shortcuts (`⌘J` / `Ctrl+J`).
- **Browser-Native Speech-to-Text**: Utilize the standard browser Web Speech API (`SpeechRecognition` / `webkitSpeechRecognition`) wrapped in a clean, reusable custom hook (`useSpeechRecognition`) with zero third-party audio packages.
- **Context Injection**: Every chat request to `POST /api/ai-chat/chat` automatically injects `workspaceId`, `projectId`, and `currentOrganizationId` from Redux auth and workspace state, allowing the LLM to access current project and sprint context.
- **Admin Configuration with Fallback**: Support AI provider credentials (OpenAI, OpenRouter, Anthropic, or local Ollama) configurable in `/admin/settings` as well as `.env` environment variables.

---

## Task List

### Task 1: API & Redux State Foundation (`S`)
**Description:** Extend `frontend/src/lib/store/slices/uiSlice.ts` with `isAiChatOpen` actions and update `frontend/src/lib/store/api/aiChatApi.ts` with full conversation management endpoints (`createConversation`, `renameConversation`, `deleteConversation`, `generateDescription`, `testConnection`).

**Acceptance criteria:**
- [ ] `uiSlice` exports `isAiChatOpen`, `setIsAiChatOpen`, and `toggleAiChat`.
- [ ] `aiChatApi` provides hooks: `useSendChatMessageMutation`, `useGetConversationsQuery`, `useCreateConversationMutation`, `useRenameConversationMutation`, `useDeleteConversationMutation`, `useGenerateDescriptionMutation`, and `useTestAiConnectionMutation`.

**Verification:**
- [ ] Frontend builds with `npm run build --workspace=@orbito/frontend`

**Files touched:**
- `frontend/src/lib/store/slices/uiSlice.ts`
- `frontend/src/lib/store/api/aiChatApi.ts`
- `frontend/src/types/ai.ts`

---

### Task 2: Speech-to-Text Voice Hook (`S`)
**Description:** Implement `frontend/src/lib/hooks/useSpeechRecognition.ts` to provide speech-to-text functionality via Web Speech API with browser compatibility detection, real-time transcription, and error handling.

**Acceptance criteria:**
- [ ] Hook returns `{ isListening, transcript, isSupported, startListening, stopListening, resetTranscript }`.
- [ ] Gracefully falls back if Web Speech API is not supported in the user's browser.

**Verification:**
- [ ] Hook exports cleanly and passes TypeScript typecheck.

**Files touched:**
- `frontend/src/lib/hooks/useSpeechRecognition.ts`

---

### Task 3: Floating AI Copilot Drawer & Conversation History (`M`)
**Description:** Build `frontend/src/features/ai-chat/components/AiChatDrawer.tsx` featuring a slide-out panel, conversation history sidebar, message stream with markdown formatting, prompt suggestion chips, voice microphone button, and context header.

**Acceptance criteria:**
- [ ] Slide-out drawer with Obsidian dark aesthetic and backdrop blur.
- [ ] Multi-session history: create new chat, switch sessions, rename, and delete sessions.
- [ ] Suggested prompt chips: "Summarize this sprint", "What tasks are overdue?", "What high-priority bugs need attention?".
- [ ] Microphone button activates speech-to-text and populates input.
- [ ] Automatically injects active workspace & project ID into request payload.

**Verification:**
- [ ] Renders messages and handles session switching without errors.

**Files touched:**
- `frontend/src/features/ai-chat/components/AiChatDrawer.tsx`
- `frontend/src/features/ai-chat/index.ts`

---

### Task 4: Global AppShell Integration & Keyboard Shortcut (`S`)
**Description:** Mount `AiChatDrawer` in `AppShell.tsx`, register global `⌘J` / `Ctrl+J` shortcut listener, and add an AI Copilot trigger button in `Navbar.tsx`.

**Acceptance criteria:**
- [ ] Pressing `⌘J` or `Ctrl+J` anywhere in the app toggles the AI Copilot drawer.
- [ ] Floating or navbar button with `Sparkles` icon opens the drawer.

**Verification:**
- [ ] Shortcut toggles drawer smoothly across pages.

**Files touched:**
- `frontend/src/components/layout/AppShell.tsx`
- `frontend/src/components/layout/Navbar.tsx`

---

### Task 5: AI Description Generator in Tasks (`S`)
**Description:** Add "Generate with AI" button (`Sparkles` icon) to `CreateTaskModal.tsx` and `TaskDetailDrawer.tsx` to automatically draft structured task descriptions from the task title.

**Acceptance criteria:**
- [ ] Clicking "Generate with AI" calls `useGenerateDescriptionMutation`.
- [ ] Auto-populates description textarea with generated content.
- [ ] Displays loading state during generation.

**Verification:**
- [ ] Generating description populates the form without modifying other task inputs.

**Files touched:**
- `frontend/src/features/tasks/components/CreateTaskModal.tsx`
- `frontend/src/features/tasks/components/TaskDetailDrawer.tsx`

---

### Task 6: Admin AI Provider Settings Card (`S`)
**Description:** Add an AI Copilot configuration panel in `frontend/src/app/admin/settings/page.tsx` allowing system admins to configure AI Provider URL, API Key, Model, and test connection via `POST /api/ai-chat/test-connection`.

**Acceptance criteria:**
- [ ] Form fields for `ai_enabled` (toggle), `ai_api_url`, `ai_api_key`, `ai_model`.
- [ ] "Test Connection" button with status feedback toast.
- [ ] Persists settings via `POST /api/admin/config`.

**Verification:**
- [ ] Settings save and test connection responds with provider status.

**Files touched:**
- `frontend/src/app/admin/settings/page.tsx`

---

## Checkpoint: End-to-End Verification
- [ ] `npm run build --workspace=@orbito/frontend` (code 0)
- [ ] `npm run build --workspace=@orbito/backend` (code 0)
- [ ] Verify `⌘J` shortcut opens Copilot drawer.
- [ ] Verify AI prompt suggestions and voice input.
- [ ] Verify "Generate with AI" in task creation modal.

