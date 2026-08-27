# Spec: Remove created-at timestamps from student/class cards, add to notes

## Goal
Clean up the UI by removing the creation timestamp badges from student and class cards, and instead show the timestamp on each individual note when it is created.

## Motivation
- The creation date of a class or student record is not useful to the teacher during normal workflow.
- The timestamp is meaningful for **notes** because notes are accumulated over days/months and the teacher may want to see when an observation was recorded.
- Removing the badge makes the class/student cards cleaner and easier to scan.

## Changes required

### 1. `internal/web/templates/class_list.html`
Remove the `<span class="badge">{{.CreatedAt}}</span>` from each class card.

Before:
```html
<a class="class-card" href="/classes/{{.ID}}">
    <span class="class-card-name">{{.Name}}</span>
    <span class="badge">{{.CreatedAt}}</span>
</a>
```

After:
```html
<a class="class-card" href="/classes/{{.ID}}">
    <span class="class-card-name">{{.Name}}</span>
</a>
```

### 2. `internal/web/templates/student_list.html`
Remove the `<span class="badge">{{.CreatedAt}}</span>` from each student card.

Before:
```html
<a class="class-card" href="/students/{{.ID}}">
    <span class="class-card-name">{{.Name}}</span>
    <span class="badge">{{.CreatedAt}}</span>
</a>
```

After:
```html
<a class="class-card" href="/students/{{.ID}}">
    <span class="class-card-name">{{.Name}}</span>
</a>
```

### 3. `internal/web/templates/student.html`
Ensure each note clearly shows its creation timestamp. Current layout already shows `.CreatedAt` in `.meta`, but confirm it is human-readable and styled consistently. No backend change needed.

Current (keep, verify styling):
```html
<article class="note">
    <p>{{.Content}}</p>
    <div class="meta">Added {{.CreatedAt}}</div>
</article>
```

## Optional: improve timestamp display
SQLite `CURRENT_TIMESTAMP` returns ISO-8601 format (e.g. `2026-08-25 20:30:00`). This is readable but not pretty. If desired, format it on the server before rendering, or use a small client-side formatting snippet. For v1, the raw ISO string is acceptable.

## Files to modify
- `internal/web/templates/class_list.html`
- `internal/web/templates/student_list.html`
- `internal/web/templates/student.html` (verification only)

## Acceptance criteria
- `go build ./...`, `go vet ./...`, `go test ./...` pass.
- Existing server tests still pass (may need to update `TestStudentPageShowsNotesAndSummary` if it asserts the timestamp text).
- Home page shows class cards without timestamps.
- Class page shows student cards without timestamps.
- Student page shows each note with its creation timestamp.
- Manual browser check confirms the change.
