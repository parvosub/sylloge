# Sylloge v1.1 Web Interface Enhancement Spec

## Goal
Polish the existing web interface for the single art teacher user and add summary editing/history support so she can review and iterate on generated summaries before final copy-out.

## Scope Guard
- **In scope**: Visual polish, navigation, summary editing, summary history, desktop-only UX improvements.
- **Deferred / not in scope**: auth, multi-student batching, PII anonymization, LLM sliders, mobile-first design, PDF/CSV export, external access.

---

## Task 1: Global visual polish and navigation

### Files to modify
- `internal/web/templates/base.html`
- `internal/web/templates/index.html`
- `internal/web/templates/class.html`
- `internal/web/templates/student.html`

### Changes

**`base.html`**
- Replace the current generic styles with a clean, warm, teacher-friendly design:
  - Body: `#f8f9fa` background, `system-ui, -apple-system, sans-serif` font stack
  - Container: `max-width: 900px`, white card, `border-radius: 12px`, soft shadow
  - Header: title + short subtitle, centered, with a small color accent (e.g., teal `#0d9488`)
- Add a reusable breadcrumb region at the top of the container.
- Add a `.toast` container near the top of `<body>` for transient feedback messages.
- Add shared CSS classes:
  - `.btn-primary`, `.btn-secondary`, `.btn-danger`, `.btn-sm`
  - `.card`, `.card-list`, `.empty-state`
  - `.form-group`, `.input`, `.textarea`
  - `.badge`, `.spinner`
  - `.htmx-indicator` with a subtle fade/spinner instead of plain text.
- Ensure all buttons have `min-height: 44px` and visible focus states.
- Keep HTMX script tag.

**`index.html`**
- Wrap class list in a `.card-list` with each class as a `.card` showing:
  - Class name as a prominent link
  - Created date in a muted `.badge`
- Add an `.empty-state` block when no classes exist.
- Move the "Add Class" form into a `.card` with a clear heading.
- Keep `hx-post="/classes"` targeting `#class-list`.

**`class.html`**
- Add breadcrumb: `Home > Class: {ClassName}`.
- Put the "Add Student" form and student list into `.card` containers.
- Render students as `.card` entries with name and created date badge.

**`student.html`**
- Add breadcrumb: `Home > {ClassName} > {StudentName}`.
- Split the page into two clear cards:
  1. **Notes** — existing form + list of prior notes.
  2. **Summary** — Generate Summary button + `#summary-section`.
- Add a clear visual separator between notes and summary.
- Improve note rendering with date badges and alternating background.

### Acceptance criteria
- `go build ./...` and `go vet ./...` pass.
- All existing server tests still pass.
- Browser check: home, class, and student pages render cleanly with the new styles, breadcrumbs, and card layout.
- Empty states are friendly and not just blank.

---

## Task 2: Toast notifications and loading states

### Files to modify
- `internal/web/templates/base.html`
- `internal/web/templates/summary_partial.html`
- `internal/web/templates/student.html`

### Changes

**`base.html`**
- Add a small inline `<script>` block at the end of `<body>`:
  - `function showToast(message, type)` appends a toast div to the toast container.
  - `htmx:afterSwap` listener: if the swapped response contains a toast trigger, show it.
  - `htmx:responseError` listener: show a generic error toast.
- Keep the script tiny and dependency-free.

**`summary_partial.html`**
- On successful save, include a `data-toast` attribute in the rendered summary div so the toast script can show "Summary saved."
- Keep the existing summary textarea + Save/Copy buttons.

**`student.html`**
- Replace the "Processing..." text with a spinner icon inside `#summary-loading`.
- Ensure the Generate Summary button has `hx-disabled-elt="this"` and a clear loading state.

### Acceptance criteria
- `go build ./...` and `go vet ./...` pass.
- Existing tests pass.
- Manual check: clicking Generate Summary shows a loading indicator; clicking Save Summary shows a toast.

---

## Task 3: Summary history list

### Files to modify
- `internal/store/store.go`
- `internal/store/store_test.go`
- `internal/server/server.go`
- `internal/server/server_test.go`
- `internal/web/templates/student.html`
- `internal/web/templates/summary_partial.html` (or new `summary_history_partial.html`)

### Changes

**`internal/store/store.go`**
- Add a new method:

```go
// ListSummariesByStudent returns all saved summaries for a student, newest first.
func (s *Store) ListSummariesByStudent(ctx context.Context, studentID int64) ([]Summary, error)
```

- Implementation: query `summaries` table for `student_id = ? ORDER BY id DESC`, scan into `[]Summary`.
- Update the existing `SaveSummary` to return the newly inserted `Summary` (ID, timestamps) rather than only the ID. Change signature to:

```go
func (s *Store) SaveSummary(ctx context.Context, studentID int64, content string) (Summary, error)
```

  This requires updating all callers.

**`internal/store/store_test.go`**
- Add table-driven test `TestListSummariesByStudent`:
  - Save 3 summaries for a student.
  - Verify returned order is newest first.
  - Verify each has a non-zero ID and content.

- Update existing `TestSaveAndGetSummary` and `TestMultipleSummaries` to expect `Summary` return values.

