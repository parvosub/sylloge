package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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

func TestOpen(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	id, err := s.CreateClass(context.Background(), "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	c, err := s.GetClass(context.Background(), id)
	if err != nil {
		t.Fatalf("GetClass: %v", err)
	}
	if c.Name != "Art" {
		t.Errorf("expected class 'Art', got %q", c.Name)
	}
}

func TestGetClass(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateClass(ctx, "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	c, err := s.GetClass(ctx, id)
	if err != nil {
		t.Fatalf("GetClass: %v", err)
	}
	if c.Name != "Art" {
		t.Errorf("expected class 'Art', got %q", c.Name)
	}

	_, err = s.GetClass(ctx, 999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateAndListClasses(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create 3 classes
	_, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	_, err = store.CreateClass(ctx, "Science")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	_, err = store.CreateClass(ctx, "English")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// List classes
	classes, err := store.ListClasses(ctx)
	if err != nil {
		t.Fatalf("ListClasses: %v", err)
	}

	// Verify count and content
	if len(classes) != 3 {
		t.Errorf("Expected 3 classes, got %d", len(classes))
	}

	// Check that all classes are present and ordered by name
	expectedNames := []string{"English", "Math", "Science"}
	for i, name := range expectedNames {
		if classes[i].Name != name {
			t.Errorf("Expected class %d to be %s, got %s", i, name, classes[i].Name)
		}
	}

	// IDs are assigned in insertion order (Math=1, Science=2, English=3),
	// independent of the name-sorted listing order. Verify all three exist.
	seen := map[int64]bool{}
	for _, c := range classes {
		seen[c.ID] = true
	}
	for id := int64(1); id <= 3; id++ {
		if !seen[id] {
			t.Errorf("Expected class ID %d to be present, got IDs %v", id, classes)
		}
	}
}

func TestCreateClassDuplicate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	_, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Try to create another class with the same name
	_, err = store.CreateClass(ctx, "Math")
	if err == nil {
		t.Error("Expected error when creating duplicate class")
	}

	// Verify the first class still exists
	classes, err := store.ListClasses(ctx)
	if err != nil {
		t.Fatalf("ListClasses: %v", err)
	}
	if len(classes) != 1 {
		t.Errorf("Expected 1 class after duplicate attempt, got %d", len(classes))
	}
	if classes[0].Name != "Math" {
		t.Errorf("Expected class name 'Math', got %s", classes[0].Name)
	}
}

func TestDeleteClass(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Delete the class
	err = store.DeleteClass(ctx, classID)
	if err != nil {
		t.Fatalf("DeleteClass: %v", err)
	}

	// List classes should return empty
	classes, err := store.ListClasses(ctx)
	if err != nil {
		t.Fatalf("ListClasses: %v", err)
	}
	if len(classes) != 0 {
		t.Errorf("Expected 0 classes after deletion, got %d", len(classes))
	}
}

func TestCreateAndListStudents(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Create 2 students
	store.CreateStudent(ctx, classID, "Alice")
	store.CreateStudent(ctx, classID, "Bob")

	// List students by class
	students, err := store.ListStudentsByClass(ctx, classID)
	if err != nil {
		t.Fatalf("ListStudentsByClass: %v", err)
	}

	// Verify count and content
	if len(students) != 2 {
		t.Errorf("Expected 2 students, got %d", len(students))
	}

	// Check that students are ordered by name
	expectedNames := []string{"Alice", "Bob"}
	for i, name := range expectedNames {
		if students[i].Name != name {
			t.Errorf("Expected student %d to be %s, got %s", i, name, students[i].Name)
		}
		if students[i].ClassID != classID {
			t.Errorf("Expected student %d to have class ID %d, got %d", i, classID, students[i].ClassID)
		}
	}
}

func TestGetStudent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Create a student
	studentID, err := store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Get the student
	student, err := store.GetStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("GetStudent: %v", err)
	}

	// Verify fields
	if student.ID != studentID {
		t.Errorf("Expected student ID %d, got %d", studentID, student.ID)
	}
	if student.ClassID != classID {
		t.Errorf("Expected student class ID %d, got %d", classID, student.ClassID)
	}
	if student.Name != "Alice" {
		t.Errorf("Expected student name 'Alice', got %s", student.Name)
	}

	// Try to get a non-existent student
	_, err = store.GetStudent(ctx, 999)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent student, got %v", err)
	}
}

