package store

import (
	"context"
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("not found")

// Class represents a class
type Class struct {
	ID        int64
	Name      string
	CreatedAt string
}

// Student represents a student
type Student struct {
	ID        int64
	ClassID   int64
	Name      string
	CreatedAt string
}

// Note represents a note
type Note struct {
	ID        int64
	StudentID int64
	Content   string
	CreatedAt string
}

// Summary represents a summary
type Summary struct {
	ID        int64
	StudentID int64
	Content   string
	CreatedAt string
}

// Store represents a database store
type Store struct {
	db *sql.DB
}

// NewStore creates a new store backed by the default sylloge.db file
func NewStore() (*Store, error) {
	return Open("sylloge.db")
}

// Open creates a store backed by the SQLite file at dbPath
func Open(dbPath string) (*Store, error) {
	// Create database file if it doesn't exist
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// Create the file
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// createTables creates the necessary database tables
func createTables(db *sql.DB) error {
	query := `
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
	);`

	_, err := db.Exec(query)
	return err
}

// ListClasses returns all classes ordered by name
func (s *Store) ListClasses(ctx context.Context) ([]Class, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, created_at FROM classes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

// CreateClass creates a new class and returns its ID
func (s *Store) CreateClass(ctx context.Context, name string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO classes(name) VALUES(?) RETURNING id", name).Scan(&id)
	return id, err
}

// DeleteClass deletes a class by ID
func (s *Store) DeleteClass(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM classes WHERE id = ?", id)
	return err
}

// DeleteAllNotesByStudent deletes all notes for a student
func (s *Store) DeleteAllNotesByStudent(ctx context.Context, studentID int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM notes WHERE student_id = ?", studentID)
	return err
}

// DeleteAllSummariesByStudent deletes all summaries for a student
func (s *Store) DeleteAllSummariesByStudent(ctx context.Context, studentID int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM summaries WHERE student_id = ?", studentID)
	return err
}

// ListStudentsByClass returns all students in a class ordered by name
func (s *Store) ListStudentsByClass(ctx context.Context, classID int64) ([]Student, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, class_id, name, created_at FROM students WHERE class_id = ? ORDER BY name", classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var st Student
		if err := rows.Scan(&st.ID, &st.ClassID, &st.Name, &st.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, st)
	}
	return students, nil
}

// GetClass returns a class by ID
func (s *Store) GetClass(ctx context.Context, id int64) (Class, error) {
	var c Class
	err := s.db.QueryRowContext(ctx, "SELECT id, name, created_at FROM classes WHERE id = ?", id).Scan(&c.ID, &c.Name, &c.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c, ErrNotFound
		}
		return c, err
	}
	return c, nil
}

// GetStudent returns a student by ID
func (s *Store) GetStudent(ctx context.Context, id int64) (Student, error) {
	var st Student
	err := s.db.QueryRowContext(ctx, "SELECT id, class_id, name, created_at FROM students WHERE id = ?", id).Scan(&st.ID, &st.ClassID, &st.Name, &st.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return st, ErrNotFound
		}
		return st, err
	}
	return st, nil
}

// CreateStudent creates a new student and returns its ID
func (s *Store) CreateStudent(ctx context.Context, classID int64, name string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO students(class_id, name) VALUES(?, ?) RETURNING id", classID, name).Scan(&id)
	return id, err
}

// ListNotesByStudent returns all notes for a student ordered by newest first.
func (s *Store) ListNotesByStudent(ctx context.Context, studentID int64) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, student_id, content, created_at FROM notes WHERE student_id = ? ORDER BY id DESC", studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.StudentID, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// AppendNote appends a note for a student and returns its ID
func (s *Store) AppendNote(ctx context.Context, studentID int64, content string) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO notes(student_id, content) VALUES(?, ?) RETURNING id", studentID, content).Scan(&id)
	return id, err
}

// GetSummaryByStudent returns the most recent summary for a student
func (s *Store) GetSummaryByStudent(ctx context.Context, studentID int64) (Summary, error) {
	var sm Summary
	err := s.db.QueryRowContext(ctx, "SELECT id, student_id, content, created_at FROM summaries WHERE student_id = ? ORDER BY id DESC LIMIT 1", studentID).Scan(&sm.ID, &sm.StudentID, &sm.Content, &sm.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return sm, ErrNotFound
		}
		return sm, err
	}
	return sm, nil
}

// ListSummariesByStudent returns all saved summaries for a student, newest first.
func (s *Store) ListSummariesByStudent(ctx context.Context, studentID int64) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, student_id, content, created_at FROM summaries WHERE student_id = ? ORDER BY id DESC", studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []Summary
	for rows.Next() {
		var sm Summary
		if err := rows.Scan(&sm.ID, &sm.StudentID, &sm.Content, &sm.CreatedAt); err != nil {
			return nil, err
		}
		summaries = append(summaries, sm)
	}
	return summaries, rows.Err()
}

// SaveSummary saves a summary for a student and returns the saved summary.
func (s *Store) SaveSummary(ctx context.Context, studentID int64, content string) (Summary, error) {
	var sm Summary
	err := s.db.QueryRowContext(ctx, "INSERT INTO summaries(student_id, content) VALUES(?, ?) RETURNING id, student_id, content, created_at", studentID, content).Scan(&sm.ID, &sm.StudentID, &sm.Content, &sm.CreatedAt)
	return sm, err
}

// GetDB returns the underlying database connection
func (s *Store) GetDB() *sql.DB {
	return s.db
}