**`internal/server/server.go`**
- Update `handleGenerateSummary` to use the new `SaveSummary` return value and pass the saved summary ID to the template.
- Update `handleSaveSummary` to use the new `SaveSummary` return value.
- Add new handler:

```go
func (s *Server) handleSummariesHistory(w http.ResponseWriter, r *http.Request)
```

  Route: `GET /students/{id}/summaries`
  - Parse student ID.
  - Fetch student and class (for breadcrumb context).
  - Call `ListSummariesByStudent`.
  - Render a new template set or partial showing the history list.
  - Return `404` if student not found.

- Register the new route in `routes()`:

```go
mux.HandleFunc("GET /students/{id}/summaries", s.handleSummariesHistory)
```

- Update `handleStudent` to pass `Summaries` (latest 5 or all) to the `student` template so the history preview is visible on the student page.

**`internal/web/templates.go`**
- Add `summaries_history` to the partials list or load it as a standalone page template.
- If shown as a standalone page, add it to the `pages` list.

**`internal/web/templates/student.html`**
- Add a "Summary History" section below the current summary card.
- Show a compact list of prior summaries with date and first ~100 characters.
- Each history entry links to `GET /students/{id}/summaries` (or expands inline if preferred).
- Add a "View full history" link.

**`internal/web/templates/summary_history.html` (new)**
- Full-page template with breadcrumb and a list of all summaries.
- Each summary rendered as a card with:
  - Timestamp
  - Full content in a readonly textarea
  - A "Restore to current" button that copies content into the current summary textarea (optional, may be deferred).

**`internal/web/templates/summary_partial.html`**
- After saving, display the summary's saved timestamp in a badge.

### Acceptance criteria
- `go build ./...`, `go vet ./...`, `go test ./...` pass.
- New `TestListSummariesByStudent` passes.
- Server tests for summary generation and save still pass.
- Manual check: generate multiple summaries for a student; the history section shows all versions newest first.

---

## Task 4: Enhanced summary editing experience

### Files to modify
- `internal/web/templates/student.html`
- `internal/web/templates/summary_partial.html`

### Changes

**`summary_partial.html`**
- Keep the editable `<textarea id="summary" name="summary">`.
- Add a character counter below the textarea.
- Add clear primary/secondary buttons:
  - **Save Summary** (primary)
  - **Copy to Clipboard** (secondary)
- After save, render the summary content and a "Saved at {timestamp}" badge.

**`student.html`**
- Ensure the "Add a Note" form also uses the new `.card` and `.form-group` styles.
- Add a small help text: "Summaries are editable. Generate a new version anytime."

### Acceptance criteria
- `go build ./...` and `go vet ./...` pass.
- Existing tests pass.
- Manual check: generate a summary, edit the text, click Save — the saved version appears in history and the textarea keeps the edited text.

---

## Task 5: Accessibility and semantic HTML pass

### Files to modify
- `internal/web/templates/base.html`
- `internal/web/templates/index.html`
- `internal/web/templates/class.html`
- `internal/web/templates/student.html`
- `internal/web/templates/summary_partial.html`
- `internal/web/templates/summary_history.html` (if created)

### Changes
- Use semantic elements: `<main>`, `<nav>`, `<section>`, `<article>`, `<header>`.
- Add `aria-label` to navigation links and buttons.
- Add `aria-live="polite"` to the toast container.
- Ensure every form input has a `<label>` with `for` matching `id`.
- Ensure focus styles are visible.
- Use sufficient color contrast for all text.

### Acceptance criteria
- `go build ./...` and `go vet ./...` pass.
- Existing tests pass.
- Visual inspection: no broken layouts, clear labels, visible focus rings.

---

## Task 6: Run tests and verify locally

### Verification steps
1. `go test ./...`
2. `go build ./...`
3. `go vet ./...`
4. Start the server with fake/local Ollama if available:
   - `OLLAMA_BASE_URL=http://localhost:11434 OLLAMA_MODEL=sylloge go run ./cmd/sylloge`
   - Open `http://localhost:8080`
   - Add a class, student, and note.
   - Generate summary, edit it, save it.
   - Confirm history shows multiple versions.
   - Confirm copy-to-clipboard and toast notifications work.

### Definition of done
- All tests pass.
- No lint/build errors.
- UI matches the polished design described above.
- Summary editing and history are functional end-to-end.

---

## Handoff to build

Execute in this order:

1. **Task 1**: Global visual polish and navigation — update `base.html`, `index.html`, `class.html`, `student.html`.
2. **Task 2**: Toast notifications and loading states — update `base.html`, `summary_partial.html`, `student.html`.
3. **Task 3**: Summary history list — update `store.go`, `store_test.go`, `server.go`, `server_test.go`, templates, add `summary_history.html`.
4. **Task 4**: Enhanced summary editing experience — update `student.html`, `summary_partial.html`.
5. **Task 5**: Accessibility and semantic HTML pass — all templates.
6. **Task 6**: Run tests and verify locally.

Stop after each task for review. Do not proceed to the next task until the current one passes `go build ./...`, `go vet ./...`, and the relevant tests.
