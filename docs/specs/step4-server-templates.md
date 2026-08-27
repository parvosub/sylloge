# Spec: Step 4 — HTTP server + templates

Status: ready for `build`
Roadmap step: 4

## Context for the builder

`internal/server/server.go` is a stub (renders a plain-text string, no store, no
templates). `internal/web/templates.go` embeds `templates/*.html` but parsing all
four files at once redefines the `main` block (index/class/student each define
it), so only the last-parsed page would render. `main.go` calls
`server.NewServer()` with no wiring.

This step turns the server into a real `html/template` app with three read-only
GET pages wired to the store, plus a 404. The dynamic HTMX submit flows
(notes entry, summary generation, save/copy) are **step 5** — do not implement
POST handlers that call the summarizer in this step.

## Task 1 — Add `GetClass` to the store

`internal/store/store.go` already has `GetStudent`, `ListClasses`, etc. Add the
missing lookup needed by the class/student pages:

```go
func (s *Store) GetClass(ctx context.Context, id int64) (Class, error)
```
- `SELECT id, name, created_at FROM classes WHERE id = ?`
- Return `ErrNotFound` when no row matches (mirror `GetStudent`).

Acceptance: `go build ./...` passes.

## Task 2 — Per-page template loading

Rewrite `LoadTemplates` in `internal/web/templates.go` so each page parses
`base.html` plus exactly one page file (no `main` redefinition conflicts):

```go
func LoadTemplates() (map[string]*template.Template, error)
```
- Key: `"index"`, `"class"`, `"student"`.
- For each key, parse `templates/base.html` + `templates/<key>.html` into a
  fresh `template.Template`.
- Render a page with `t.ExecuteTemplate(w, "base.html", data)`.
- Keep the `//go:embed templates/*` and exported `Templates` var unchanged.

Acceptance: `go build ./...` passes.

## Task 3 — Update templates to render store data

Adjust the three page templates so field names match the data structs the server
will pass. Keep the existing HTMX attributes and CSS untouched.

### `index.html` — data `{ Classes []store.Class }`
Already ranges over `.Classes` — keep as-is. Verify it compiles when given
`[]store.Class`.

### `class.html` — data `{ ClassID int64, ClassName string, Students []store.Student }`
- Title uses `.ClassName`.
- Range over `.Students` (already does).
- Keep the "Back to Classes" link pointing at `/`.

### `student.html` — data `{ StudentID int64, StudentName string, ClassID int64, ClassName string, Notes []store.Note, Summary string }`
- `.Notes` is now a slice: change `{{.Notes}}` to a `{{range .Notes}}` list.
- `.Summary` is a plain string (empty when none exists). Render the editable
  `<textarea>` **only when `.Summary != ""`**; otherwise show nothing.
- Keep the notes form and its `hx-post` attribute (step 5 wires the handler).
- Keep the copy-to-clipboard button and script.

Acceptance: templates parse with `LoadTemplates()` (tested by Task 5).

## Task 4 — Wire the server

Rewrite `internal/server/server.go`:

```go
type Server struct {
    store     *store.Store
    templates map[string]*template.Template
}

func NewServer(st *store.Store, tmpl map[string]*template.Template) *Server
func (s *Server) Run(addr string) error
```

`Run` sets up a `http.ServeMux` (Go 1.22+ method patterns) with:
- `GET /` → render `index` with `ListClasses(ctx)`.
- `GET /classes/{id}` → `GetClass(id)` + `ListStudentsByClass(id)`; render
  `class`. If `GetClass` returns `ErrNotFound`, respond 404.
- `GET /students/{id}` → `GetStudent(id)`, `GetClass(student.ClassID)`,
  `ListNotesByStudent(id)`, `GetSummaryByStudent(id)` (treat `ErrNotFound` as
  empty summary); render `student`. 404 on unknown student.
- All other paths/methods → 404.

Use `http.Server{Addr: addr, Handler: mux}` and `ListenAndServe`. Use
`log.Printf` for startup lines. Do **not** hardcode the port; it comes from the
`addr` argument.

Parse the class/student IDs with `strconv.ParseInt`; on error respond 400.

Acceptance: `go build ./...` and `go vet ./...` pass.

## Task 5 — Update `main.go` wiring

`cmd/sylloge/main.go` must construct the store and templates and pass them in:

```go
st, err := store.NewStore()
tmpl, err := web.LoadTemplates()
srv := server.NewServer(st, tmpl)
if err := srv.Run(os.Getenv("SYLLOGE_ADDR")); err != nil { ... }
```
- Read `SYLLOGE_ADDR`; default to `":8080"` when unset.
- `log.Fatal` on any error (keep existing style).
- Keep `store.NewStore()` and `web.LoadTemplates()` calls; add the new imports.

Acceptance: `go build ./...` passes; `go vet ./...` passes.

## Task 6 — Server tests (table-driven)

Create `internal/server/server_test.go`. Build a test store with a temp DB file
(`t.TempDir()` + `store.Open(path)` — **add** `func Open(dbPath string) (*Store, error)` to `internal/store/store.go` that opens the given path and runs `createTables`, and make `NewStore()` call `Open("sylloge.db")` so behavior is unchanged), seed classes/students/notes/summaries, load templates once, and use `httptest`.

Tests:
1. **TestRoutes** (table-driven): for paths `GET /`, `GET /classes/{id}`,
   `GET /students/{id}` — expect 200 and that the response body contains a known
   seeded name. For `GET /classes/999` and `GET /students/999` — expect 404.
2. **TestIndexListsClasses** — seed two classes, assert both names appear.
3. **TestStudentPageShowsNotesAndSummary** — seed a note and a summary, assert
   both texts appear in the rendered body.

Notes:
- Seed via the store's exported CRUD methods (`CreateClass`, `CreateStudent`,
  `AppendNote`, `SaveSummary`).
- The new `store.Open` needs a table-driven test too (open a temp path, create a
  row, read it back) so the store package keeps real coverage.

Acceptance: `go test ./...` passes with real assertions.

## Explicitly out of scope (step 5)

- POST handlers for creating classes/students, appending notes, generating or
  saving summaries.
- Calling the `Summarizer` from any handler.
- Editing templates beyond the field-name reconciliation in Task 3.

## Handoff to build

1. `store.Open(dbPath)` + `store.GetClass(id)` (keep `NewStore` behavior).
2. `web.LoadTemplates()` → per-page map.
3. Reconcile `index.html`/`class.html`/`student.html` data shapes.
4. Rewrite `server.Server` + `Run(addr)` with the five routes + 404.
5. Wire `main.go` with `SYLLOGE_ADDR`.
6. Add table-driven `server_test.go` + `store.Open` test.
7. Verify: `go build ./...`, `go vet ./...`, `go test ./...`.