func TestAppendAndListNotes(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Create a student
	studentID, err := store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Append 3 notes
	store.AppendNote(ctx, studentID, "First note")
	store.AppendNote(ctx, studentID, "Second note")
	store.AppendNote(ctx, studentID, "Third note")

	// List notes by student
	notes, err := store.ListNotesByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("ListNotesByStudent: %v", err)
	}

	// Verify count and content
	if len(notes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(notes))
	}

	// Notes are returned newest first.
	expectedContents := []string{"Third note", "Second note", "First note"}
	for i, content := range expectedContents {
		if notes[i].Content != content {
			t.Errorf("Expected note %d content '%s', got '%s'", i, content, notes[i].Content)
		}
		if notes[i].StudentID != studentID {
			t.Errorf("Expected note %d to have student ID %d, got %d", i, studentID, notes[i].StudentID)
		}
	}
}

func TestSaveAndGetSummary(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Create a student
	studentID, err := store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Save a summary
	summaryContent := "Alice is a good student."
	saved, err := store.SaveSummary(ctx, studentID, summaryContent)
	if err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	// Get the summary
	summary, err := store.GetSummaryByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("GetSummaryByStudent: %v", err)
	}

	// Verify content
	if summary.ID != saved.ID {
		t.Errorf("Expected summary ID %d, got %d", saved.ID, summary.ID)
	}
	if summary.StudentID != studentID {
		t.Errorf("Expected summary student ID %d, got %d", studentID, summary.StudentID)
	}
	if summary.Content != summaryContent {
		t.Errorf("Expected summary content '%s', got '%s'", summaryContent, summary.Content)
	}

	// Try to get a summary for a student with no summary
	studentID2, err := store.CreateStudent(ctx, classID, "Bob")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	_, err = store.GetSummaryByStudent(ctx, studentID2)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for student with no summary, got %v", err)
	}
}

func TestMultipleSummaries(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a class
	classID, err := store.CreateClass(ctx, "Math")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	// Create a student
	studentID, err := store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Save 2 summaries
	summary1Content := "First summary."
	summary2Content := "Second summary (most recent)."
	saved1, err := store.SaveSummary(ctx, studentID, summary1Content)
	if err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	saved2, err := store.SaveSummary(ctx, studentID, summary2Content)
	if err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	if saved1.ID >= saved2.ID {
		t.Errorf("Expected second summary ID (%d) to be greater than first (%d)", saved2.ID, saved1.ID)
	}

	// Get the summary - should get the most recent one
	summary, err := store.GetSummaryByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("GetSummaryByStudent: %v", err)
	}

	// Verify we get the most recent summary
	if summary.Content != summary2Content {
		t.Errorf("Expected most recent summary content '%s', got '%s'", summary2Content, summary.Content)
	}
}

func TestListSummariesByStudent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	classID, err := store.CreateClass(ctx, "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	studentID, err := store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	contents := []string{"First summary.", "Second summary.", "Third summary."}
	var savedIDs []int64
	for _, content := range contents {
		sm, err := store.SaveSummary(ctx, studentID, content)
		if err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
		if sm.ID == 0 {
			t.Errorf("Expected non-zero summary ID for %q", content)
		}
		if sm.Content != content {
			t.Errorf("Expected saved content %q, got %q", content, sm.Content)
		}
		savedIDs = append(savedIDs, sm.ID)
	}

	summaries, err := store.ListSummariesByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("ListSummariesByStudent: %v", err)
	}
	if len(summaries) != len(contents) {
		t.Fatalf("Expected %d summaries, got %d", len(contents), len(summaries))
	}

	for i, sm := range summaries {
		wantIdx := len(contents) - 1 - i
		wantContent := contents[wantIdx]
		if sm.Content != wantContent {
			t.Errorf("Summary %d: expected content %q, got %q", i, wantContent, sm.Content)
		}
		if sm.ID != savedIDs[wantIdx] {
			t.Errorf("Summary %d: expected ID %d, got %d", i, savedIDs[wantIdx], sm.ID)
		}
		if sm.StudentID != studentID {
			t.Errorf("Summary %d: expected student ID %d, got %d", i, studentID, sm.StudentID)
		}
	}

	// A student with no summaries should return an empty slice, not an error.
	otherID, err := store.CreateStudent(ctx, classID, "Bob")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}
	summaries, err = store.ListSummariesByStudent(ctx, otherID)
	if err != nil {
		t.Fatalf("ListSummariesByStudent: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("Expected 0 summaries for new student, got %d", len(summaries))
	}
}