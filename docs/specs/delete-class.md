# Spec: Delete a class from the Classes dashboard

## Goal
Let the teacher delete a class from the Classes page. Clicking **Delete Class**, next to
**Add Class**, opens a small card with a dropdown of existing classes. The teacher picks one,
confirms, and the class is removed along with all of its students, notes and summaries.

This is the first "clear data" feature. Deleting students, notes or summaries on their own is
left for later specs.

## Motivation
- The teacher can't currently remove a class made by mistake, or a class from a past term.
- `store.DeleteClass` already exists but only deletes the `classes` row. The schema has no
  `ON DELETE CASCADE`, and `PRAGMA foreign_keys` is off, so this would leave orphaned students,
  notes and summaries behind.

## Changes required

### 1. `internal/store/store.go`
Rewrite `DeleteClass(ctx, id)` to run inside one transaction (`s.db.BeginTx`):
1. `DELETE FROM summaries WHERE student_id IN (SELECT id FROM students WHERE class_id = ?)`
2. `DELETE FROM notes WHERE student_id IN (SELECT id FROM students WHERE class_id = ?)`
3. `DELETE FROM students WHERE class_id = ?`
4. `DELETE FROM classes WHERE id = ?`. If `RowsAffected() == 0`, roll back and return `ErrNotFound`.
5. Commit.

Don't change the schema or add migrations.

### 2. `internal/server/server.go`
- Route: `POST /classes/delete` → `handleDeleteClass`.
- `handleDeleteClass` parses the `class_id` form field (int64). Invalid input returns 400 and
  `ErrNotFound` returns 404. On success it re-lists the classes and renders the `classListResponse`
  block with `OOB: true` and `Toast: "Class deleted"`.
- `handleCreateClass` renders `classListResponse` too (with `OOB: true`), so the delete dropdown
  stays in sync when a class is added.

### 3. `internal/web/templates/class_list.html`
- New block `deleteClassSelect`: `<select id="delete-class-select" name="class_id" required>`,
  with `hx-swap-oob="true"` when `.OOB` is set. It has a placeholder option with an empty value and
  one option per class, or a single disabled "No classes to delete" option when the list is empty.
- New block `classListResponse`: an optional `<span hidden data-toast="...">`, then `classList`,
  then `deleteClassSelect`.

### 4. `internal/web/templates/index.html`
- Header buttons sit in a `.button-row`: **Add Class** (primary) and **Delete Class** (`btn-danger`).
  Each one shows its own card and hides the other.
- Hidden `#delete-class-form` card with a form:
  `hx-post="/classes/delete" hx-target="#class-list" hx-swap="innerHTML"` and an `hx-confirm` warning
  that all students, notes and summaries will be deleted and that this can't be undone. The card
  hides itself after a successful request.

### 5. `internal/web/templates/base.html`
- Style `select` to match `input[type="text"]`.

## Out of scope
Auth, undo or soft-delete, and deleting students, notes or summaries on their own.

## Acceptance criteria
- `go build ./...`, `go vet ./...`, `go test ./...` pass.
- Store test: deleting a class removes its students, notes and summaries and leaves other classes
  alone. Deleting a missing id returns `ErrNotFound`.
- Server test: `POST /classes/delete` returns 200 and removes the class, returns 400 for a
  non-numeric id, and returns 404 for an unknown id. The `POST /classes` response includes the
  updated dropdown.
- Manual check: the class card and dropdown entry disappear, a "Class deleted" toast shows, and
  cancelling the confirm dialog deletes nothing.
