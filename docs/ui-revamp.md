# Full UI Revamp — Layout, UX & Color Token Migration

## Context
The user wants three tiers of change:
1. **Color token migration** — every hardcoded rgba/hex or undefined CSS custom var replaced with shadcn tokens
2. **Layout revamp** — project detail page becomes a JIRA-style two-column layout: sticky stats sidebar on the left, thin filter bar + compact task board on the right
3. **UX revamp** — JIRA-style compact task cards (show partial info only), clicking a card opens the full edit modal; task board columns styled like JIRA (status header + count)

---

## Target Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  Nav (sticky, always dark green)                                │
├────────────────────┬────────────────────────────────────────────┤
│ Sidebar (sticky)   │ Filter bar (thin, sticky below nav)        │
│ ← Back             │ Status: [All▾]  Assignee: [All▾]  [+ New] [↻] │
│ PROJECT DETAIL     ├────────────────────────────────────────────┤
│ bk                 │  TO DO  2     IN PROGRESS  1     DONE  1   │
│ description…       │  ─────────────────────────────────────────│
│                    │  [compact    │  [compact    │  [compact     │
│ ── Stats ──────    │   card]      │   card]      │   card]       │
│  Visible    3      │              │              │               │
│  Total      3      │              │              │               │
│  Done       1      │              │              │               │
│  In progress 1     │              │              │               │
│  To do      1      │              │              │               │
│  Role       Owner  │              │              │               │
└────────────────────┴────────────────────────────────────────────┘
```

---

## Part 1 — CSS Token Fixes

### `frontend/src/styles.css`
Add to `:root` (no `.dark` block needed — they resolve to already-themed shadcn tokens):
```css
--ink-soft: var(--muted-foreground);
--accent-strong: var(--primary);
--surface-strong: var(--card);
--line-strong: var(--border);
```
Add to `@theme inline`:
```css
--color-ink-soft: var(--ink-soft);
--color-accent-strong: var(--accent-strong);
```

### `frontend/src/components/ui/badge.tsx`
Fix hardcoded and undefined custom vars:
- `bg-[rgba(19,59,51,0.1)] text-[var(--panel)]` → `bg-panel/10 text-panel`
- `bg-[rgba(19,59,51,0.08)]` (secondary) → `bg-panel/[0.08]`
- `border-[var(--line-strong)]` (outline) → `border-border`

---

## Part 2 — Auth Page

### `frontend/src/components/auth/AuthHero.tsx`
Hero card (intentionally always-dark brand element):
- `bg-[linear-gradient(145deg,rgba(19,59,51,0.98),rgba(36,86,76,0.9))]` → `bg-gradient-to-br from-panel to-panel/80`
- `border-white/40` → `border-panel-foreground/20`
- `text-[#f9f4ec]` → `text-panel-foreground`
- `text-[#d6bfa5]` → `text-panel-foreground/70`
- `text-[#f1e6d7]/82` → `text-panel-foreground/80`
- `shadow-[0_24px_80px_rgba(22,33,30,0.18)]` → `shadow-2xl`

Stat mini-cards:
- `bg-[rgba(255,251,246,0.78)] backdrop-blur-sm` → `bg-card shadow-sm`
- `bg-[rgba(19,59,51,0.08)] text-[var(--panel)]` → `bg-panel/10 text-panel`

---

## Part 3 — Projects Page

### `frontend/src/components/projects/ProjectsPageHeader.tsx`
- Section: `border-white/40 bg-[rgba(255,251,246,0.6)] backdrop-blur-sm` → `border-border bg-card shadow-sm`
- `text-[var(--accent-strong)]` → `text-primary`
- `text-[var(--ink-soft)]` → `text-muted-foreground`
- Stat cards: `bg-[rgba(255,251,246,0.78)] backdrop-blur-sm` → `bg-muted/60`
- Stat badges: `bg-[rgba(19,59,51,0.08)] text-[var(--panel)]` → `bg-panel/10 text-panel`
- Stat numbers: add `tabular-nums`

### `frontend/src/components/projects/ProjectCard.tsx`
- `text-[var(--ink-soft)]` → `text-muted-foreground`
- Description: add `line-clamp-2`
- Add `group` + `transition-shadow hover:shadow-md` to card
- "Open project" link: `bg-[var(--panel)] text-[#f9f4ec]` → `bg-primary text-primary-foreground`
- Arrow: add `→` that transitions on hover: `inline-block transition-transform group-hover:translate-x-1`

### `frontend/src/components/projects/CreateProjectCard.tsx`
- `text-[var(--ink-soft)]` → `text-muted-foreground`

### `frontend/src/components/projects/ProjectsListSection.tsx`
- `text-[var(--ink-soft)]` → `text-muted-foreground` (2×)

### `frontend/src/components/projects/ProjectsEmptyState.tsx`
- `text-[var(--ink-soft)]` → `text-muted-foreground`

---

## Part 4 — Project Detail Page: New Layout

### New: `frontend/src/components/project-detail/ProjectSidebar.tsx`
A self-contained sticky sidebar card. Receives `project`, `stats`, `roleLabel`.

Structure:
```
← Back to projects
PROJECT DETAIL (eyebrow)
[project name h2]
[description — muted, clamped to 3 lines]

─── Stats ─────────────
Visible      [count]
Total        [count]
Done         [count]
In progress  [count]
To do        [count]
Role         [Owner|Member]
```

Styling:
- `sticky top-[72px] h-[calc(100vh-72px)] overflow-y-auto`
- `w-64 shrink-0 border-r border-border bg-card p-5 flex flex-col gap-5`
- Stat rows: `flex justify-between text-sm`, value is `font-semibold tabular-nums`
- Done row: `text-success`; In progress: `text-warning`; others: default foreground

### New: `frontend/src/components/project-detail/TaskFilterBar.tsx`
Replaces `ProjectDetailFiltersCard`. A thin single-row bar, not a card.

Structure:
```
[Status: All▾] [Assignee: All▾]     [+ New task] [↻]
```

Styling:
- `flex items-center gap-3 border-b border-border bg-background px-4 py-2 sticky top-[72px] z-[4]`
- Compact selects styled as `h-7 text-sm rounded-md border border-input`
- "New task" → `<Button size="sm">`
- "Refresh" → `<Button size="icon-sm" variant="ghost">` with a ↻ character

### `frontend/src/pages/ProjectDetailPage.tsx`
Replace the current `flex-col` single-column layout with a two-column split:

```tsx
<div className="flex min-h-[calc(100vh-72px)]">
  {/* Sticky sidebar */}
  <ProjectSidebar project={project} stats={stats} roleLabel={...} />
  
  {/* Main content */}
  <div className="flex flex-1 flex-col min-w-0">
    <TaskFilterBar ... />
    
    {loadingTasks ? <LoadingState /> : null}
    {!loadingTasks && tasks.length === 0 ? <EmptyState /> : null}
    {!loadingTasks && tasks.length > 0 ? <TaskBoard ... /> : null}
  </div>
</div>
```

- Remove `ProjectDetailHeader` (replaced by sidebar)
- Remove `ProjectDetailFiltersCard` (replaced by filter bar)
- Stats fetched same as before, passed to sidebar
- `max-w-6xl` wrapper removed — the page now fills full width

---

## Part 5 — JIRA-style Task Board

### `frontend/src/components/project-detail/TaskColumn.tsx`
JIRA-style column header:
- Column container: `flex flex-col` with no rounded corners (full-height columns)
- Header: `flex items-center gap-2 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground border-b border-border`
  - Status label: `TO DO`, `IN PROGRESS`, `DONE`
  - Count badge: `ml-auto rounded-full bg-muted px-2 py-0.5 text-xs font-medium`
  - Status indicator dot: `size-2 rounded-full` colored by status (warning for in_progress, success for done, muted for todo)
- Drop zone: `flex-1 p-2 bg-muted/30 min-h-[400px]` with `bg-primary/5` when `isOver`
- Empty column card: `m-2 rounded-lg border-2 border-dashed border-border/50 p-6 text-center text-sm text-muted-foreground`

### `frontend/src/components/project-detail/TaskCard.tsx` — JIRA compact
Remove: status select, edit/delete buttons, assignee/due-date paragraph  
Keep: drag handle, title, priority badge, assignee initials bubble  
Add: card is clickable to open edit modal (click handler on the card body)

New card structure:
```
┌────────────────────────────────┐
│ ⠿  [Priority badge]            │  ← top row: drag handle + priority
│ Task title (2-line truncated)  │  ← main content, clickable
│                                │
│ [creator initials] ... [assignee initials] Due: date │  ← bottom meta
└────────────────────────────────┘
```

Styling:
- Card: `cursor-pointer hover:shadow-md transition-shadow`
- Top row: `flex items-center justify-between gap-2`
- Drag handle: `cursor-grab p-1 text-muted-foreground hover:bg-muted rounded shrink-0`
- Title: `text-sm font-medium leading-snug line-clamp-2 flex-1`
- Bottom meta: `flex items-center gap-2 text-xs text-muted-foreground`
- Assignee initials: `size-5 rounded-full bg-primary/20 text-primary text-[10px] font-bold grid place-items-center`

Props changes: add `onCardClick: (task: Task) => void`

### `frontend/src/components/project-detail/TaskBoard.tsx`
- Remove the `text-[var(--ink-soft)]` → `text-muted-foreground` (2×)
- Thread `onCardClick` prop down to TaskColumn → TaskCard
- Columns section: use `grid grid-cols-1 md:grid-cols-3` with no gap between columns (or `divide-x divide-border`)

### `frontend/src/components/tasks/TaskModal.tsx`
UX additions when `mode === "edit"`:
- Add a Delete button to the header row (alongside "Close")
- Needs `onDelete?: () => void` prop
- Overlay: `bg-[rgba(11,27,23,0.44)]` → `bg-black/50`
- Card: remove `bg-[var(--surface-strong)]` (use default `bg-card`)
- `text-[var(--accent-strong)]` → `text-primary`
- Card: add `shadow-2xl`
- Header: title + close/delete in a `flex items-start justify-between`

### Store changes (`frontend/src/features/project-detail/store.ts`)
Add a delete action flow: `openEditModal` triggers from card click; delete button in modal calls `handleDeleteTask` then closes.

No new state needed — the existing `openEditModal(task)` becomes the "show full detail" action.

### `frontend/src/features/project-detail/controllers/useProjectDetailTaskController.ts`
Export `handleDeleteTask` so TaskModal can call it (it's already returned — just needs to be accessible in the modal via props).

---

## Part 6 — ProtectedLayout

### `frontend/src/components/layout/ProtectedLayout.tsx`
- Outer `div`: add `bg-gradient-to-br from-background to-muted/30` (subtle page depth)
- Nav: add `shadow-sm`
- All existing token fixes from previous session remain

---

## Files Changed Summary

| File | Type of change |
|------|---------------|
| `styles.css` | Add 4 alias tokens + 2 @theme entries |
| `ui/badge.tsx` | Fix 3 hardcoded color values |
| `auth/AuthHero.tsx` | Full token migration |
| `projects/ProjectsPageHeader.tsx` | Token migration + solid card |
| `projects/ProjectCard.tsx` | Hover lift, primary CTA, line-clamp |
| `projects/CreateProjectCard.tsx` | Token fix |
| `projects/ProjectsListSection.tsx` | Token fix ×2 |
| `projects/ProjectsEmptyState.tsx` | Token fix |
| `layout/ProtectedLayout.tsx` | Page bg gradient, nav shadow |
| **NEW** `project-detail/ProjectSidebar.tsx` | Sticky stats sidebar |
| **NEW** `project-detail/TaskFilterBar.tsx` | Thin filter bar |
| `project-detail/TaskCard.tsx` | JIRA compact design, click-to-open |
| `project-detail/TaskColumn.tsx` | JIRA column header + dashed empty zone |
| `project-detail/TaskBoard.tsx` | Token fix, thread onCardClick |
| `tasks/TaskModal.tsx` | Token fix, delete button, backdrop fix |
| `pages/ProjectDetailPage.tsx` | Two-column layout, remove old header/filter |
| `features/.../useProjectDetailTaskController.ts` | Expose handleDeleteTask to modal |

---

## Verification
1. `npm run dev` — no type errors
2. Light mode: projects page has elevated solid cards, "Open project" is blue CTA, hover lifts cards
3. Dark mode: all colors adapt; no transparent text anywhere
4. Project detail: sidebar is sticky on scroll, stats update when tasks change
5. Filter bar: status + assignee filters work; New task opens modal; Refresh works
6. Task cards: compact JIRA style; click opens full edit modal with delete button
7. Drag and drop: still works (drag handle still present)
8. Auth page: hero card stays dark-green in both modes, stat cards use bg-card
