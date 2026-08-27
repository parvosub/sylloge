package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/parvosub/sylloge/internal/store"
	"github.com/parvosub/sylloge/internal/web"
)

type fakeSummarizer struct {
	out string
	err error
}

func (f fakeSummarizer) Summarize(notes string) (string, error) {
	return f.out, f.err
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	return newTestServerWithSummarizer(t, fakeSummarizer{out: "default summary"})
}

func newTestServerWithSummarizer(t *testing.T, sum fakeSummarizer) *Server {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	tmpl, err := web.LoadTemplates()
	if err != nil {
		t.Fatalf("web.LoadTemplates: %v", err)
	}
	return NewServer(st, tmpl, sum)
}

func postForm(t *testing.T, path, form string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	return rec, req
}

func TestRoutes(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()

	classID, err := s.store.CreateClass(ctx, "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	studentID, err := s.store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "index", path: "/", wantStatus: http.StatusOK, wantBody: "Art"},
		{name: "class page", path: "/classes/" + strconv.FormatInt(classID, 10), wantStatus: http.StatusOK, wantBody: "Alice"},
		{name: "student page", path: "/students/" + strconv.FormatInt(studentID, 10), wantStatus: http.StatusOK, wantBody: "Alice"},
		{name: "missing class", path: "/classes/999", wantStatus: http.StatusNotFound},
		{name: "missing student", path: "/students/999", wantStatus: http.StatusNotFound},
		{name: "unknown path", path: "/nope", wantStatus: http.StatusNotFound},
	}

	mux := s.routes()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body missing %q, got: %s", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestIndexListsClasses(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()

	if _, err := s.store.CreateClass(ctx, "Painting"); err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	if _, err := s.store.CreateClass(ctx, "Sculpture"); err != nil {
		t.Fatalf("CreateClass: %v", err)
	}

	mux := s.routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, name := range []string{"Painting", "Sculpture"} {
		if !strings.Contains(rec.Body.String(), name) {
			t.Errorf("body missing class %q: %s", name, rec.Body.String())
		}
	}
}

func TestStudentPageShowsNotesAndSummary(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()

	classID, err := s.store.CreateClass(ctx, "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	studentID, err := s.store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}
	if _, err := s.store.AppendNote(ctx, studentID, "Great use of color"); err != nil {
		t.Fatalf("AppendNote: %v", err)
	}
	if _, err := s.store.SaveSummary(ctx, studentID, "Alice shows strong color sense."); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	mux := s.routes()
	req := httptest.NewRequest(http.MethodGet, "/students/"+strconv.FormatInt(studentID, 10), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for _, want := range []string{"Great use of color", "Alice shows strong color sense."} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("body missing %q: %s", want, rec.Body.String())
		}
	}
}

func seedClassAndStudent(t *testing.T, s *Server) (int64, int64) {
	t.Helper()
	ctx := context.Background()
	classID, err := s.store.CreateClass(ctx, "Art")
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	studentID, err := s.store.CreateStudent(ctx, classID, "Alice")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}
	return classID, studentID
}

func TestCreateClassPOST(t *testing.T) {
	s := newTestServer(t)
	mux := s.routes()
	rec, req := postForm(t, "/classes", "name=Math")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Math") {
		t.Errorf("body missing new class: %s", rec.Body.String())
	}
	classes, err := s.store.ListClasses(context.Background())
	if err != nil {
		t.Fatalf("ListClasses: %v", err)
	}
	if len(classes) != 1 {
		t.Errorf("want 1 class, got %d", len(classes))
	}
}

func TestCreateClassEmptyName(t *testing.T) {
	s := newTestServer(t)
	mux := s.routes()
	rec, req := postForm(t, "/classes", "")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAddStudentPOST(t *testing.T) {
	s := newTestServer(t)
	classID, _ := seedClassAndStudent(t, s)
	mux := s.routes()
	rec, req := postForm(t, "/classes/"+strconv.FormatInt(classID, 10)+"/students", "name=Bob")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Bob") {
		t.Errorf("body missing new student: %s", rec.Body.String())
	}
	students, err := s.store.ListStudentsByClass(context.Background(), classID)
	if err != nil {
		t.Fatalf("ListStudentsByClass: %v", err)
	}
	if len(students) != 2 {
		t.Errorf("want 2 students, got %d", len(students))
	}
}

