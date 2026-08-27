# Spec: Step 5 — Core UI flow (pick class → student → notes → summary)

Status: ready for `build`
Roadmap step: 5

## Context for the builder

Step 4 built read-only GET pages: `/`, `/classes/{id}`, `/students/{id}`. The
templates already carry HTMX attributes (class `<select hx-get="/classes">`,
student `<select hx-get="/students/{id}">`, notes form `hx-post="/notes/{id}"`),
but none of those handlers exist yet.

This step implements the full single-user flow, all with HTMX partials +
server-rendered HTML. **No auth, no multi-student batching, no LLM settings in
the UI.**

## The flow to implement

1. `/` shows classes. Teacher can add a class.
2. Click a class → `/classes/{id}` shows its students. Teacher can add a student.
3. Click a student → `/students/{id}` shows notes + editable summary.
4. Teacher pastes/enters notes, hits "Generate Summary".
5. Server sends notes to the `Summarizer`, saves the result, returns an editable
   `<textarea>` partial.
6. Teacher edits and hits "Save", then copies out.

## Task 1 — Create class (form + POST)

On `/` (index page), add a small form: text input `name` + button "Add Class",
`hx-post="/classes"` targeting the class list `#class-list` with `hx-swap="innerHTML"`.

Add handler `POST /classes`:
- Parse `r.FormValue("name")`; if empty → 400.
- `store.CreateClass(ctx, name)`.
- Respond with the re-rendered class-list partial (all classes, name-sorted).

Add a reusable partial `class_list.html` (defines `classList`) listing classes
as links to `/classes/{id}`. The index page uses it; the POST re-renders it.

Acceptance: `go build ./...` passes.

## Task 2 — Add student (form + POST)

On `/classes/{id}` (class page), add a form: text input `name` + button "Add
Student", `hx-post="/classes/{id}/students"` targeting `#student-list` with
`hx-swap="innerHTML"`.

Add handler `POST /classes/{id}/students`:
- Parse id from path; on error → 400.
- Parse `r.FormValue("name")`; if empty → 400.
- `store.CreateStudent(ctx, classID, name)`.
- Re-render the student-list partial for that class.

Add a reusable `student_list.html` partial (defines `studentList`) listing
students as links to `/students/{id}`.

Acceptance: `go build ./...` passes.

## Task 3 — Append notes (form + POST)

On `/students/{id}`, the existing "Enter Notes" form posts to `/notes/{id}`.
Add handler `POST /notes/{id}`:
- Parse id; on error → 400.
- Parse `r.FormValue("notes")`; if empty → 400.
- `store.AppendNote(ctx, studentID, notes)`.
- Redirect (`303 See Other`) to `/students/{id}` so the page reloads with the
  new note listed.

Use a redirect (not a partial) — appending is a discrete save, and a full
reload is the simplest mental model for a non-technical user.

Acceptance: `go build ./...` passes.

## Task 4 — Generate summary (POST + HTMX partial)

Add handler `POST /students/{id}/summary`:
- Parse id; on error → 400.
- Load notes via `store.ListNotesByStudent(ctx, studentID)`.
- Concatenate note contents (newline-joined) into a single prompt string.
- Call `s.summarizer.Summarize(notesText)`.
  - On error: respond with an HTMX partial `summary_partial.html` showing an
    `.error` div and HTTP 502 so HTMX swaps in the error message.
- On success: `store.SaveSummary(ctx, studentID, summary)`.
- Respond with partial `summary_partial.html` rendering an editable
  `<textarea name="summary">{{.Summary}}</textarea>` + a "Save Summary" button
  (`hx-post="/students/{id}/summary/save"`, `hx-target="#summary-section"`,
  `hx-swap="outerHTML"`) + the copy button.

Add a button on `/students/{id}`: "Generate Summary", `hx-post="/students/{id}/summary"`,
targeting `#summary-section` with `hx-swap="innerHTML"` and an indicator.

### Wire the Summarizer into the server

- `Server` gains a `summarizer summarize.Summarizer` field.
- `NewServer(st, tmpl, summarizer)` takes it. Update `main.go` to build it via
  `summarize.OllamaSummarizerFromEnv()`; on error, `log.Fatal` (server must not
  start misconfigured). `server_test.go`'s `newTestServer` passes a fake.

Add a fake in the test package:

```go
type fakeSummarizer struct{ out string; err error }
func (f fakeSummarizer) Summarize(notes string) (string, error) { return f.out, f.err }
```

Acceptance: `go build ./...`, `go vet ./...` pass.

## Task 5 — Save edited summary (POST + HTMX partial)

Add handler `POST /students/{id}/summary/save`:
- Parse id; on error → 400.
- Parse `r.FormValue("summary")`; if empty → 400.
- `store.SaveSummary(ctx, studentID, summary)`.
- Re-render `summary_partial.html` with the saved text (same partial as Task 4,
  so the editable box stays editable).

Acceptance: `go build ./...` passes.

## Task 6 — Templates

- New partials: `class_list.html`, `student_list.html`, `summary_partial.html`.
  Each defines a named block (`classList`, `studentList`, `summary`).
- `index.html` and `class.html` use the partials via `{{template "classList" .}}`
  etc. with `id` wrappers (`#class-list`, `#student-list`) so HTMX can target.
- `student.html` gets the Generate button and a `#summary-section` div (initially
  empty unless a saved summary exists).
- Update `LoadTemplates` page list to parse the new partial files for the pages
  that need them (`index`, `class`, `student`). Keep one template set per page.

Acceptance: `web.LoadTemplates()` succeeds; pages render (covered by tests).

## Task 7 — Tests

Extend `internal/server/server_test.go` with table-driven tests (keep the fake
summarizer):

1. **TestCreateClassPOST** — POST `/classes` with `name=Math` → 200, body
   contains `Math`, and `store.ListClasses` shows 1.
2. **TestCreateClassEmptyName** — POST `/classes` with no name → 400.
3. **TestAddStudentPOST** — seed a class, POST `/classes/{id}/students` with
   `name=Bob` → 200, body contains `Bob`, store shows the student.
4. **TestAppendNotePOST** — seed class+student, POST `/notes/{id}` with
   `notes=hello` → 303 redirect to `/students/{id}`, store shows the note.
5. **TestGenerateSummary** — fake returns `"A fine summary"`. POST
   `/students/{id}/summary` → 200, body contains the summary, store has it
   saved.
6. **TestGenerateSummaryError** — fake returns an error → 502, body contains an
   error message.
7. **TestSaveSummaryPOST** — POST `/students/{id}/summary/save` with edited text
   → 200, body contains the new text, store has it.

Each test asserts real output (status + body + store state).

Acceptance: `go test ./...` passes.

## Task 8 — Verify tree is clean

- `go build ./...`, `go vet ./...` — clean.
- `go test ./...` — all pass.

## Explicitly out of scope

- Auth, external access, multi-student batching — v1 "Do not" list.
- LLM settings in the UI — modelfile/step 7 handles temp/top_p.
- Anonymizing notes before sending to the model — deferred.

## Handoff to build

1. Partials + index/class/student template updates + `LoadTemplates` page list.
2. `Server.summarizer` field, `NewServer` signature, `main.go` wiring, fake in tests.
3. POST handlers: `/classes`, `/classes/{id}/students`, `/notes/{id}`,
   `/students/{id}/summary`, `/students/{id}/summary/save`.
4. Table-driven tests for all five POST flows.
5. Verify build/vet/test.