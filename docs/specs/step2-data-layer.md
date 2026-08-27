# Spec: Step 2 — Data layer (CRUD + tests)

Status: ready for `build`
Roadmap step: 2 (SQLite schema + store package CRUD)

## Context for the builder

The store package (`internal/store/store.go`) already has:
- A `Store` struct wrapping `*sql.DB`
- `NewStore()` that opens `sylloge.db` (SQLite) and runs `createTables()`
- A working schema: `classes`, `students`, `notes`, `summaries`
- `Close()` and `GetDB()` methods

What's missing: **CRUD methods** and **tests**. This step adds them.

The `Store` struct currently exposes `db` as unexported (`db *sql.DB`). Keep it unexported — all access goes through store methods.

## Existing schema (no changes needed)

```sql
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes (id)
);

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students (id)
);

CREATE TABLE IF NOT EXISTS summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students (id)
);
```

## Task 1 — Define entity types

Add Go structs to `internal/store/store.go` (top of file, above existing code):

```go
type Class struct {
    ID        int64
    Name      string
    CreatedAt string
}

type Student struct {
    ID        int64
    ClassID   int64
    Name      string
    CreatedAt string
}

type Note struct {
    ID        int64
    StudentID int64
    Content   string
    CreatedAt string
}

type Summary struct {
    ID        int64
    StudentID int64
    Content   string
    CreatedAt string
}
```

Acceptance: code compiles (`go build ./...`).

## Task 2 — Add CRUD methods to Store

Add these methods to `internal/store/store.go`. Use the existing `s.db` field for all queries. Use `context.Background()` for all context args.

### Classes
- `ListClasses(ctx) ([]Class, error)` — `SELECT id, name, created_at FROM classes ORDER BY name`
- `CreateClass(ctx, name string) (int64, error)` — `INSERT INTO classes(name) VALUES(?) RETURNING id`
- `DeleteClass(ctx, id int64) (error)` — `DELETE FROM classes WHERE id = ?`

### Students
- `ListStudentsByClass(ctx, classID int64) ([]Student, error)` — `SELECT id, class_id, name, created_at FROM students WHERE class_id = ? ORDER BY name`
- `GetStudent(ctx, id int64) (Student, error)` — `SELECT id, class_id, name, created_at FROM students WHERE id = ?` — return a custom error `ErrNotFound` (defined as `var ErrNotFound = errors.New("not found")`) if no row matches
- `CreateStudent(ctx, classID int64, name string) (int64, error)` — `INSERT INTO students(class_id, name) VALUES(?, ?) RETURNING id`

### Notes
- `ListNotesByStudent(ctx, studentID int64) ([]Note, error)` — `SELECT id, student_id, content, created_at FROM notes WHERE student_id = ? ORDER BY created_at`
- `AppendNote(ctx, studentID int64, content string) (int64, error)` — `INSERT INTO notes(student_id, content) VALUES(?, ?) RETURNING id`

### Summaries
- `GetSummaryByStudent(ctx, studentID int64) (Summary, error)` — `SELECT id, student_id, content, created_at FROM summaries WHERE student_id = ? ORDER BY created_at DESC LIMIT 1` — return `ErrNotFound` if no row matches
- `SaveSummary(ctx, studentID int64, content string) (int64, error)` — `INSERT INTO summaries(student_id, content) VALUES(?, ?) RETURNING id`

Acceptance: code compiles; all methods follow the signatures above.

## Task 3 — Add table-driven tests

Create `internal/store/store_test.go`.

### Test helper

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    // Run the same createTables logic
    if err := createTables(db); err != nil {
        t.Fatalf("createTables: %v", err)
    }
    return &Store{db: db}
}
```

### Tests (table-driven)

1. **TestCreateAndListClasses** — Create 3 classes, ListClasses returns all 3, ordered by name. Verify count and content.

2. **TestCreateClassDuplicate** — Create a class, then create another with the same name. Expect an error (SQLite UNIQUE constraint). Verify the first class still exists.

3. **TestDeleteClass** — Create a class, delete it, ListClasses returns empty.

4. **TestCreateAndListStudents** — Create a class, add 2 students to it. ListStudentsByClass returns both, ordered by name.

5. **TestGetStudent** — Create a student, GetStudent returns correct fields. GetStudent with a nonexistent ID returns `ErrNotFound`.

6. **TestAppendAndListNotes** — Create a student, append 3 notes. ListNotesByStudent returns all 3 in `created_at` order. Verify content of each.

7. **TestSaveAndGetSummary** — Create a student, SaveSummary saves content. GetSummaryByStudent returns the same content. GetSummaryByStudent for a student with no summary returns `ErrNotFound`.

8. **TestMultipleSummaries** — Create a student, save 2 summaries. GetSummaryByStudent returns the most recent (last inserted).

Acceptance: `go test ./internal/store/... -v` passes all tests.

## Task 4 — Verify the tree is clean

Run from repo root:
- `go build ./...` — no output (success)
- `go vet ./...` — no output (success)
- `go test ./internal/store/... -v` — all tests pass

## Explicitly out of scope for this task

- `internal/summarize/ollama.go` hardcoded defaults — fix in step 3.
- `internal/server/server.go` routes — step 4.
- `internal/web` templates — step 4.
- Deleting `GetDB()` or `NewStore()` — keep both, they are used by server (step 4).

## Handoff to build

1. Add the 4 entity types (`Class`, `Student`, `Note`, `Summary`) to `internal/store/store.go`.
2. Add `var ErrNotFound = errors.New("not found")` (needs `"errors"` import).
3. Add all CRUD methods (`ListClasses`, `CreateClass`, `DeleteClass`, `ListStudentsByClass`, `GetStudent`, `CreateStudent`, `ListNotesByStudent`, `AppendNote`, `GetSummaryByStudent`, `SaveSummary`).
4. Create `internal/store/store_test.go` with the helper and 8 table-driven tests.
5. Verify: `go build ./...`, `go vet ./...`, `go test ./internal/store/... -v` all pass.