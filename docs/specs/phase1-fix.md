# Spec: Restore clean Phase 1 (project skeleton)

Status: ready for `build`
Roadmap step: 1 (skeleton hygiene + phase re-scoping)

## Context for the builder

The roadmap in `AGENTS.md` defines **Phase 1** as *skeleton only*: Go module,
`cmd/` + `internal/` layout, `.gitignore`, README stub. The current tree has the
skeleton but also (a) two `.gitignore` rules that fight the committed stack, and
(b) partially-stubbed code from later roadmap steps.

This task fixes the phase-1 hygiene faults and leaves steps 2–4 formally **open**.
Do **not** add store CRUD, real Ollama logic, HTTP routes, or template wiring in
this task — those belong to their own roadmap steps.

## Decision (re-scoping steps 2–4)

Keep the already-written files so the tree keeps compiling, but treat them as
**not-yet-accepted placeholders**. Steps 2 (store CRUD), 3 (Summarizer/Ollama),
and 4 (server/templates) remain open and will be done in sequence later. This task
only makes the skeleton correct and removes misleading stubs.

## Tasks

### Task 1 — Fix `.gitignore` (must-fix)
Edit `.gitignore`:
- Delete the `go.sum` line. `go.sum` is a committed lockfile and must be tracked.
- Delete the `_test.go` line. This glob matches *every* test file in the repo and
  would prevent all tests from ever being committed.
- Leave everything else unchanged.

Acceptance: `.gitignore` no longer contains `go.sum` or `_test.go`.

### Task 2 — Reconcile module metadata
- Run `go mod tidy`. This should drop the `// indirect` marker on
  `github.com/mattn/go-sqlite3` in `go.mod` (it is a direct dependency via
  `internal/store/store.go`) and (re)generate `go.sum`.

Acceptance: the sqlite3 line in `go.mod` has **no** `// indirect` comment; `go.sum`
exists on disk and is not gitignored.

### Task 3 — Fix README version claim
In `README.md`, "Getting Started" step 1 says "Go (version 1.21 or higher)".
`go.mod` targets `go 1.24`. Change the README text to `Go 1.24 or higher`.

Acceptance: README and `go.mod` agree on the Go version.

### Task 4 — Remove misleading placeholder tests
Delete `internal/web/web_test.go` and `internal/summarize/ollama_test.go`. Both only
call `t.Log(...)` and assert nothing. Per `AGENTS.md`, real *table-driven* tests get
added when each package gets real logic (their proper roadmap steps). Leaving empty
stubs invites a false sense of coverage.

Acceptance: no `*_test.go` files remain; the two packages still build.

### Task 5 — Verify the skeleton is green
Run, from repo root:
- `go build ./...` → no output (success)
- `go vet ./...` → no output (success)
- `go mod tidy` → no further changes on a second run (tree is stable)

Acceptance: all three clean.

## Explicitly out of scope for this task (do NOT touch)
- `internal/store/store.go` CRUD/schema behavior → **Step 2**.
- `internal/summarize/ollama.go` hardcoded `OLLAMA_HOST:11434` / `llama3` defaults →
  fix as part of **Step 3** ("config via env vars — never hardcode host/model").
  Flagged now so it is not forgotten.
- `internal/server/server.go` routes and `internal/web` template wiring → **Step 4**.

## Handoff checklist
1. Edit `.gitignore`: remove the `go.sum` line and the `_test.go` line; leave the rest.
2. Run `go mod tidy`; confirm sqlite3 loses its `// indirect` marker and `go.sum` is present.
3. In `README.md` Getting Started, change "Go (version 1.21 or higher)" to "Go 1.24 or higher".
4. Delete `internal/web/web_test.go` and `internal/summarize/ollama_test.go`.
5. Verify: `go build ./...`, `go vet ./...` both clean, and a second `go mod tidy` is a no-op.
6. Do not implement store CRUD, Ollama logic, routes, or template wiring — those stay
   open for roadmap steps 2–4.
