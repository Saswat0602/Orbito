# Task List: Phase 3 — Orbito AI Copilot & Conversational Task Assistant

- [x] **Task 1: API & Redux State Foundation**
  - [x] Add `isAiChatOpen`, `setIsAiChatOpen`, `toggleAiChat` to `uiSlice.ts`
  - [x] Extend `aiChatApi.ts` with conversation mutations and description generator
  - [x] Declare AI TypeScript contracts in `types/ai.ts`

- [x] **Task 2: Speech-to-Text Voice Hook**
  - [x] Create `frontend/src/lib/hooks/useSpeechRecognition.ts` with Web Speech API
  - [x] Handle listening state, permission handling, error fallback

- [x] **Task 3: Floating AI Copilot Drawer & Conversation History**
  - [x] Create `frontend/src/features/ai-chat/components/AiChatDrawer.tsx`
  - [x] Build conversation history drawer (new chat, rename, delete)
  - [x] Build message feed with markdown renderer and suggested prompt chips
  - [x] Wire voice microphone button and auto-context injection
  - [x] Export from `frontend/src/features/ai-chat/index.ts`

- [x] **Task 4: Global AppShell Integration & Keyboard Shortcut**
  - [x] Mount `AiChatDrawer` in `AppShell.tsx`
  - [x] Add global `⌘J` / `Ctrl+J` shortcut listener
  - [x] Add Copilot trigger button in `Navbar.tsx`

- [x] **Task 5: AI Description Generator in Tasks**
  - [x] Add "Generate with AI" button in `CreateTaskModal.tsx`
  - [x] Add "Generate with AI" button in `TaskDetailDrawer.tsx`

- [x] **Task 6: Admin AI Provider Settings Card**
  - [x] Add AI Copilot card in `frontend/src/app/admin/settings/page.tsx`
  - [x] Add Provider URL, API key, Model fields and Test Connection button

- [x] **Checkpoint: Complete Phase 3 Verification**
  - [x] Frontend build succeeds (`npm run build --workspace=@orbito/frontend`)
  - [x] Backend build succeeds (`npm run build --workspace=@orbito/backend`)
  - [x] Manual test of `⌘J`, chat messages, suggestions, voice, and description generator

