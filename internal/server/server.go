package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/parvosub/sylloge/internal/store"
	"github.com/parvosub/sylloge/internal/summarize"
)

// Server represents the HTTP server
type Server struct {
	store      *store.Store
	templates  map[string]*template.Template
	summarizer summarize.Summarizer
}

// NewServer creates a new server instance
func NewServer(st *store.Store, tmpl map[string]*template.Template, summarizer summarize.Summarizer) *Server {
	return &Server{store: st, templates: tmpl, summarizer: summarizer}
}

// Run starts the HTTP server
func (s *Server) Run(addr string) error {
	log.Printf("Starting server on %s", addr)
	srv := &http.Server{Addr: addr, Handler: s.routes()}
	return srv.ListenAndServe()
}

// routes returns the HTTP handler for the application routes.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /classes/{id}", s.handleClass)
	mux.HandleFunc("GET /students/{id}", s.handleStudent)
	mux.HandleFunc("POST /classes", s.handleCreateClass)
	mux.HandleFunc("POST /classes/{id}/students", s.handleAddStudent)
	mux.HandleFunc("POST /notes/{id}", s.handleAppendNote)
	mux.HandleFunc("POST /students/{id}/summary", s.handleGenerateSummary)
	mux.HandleFunc("GET /students/{id}/summaries", s.handleSummariesHistory)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	classes, err := s.store.ListClasses(r.Context())
	if err != nil {
		s.internalError(w, err)
		return
	}
	if err := s.render(w, "index", map[string]any{"Classes": classes}); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleClass(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	class, err := s.store.GetClass(ctx, classID)
	if err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	students, err := s.store.ListStudentsByClass(ctx, classID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	data := map[string]any{
		"ClassID":   class.ID,
		"ClassName": class.Name,
		"Students":  students,
		"Breadcrumb": template.HTML(fmt.Sprintf(`<a href="/">Home</a> &rsaquo; <span class="current">Class: %s</span>`, template.HTMLEscapeString(class.Name))),
	}
	if err := s.render(w, "class", data); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	student, err := s.store.GetStudent(ctx, studentID)
	if err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	class, err := s.store.GetClass(ctx, student.ClassID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	notes, err := s.store.ListNotesByStudent(ctx, studentID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	var recentNotes, olderNotes []store.Note
	if len(notes) > 3 {
		recentNotes = notes[:3]
		olderNotes = notes[3:]
	} else {
		recentNotes = notes
	}
	summary, err := s.store.GetSummaryByStudent(ctx, studentID)
	if err != nil && err != store.ErrNotFound {
		s.internalError(w, err)
		return
	}
	summaries, err := s.store.ListSummariesByStudent(ctx, studentID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	data := map[string]any{
		"StudentID":   student.ID,
		"StudentName": student.Name,
		"ClassID":     class.ID,
		"ClassName":   class.Name,
		"RecentNotes": recentNotes,
		"OlderNotes":  olderNotes,
		"Summary":     summary.Content,
		"SummaryHTML": template.HTML(summaryToHTML(summary.Content)),
		"SavedAt":     summary.CreatedAt,
		"Summaries":   summaries,
		"Breadcrumb": template.HTML(fmt.Sprintf(`<a href="/">Home</a> &rsaquo; <a href="/classes/%d">%s</a> &rsaquo; <span class="current">%s</span>`,
			class.ID, template.HTMLEscapeString(class.Name), template.HTMLEscapeString(student.Name))),
	}
	if err := s.render(w, "student", data); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleCreateClass(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "class name is required", http.StatusBadRequest)
		return
	}
	if _, err := s.store.CreateClass(ctx, name); err != nil {
		s.internalError(w, err)
		return
	}
	classes, err := s.store.ListClasses(ctx)
	if err != nil {
		s.internalError(w, err)
		return
	}
	if err := s.renderPartial(w, "index", "classList", map[string]any{"Classes": classes}); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleAddStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	if _, err := s.store.GetClass(ctx, classID); err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "student name is required", http.StatusBadRequest)
		return
	}
	if _, err := s.store.CreateStudent(ctx, classID, name); err != nil {
		s.internalError(w, err)
		return
	}
	students, err := s.store.ListStudentsByClass(ctx, classID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	if err := s.renderPartial(w, "class", "studentList", map[string]any{"Students": students}); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleAppendNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	if _, err := s.store.GetStudent(ctx, studentID); err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	notes := strings.TrimSpace(r.FormValue("notes"))
	if notes == "" {
		http.Error(w, "notes are required", http.StatusBadRequest)
		return
	}
	if _, err := s.store.AppendNote(ctx, studentID, notes); err != nil {
		s.internalError(w, err)
		return
	}
	http.Redirect(w, r, "/students/"+strconv.FormatInt(studentID, 10), http.StatusSeeOther)
}

func (s *Server) handleGenerateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	student, err := s.store.GetStudent(ctx, studentID)
	if err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	notes, err := s.store.ListNotesByStudent(ctx, studentID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	var sb strings.Builder
	for _, n := range notes {
		sb.WriteString(n.Content)
		sb.WriteString("\n")
	}
	summary, err := s.summarizer.Summarize(sb.String())
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		if renderErr := s.renderPartial(w, "student", "summary", map[string]any{"Error": err.Error(), "StudentName": student.Name}); renderErr != nil {
			s.internalError(w, renderErr)
		}
		return
	}
	sm, err := s.store.SaveSummary(ctx, studentID, summary)
	if err != nil {
		s.internalError(w, err)
		return
	}
	if err := s.renderPartial(w, "student", "summary", map[string]any{
		"Summary":     sm.Content,
		"SummaryHTML": template.HTML(summaryToHTML(sm.Content)),
		"StudentID":   studentID,
		"StudentName": student.Name,
		"SavedAt":     sm.CreatedAt,
	}); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) handleSummariesHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	student, err := s.store.GetStudent(ctx, studentID)
	if err != nil {
		if err == store.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		s.internalError(w, err)
		return
	}
	class, err := s.store.GetClass(ctx, student.ClassID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	summaries, err := s.store.ListSummariesByStudent(ctx, studentID)
	if err != nil {
		s.internalError(w, err)
		return
	}
	data := map[string]any{
		"StudentID":   student.ID,
		"StudentName": student.Name,
		"ClassID":     class.ID,
		"ClassName":   class.Name,
		"Summaries":   summaries,
		"Breadcrumb": template.HTML(fmt.Sprintf(`<a href="/">Home</a> &rsaquo; <a href="/classes/%d">%s</a> &rsaquo; <a href="/students/%d">%s</a> &rsaquo; <span class="current">Summary History</span>`,
			class.ID, template.HTMLEscapeString(class.Name), student.ID, template.HTMLEscapeString(student.Name))),
	}
	if err := s.render(w, "summaries_history", data); err != nil {
		s.internalError(w, err)
	}
}

func (s *Server) render(w http.ResponseWriter, page string, data any) error {
	t, ok := s.templates[page]
	if !ok {
		return http.ErrAbortHandler
	}
	return t.ExecuteTemplate(w, "base.html", data)
}

func (s *Server) renderPartial(w http.ResponseWriter, page, block string, data any) error {
	t, ok := s.templates[page]
	if !ok {
		return http.ErrAbortHandler
	}
	return t.ExecuteTemplate(w, block, data)
}

func (s *Server) internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}