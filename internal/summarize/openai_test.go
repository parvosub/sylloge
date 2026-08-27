package summarize

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}
		w.WriteHeader(status)
		io.WriteString(w, body)
	}))
}

func TestSummarizeSuccess(t *testing.T) {
	var authHdr string
	var captured map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Errorf("failed to unmarshal request: %v", err)
		}
		authHdr = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"choices":[{"message":{"content":"Great work"}}]}`)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s, err := NewOpenAICompatibleSummarizer(OpenAIConfig{BaseURL: srv.URL, Model: "report-model", APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleSummarizer: %v", err)
	}

	got, err := s.Summarize("notes about the student")
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if got != "Great work" {
		t.Errorf("expected 'Great work', got %q", got)
	}
	if captured["model"] != "report-model" {
		t.Errorf("expected model 'report-model', got %v", captured["model"])
	}
	messages, ok := captured["messages"].([]interface{})
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %v", captured["messages"])
	}
	userMsg := messages[1].(map[string]interface{})
	if content, _ := userMsg["content"].(string); !strings.Contains(content, "notes about the student") {
		t.Errorf("expected user message to contain notes, got %v", userMsg["content"])
	}
	if authHdr != "Bearer sk-test" {
		t.Errorf("expected Authorization 'Bearer sk-test', got %q", authHdr)
	}
}

func TestSummarizeNoAuthWhenKeyEmpty(t *testing.T) {
	var authHdr string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHdr = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}]}`)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	s, err := NewOpenAICompatibleSummarizer(OpenAIConfig{BaseURL: srv.URL, Model: "m"})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleSummarizer: %v", err)
	}
	if _, err := s.Summarize("n"); err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if authHdr != "" {
		t.Errorf("expected no Authorization header, got %q", authHdr)
	}
}

func TestSummarizeHTTPError(t *testing.T) {
	srv := testServer(t, http.StatusInternalServerError, `boom`)
	defer srv.Close()

	s, err := NewOpenAICompatibleSummarizer(OpenAIConfig{BaseURL: srv.URL, Model: "m"})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleSummarizer: %v", err)
	}

	_, err = s.Summarize("notes")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected error to contain 'boom', got %v", err)
	}
}

func TestSummarizeMalformedJSON(t *testing.T) {
	srv := testServer(t, http.StatusOK, `not-json`)
	defer srv.Close()

	s, err := NewOpenAICompatibleSummarizer(OpenAIConfig{BaseURL: srv.URL, Model: "m"})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleSummarizer: %v", err)
	}

	_, err = s.Summarize("notes")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSummarizeNoChoices(t *testing.T) {
	srv := testServer(t, http.StatusOK, `{"choices":[]}`)
	defer srv.Close()

	s, err := NewOpenAICompatibleSummarizer(OpenAIConfig{BaseURL: srv.URL, Model: "m"})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleSummarizer: %v", err)
	}
	if _, err := s.Summarize("n"); err == nil {
		t.Error("expected error for empty choices, got nil")
	}
}

func TestNewOpenAICompatibleSummarizerValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OpenAIConfig
		wantErr bool
	}{
		{name: "both present", cfg: OpenAIConfig{BaseURL: "http://x", Model: "m"}, wantErr: false},
		{name: "missing base URL", cfg: OpenAIConfig{Model: "m"}, wantErr: true},
		{name: "missing model", cfg: OpenAIConfig{BaseURL: "http://x"}, wantErr: true},
		{name: "both missing", cfg: OpenAIConfig{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOpenAICompatibleSummarizer(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAICompatibleSummarizer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
