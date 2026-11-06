# 07. UI/UX Design System: Modern Monochrome (Linear + Slack + Monday.com)

## 1. Design Philosophy

Orbito's visual identity abandons dated, toy-like dashboard aesthetics in favor of a sleek, high-velocity design system synthesized from the three leading productivity tools:

1. **Linear Elegance**:
   - Monochromatic obsidian and zinc palette.
   - Ultra-crisp typography (Inter / Geist font stack).
   - High data density with immaculate 1px border hierarchy (`border-zinc-800` / `border-zinc-200`).
   - Lightning-fast keyboard-first interaction with a global Command Palette (`Cmd+K` / `Ctrl+K`).
2. **Slack Collaboration Depth**:
   - Fluid left navigation with collapsible workspace trees, project channels, and unread badges.
   - Rich threaded task comments with inline `@mentions`, reaction emojis, and avatar stacks.
   - Real-time presence indicators (green dot online, subtle pulse on live typing).
3. **Monday.com Customization & Views**:
   - Multi-view project canvas: seamless tab switching between **Board (Kanban)**, **List**, **Sprints**, **Gantt**, and **Calendar**.
   - Custom field column expansion and inline cell editing.
   - Visual progress bars, sprint burndown charts, and workload distribution metrics.

---

## 2. Monochrome Color Palette & Tokens

### Dark Mode (Primary Default)
- **Background Root**: `#09090b` (`zinc-950`)
- **Card / Surface Background**: `#18181b` (`zinc-900`)
- **Hover / Sub-surface**: `#27272a` (`zinc-800`)
- **Borders & Dividers**: `#27272a` / `#3f3f46` (`zinc-800` / `zinc-700`)
- **Primary Text**: `#fafafa` (`zinc-50`)
- **Muted / Secondary Text**: `#a1a1aa` (`zinc-400`)
- **Tertiary / Subtle Text**: `#71717a` (`zinc-500`)

### Light Mode (High-Contrast Clean)
- **Background Root**: `#ffffff` (`white`)
- **Card / Surface Background**: `#f4f4f5` (`zinc-100`)
- **Hover / Sub-surface**: `#e4e4e7` (`zinc-200`)
- **Borders & Dividers**: `#e4e4e7` / `#d4d4d8` (`zinc-200` / `zinc-300`)
- **Primary Text**: `#09090b` (`zinc-950`)
- **Muted / Secondary Text**: `#52525b` (`zinc-600`)

### Semantic Accents (Purpose-Driven Cues)
Accents are reserved strictly for critical status and urgency recognition:
- **Success / Completed / Done**: Emerald `#10b981` (`emerald-500`)
- **In Progress / Active**: Sky / Blue `#0284c7` (`sky-600`)
- **Warning / Review / Planning**: Amber `#f59e0b` (`amber-500`)
- **Urgent / Blocker / Bug**: Crimson / Rose `#f43f5e` (`rose-500`)
- **AI Sparkles / Automation**: Violet `#8b5cf6` (`violet-500`)

---

## 3. Typography & Spacing Scale

- **Font Family**:
  ```css
  font-family: 'Geist', 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  ```
- **Hierarchy**:
  - `Display / Page Titles`: `text-xl font-semibold tracking-tight`
  - `Section / Board Column Headers`: `text-xs font-medium uppercase tracking-wider text-zinc-400`
  - `Task Titles / Card Headers`: `text-sm font-medium text-zinc-100`
  - `Meta / Badges / Dates`: `text-xs text-zinc-400 font-mono`
  - `Code / IDs / Keys (e.g. FF-104)`: `text-xs font-mono font-medium text-zinc-300`

---

## 4. Component Standards

### 4.1. Kanban Board & Task Cards
- **Columns**: Fixed width (280px-320px) with subtle background tint (`zinc-900/60`), column counter badge, and "+" quick-add button.
- **Cards**: `rounded-lg border border-zinc-800 bg-zinc-900 p-3 shadow-sm hover:border-zinc-700 hover:shadow-md transition-all cursor-grab active:cursor-grabbing`.
- **Card Elements**:
  - Top row: Task Prefix & Number badge (`FF-104`), Task Type icon (Bug, Story, Task), Priority indicator.
  - Middle: Task title (max 2 lines truncate) with hover preview.
  - Bottom row: Labels / tags, Due date indicator (red if overdue), Subtask count (`2/5`), Assignee avatar stack.

### 4.2. Global Command Palette (`Cmd+K` / `Ctrl+K`)
Instant overlay enabling:
- Navigation: Jump to any workspace, project, or sprint.
- Actions: "Create task", "Start sprint", "Open AI chat", "Toggle dark/light theme".
- Search: Full-text search across all tasks, comments, and project files.

### 4.3. Conversational AI Assistant Panel
- Resizable floating dock or right drawer (380px default width).
- Top header with LLM model indicator, clear history, and minimize buttons.
- Chat message feed with clean markdown rendering, collapsible tool-call execution cards, and live thinking indicators.
- Voice input button with waveform animation when listening.
