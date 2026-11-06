# 06. AI Engine, Integrations & Inbound Email Specification

## 1. Conversational AI Task Execution Architecture

Orbito incorporates an advanced **Conversational AI Task Execution Engine** that transforms conversational natural language into tangible project management operations.

```
       [User Input (Text or Voice)]
                    │
                    ▼
       ┌────────────────────────┐
       │     ChatPanel UI       │ <──── Context (Current Org, Workspace, Project, URL)
       └────────────┬───────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │  Backend AI Service    │ ──── [AES-256-GCM Encrypted BYOK Keys]
       │  (ai-chat.service.ts)  │ ──── [SSRF-Guarded Outbound Provider]
       └────────────┬───────────┘
                    │
                    ▼  (Structured Action Plan / Tool Calls)
       ┌────────────────────────┐
       │  In-App Browser Agent  │
       │   (browser-agent.ts)   │
       └────────────┬───────────┘
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
   [DOM Action Executor]   [API Fallback Direct Exec]
   (Clicks, Inputs, Dnd)    (Fast deterministic REST)
```

### 1.1. Multi-Provider LLM Integration (BYOK)
Orbito users can provide their own LLM API credentials or use organization defaults:
- **OpenRouter**: Access to 100+ open-source and proprietary models (`https://openrouter.ai/api/v1`).
- **OpenAI**: Native GPT-4o, GPT-4o-mini support (`https://api.openai.com/v1`).
- **Anthropic**: Claude 3.5 Sonnet, Claude 3 Opus (`https://api.anthropic.com/v1`).
- **Local AI / Ollama**: Self-hosted LLM endpoints (`http://localhost:11434/v1`).
- **SSRF Hardening**: Endpoints are validated against `AI_ALLOWED_HOSTS` with private IP resolution blocked unless `AI_ALLOW_PRIVATE_ENDPOINTS=true` is explicitly configured.

### 1.2. In-App Browser Automation Agent
When instructed (e.g., *"Create a new sprint named Q3 Sprint 1 and move task FF-12 into it"*), the frontend agent executes actions through:
1. **DOM Detector (`dom-detector.ts`)**: Identifies interactive components via `data-testid`, `aria-label`, button text, and form inputs.
2. **Action Executor (`action-executor.ts`)**: Emulates human user actions (clicking buttons, filling forms, selecting dropdowns, and confirming dialogs) with visual status feedback.
3. **Voice Controller (`voice.ts`)**: Web Speech API integration streaming audio transcripts directly into the AI prompt runner.

---

## 2. External Project Management Integrations

### 2.1. Jira Integration (Workspace & Project Levels)
- **Scope 1: Workspace-Level**: Connects an Atlassian account to an entire Orbito workspace; enables bulk browsing and importing of Jira projects.
- **Scope 2: Project-Level**: Ongoing bidirectional sync between a specific Orbito project and a Jira project key (e.g. `PROJ`).
- **SSRF Defense (`safe-outbound-http.ts`)**:
  - Validates site URL against `JIRA_ALLOWED_HOSTS` (defaults to `*.atlassian.net`).
  - Resolves domain to IP addresses using DNS pre-flight.
  - Verifies no IP falls within RFC 1918 private subnets, loopback (`127.0.0.1`), link-local (`169.254.0.0/16`), or IPv6 equivalents.
  - **DNS Pinning**: Pins connection to the validated IP to defeat DNS rebinding attacks.
  - Disallows HTTP redirects.
- **Status Mapping**: Configurable mapping between Jira status IDs and Orbito status IDs.

### 2.2. Trello Integration
- Connects Trello account via API Key and User Token.
- Maps Trello Boards to Orbito Projects, and Trello Lists to Orbito Workflow Statuses.
- Bidirectional synchronization of cards, descriptions, checklists, and labels.

---

## 3. Inbound Project Email Inbox (IMAP to Task Pipeline)

Orbito allows teams to receive tasks directly via email without leaving their inbox:

```
[Customer / Client Email]
         │
         ▼
[IMAP Mail Server (SSL/TLS)]
         │ (Cron Scheduled Poll / Scheduler)
         ▼
[Inbox Processor (inbox.processor.ts)]
         │
         ├── 1. IMAP UID & Message-ID Duplicate Check
         ├── 2. Thread-ID Matcher (Appends reply to existing task comments)
         ├── 3. Attachment Extractor (Saves to S3/Storage, creates TaskAttachments)
         ├── 4. Rule Matcher (InboxRule: filter by subject, sender, regex)
         │       └── Actions: Auto-assign, set priority, attach label
         │
         ▼
[New Orbito Task Created] ──> Emits WebSocket `task:created`
```

### Inbound Rule Engine (`InboxRule` & `MessageRuleResult`)
- Match conditions: `FROM_CONTAINS`, `SUBJECT_CONTAINS`, `BODY_REGEX`.
- Assigned actions: Target Project, Target Status (e.g. `Triage`), Default Assignee, Priority level, Labels.
