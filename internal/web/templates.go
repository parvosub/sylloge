package web

// This file is used to embed the HTML templates in the binary
import (
	"embed"
	"html/template"
	"regexp"
	"strings"
	"time"
)

//go:embed templates/*
var Templates embed.FS

// formatDate parses an RFC3339 timestamp and formats it as a local American-style date.
func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Local().Format("01/02/2006 3:04 PM")
}

var filenameCleaner = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

// sanitizeFilename converts a student name into a safe filename for downloading.
func sanitizeFilename(name string) string {
	cleaned := filenameCleaner.ReplaceAllString(name, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "summary"
	}
	return cleaned + "-summary.txt"
}

// LoadTemplates loads and parses the HTML templates, one template set per page.
// Each set parses base.html, all shared partials, plus exactly one page file,
// avoiding "main" redefinition conflicts.
func LoadTemplates() (map[string]*template.Template, error) {
	pages := []string{"index", "class", "student", "summaries_history"}
	partials := []string{"class_list", "student_list", "summary_partial", "history_partial"}
	funcMap := template.FuncMap{
		"formatDate":       formatDate,
		"sanitizeFilename": sanitizeFilename,
	}
	result := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		files := []string{"templates/base.html"}
		for _, p := range partials {
			files = append(files, "templates/"+p+".html")
		}
		files = append(files, "templates/"+page+".html")
		t, err := template.New(page).Funcs(funcMap).ParseFS(Templates, files...)
		if err != nil {
			return nil, err
		}
		result[page] = t
	}
	return result, nil
}