func TestAppendNotePOST(t *testing.T) {
	s := newTestServer(t)
	_, studentID := seedClassAndStudent(t, s)
	mux := s.routes()
	rec, req := postForm(t, "/notes/"+strconv.FormatInt(studentID, 10), "notes=hello")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/students/"+strconv.FormatInt(studentID, 10) {
		t.Errorf("Location = %q, want /students/%d", loc, studentID)
	}
	notes, err := s.store.ListNotesByStudent(context.Background(), studentID)
	if err != nil {
		t.Fatalf("ListNotesByStudent: %v", err)
	}
	if len(notes) != 1 || notes[0].Content != "hello" {
		t.Errorf("want one note %q, got %v", "hello", notes)
	}
}

func TestGenerateSummary(t *testing.T) {
	s := newTestServerWithSummarizer(t, fakeSummarizer{out: "A fine summary"})
	_, studentID := seedClassAndStudent(t, s)
	mux := s.routes()
	rec, req := postForm(t, "/students/"+strconv.FormatInt(studentID, 10)+"/summary", "")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "A fine summary") {
		t.Errorf("body missing summary: %s", rec.Body.String())
	}
	summary, err := s.store.GetSummaryByStudent(context.Background(), studentID)
	if err != nil {
		t.Fatalf("GetSummaryByStudent: %v", err)
	}
	if summary.Content != "A fine summary" {
		t.Errorf("summary = %q, want %q", summary.Content, "A fine summary")
	}
}

func TestGenerateSummaryError(t *testing.T) {
	s := newTestServerWithSummarizer(t, fakeSummarizer{err: errors.New("boom")})
	_, studentID := seedClassAndStudent(t, s)
	mux := s.routes()
	rec, req := postForm(t, "/students/"+strconv.FormatInt(studentID, 10)+"/summary", "")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("body missing error message: %s", rec.Body.String())
	}
}

func TestSummariesHistory(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	_, studentID := seedClassAndStudent(t, s)

	if _, err := s.store.AppendNote(ctx, studentID, "Note one"); err != nil {
		t.Fatalf("AppendNote: %v", err)
	}
	if _, err := s.store.SaveSummary(ctx, studentID, "First summary."); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	if _, err := s.store.SaveSummary(ctx, studentID, "Second summary."); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	mux := s.routes()
	req := httptest.NewRequest(http.MethodGet, "/students/"+strconv.FormatInt(studentID, 10)+"/summaries", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "First summary.") {
		t.Errorf("body missing first summary: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Second summary.") {
		t.Errorf("body missing second summary: %s", rec.Body.String())
	}
}

func TestSummariesHistoryMissingStudent(t *testing.T) {
	s := newTestServer(t)
	mux := s.routes()
	req := httptest.NewRequest(http.MethodGet, "/students/999/summaries", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestStudentPageCollapsesOldNotes(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	_, studentID := seedClassAndStudent(t, s)

	for _, content := range []string{"Note 1", "Note 2", "Note 3", "Note 4", "Note 5"} {
		if _, err := s.store.AppendNote(ctx, studentID, content); err != nil {
			t.Fatalf("AppendNote: %v", err)
		}
	}

	mux := s.routes()
	req := httptest.NewRequest(http.MethodGet, "/students/"+strconv.FormatInt(studentID, 10), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	for _, want := range []string{"Note 5", "Note 4", "Note 3"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing recent note %q: %s", want, body)
		}
	}
	if !strings.Contains(body, `id="older-notes"`) {
		t.Errorf("body missing #older-notes collapse section")
	}
	if !strings.Contains(body, `id="notes-toggle"`) {
		t.Errorf("body missing #notes-toggle button")
	}
	for _, want := range []string{"Note 2", "Note 1"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing older note %q: %s", want, body)
		}
	}
}