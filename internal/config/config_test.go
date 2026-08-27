package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sylloge.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadValidLocal(t *testing.T) {
	path := writeTempConfig(t, `
[database]
path = "data.db"

[llm]
provider = "local"
model = "qwen3"
system_prompt = "hi"

[api]
base_url = "http://ollama:11434/v1"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Database.Path != "data.db" {
		t.Errorf("Database.Path = %q, want data.db", cfg.Database.Path)
	}
	if cfg.LLM.Provider != "local" {
		t.Errorf("LLM.Provider = %q, want local", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "qwen3" {
		t.Errorf("LLM.Model = %q, want qwen3", cfg.LLM.Model)
	}
	if cfg.API.BaseURL != "http://ollama:11434/v1" {
		t.Errorf("API.BaseURL = %q", cfg.API.BaseURL)
	}
	if cfg.API.APIKey != "" {
		t.Errorf("API.APIKey = %q, want empty", cfg.API.APIKey)
	}
}

func TestLoadValidCloud(t *testing.T) {
	path := writeTempConfig(t, `
[database]
path = "data.db"

[llm]
provider = "cloud"
model = "gpt-4o"

[api]
base_url = "https://api.example.com/v1"
api_key = "sk-123"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LLM.Provider != "cloud" {
		t.Errorf("LLM.Provider = %q, want cloud", cfg.LLM.Provider)
	}
	if cfg.API.APIKey != "sk-123" {
		t.Errorf("API.APIKey = %q, want sk-123", cfg.API.APIKey)
	}
}

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "missing database path", content: `[llm]
provider = "local"
model = "qwen3"
[api]
base_url = "http://x"
`, wantErr: true},
		{name: "bad provider", content: `[database]
path = "d"
[llm]
provider = "edge"
model = "qwen3"
[api]
base_url = "http://x"
`, wantErr: true},
		{name: "missing model", content: `[database]
path = "d"
[llm]
provider = "local"
[api]
base_url = "http://x"
`, wantErr: true},
		{name: "missing base url", content: `[database]
path = "d"
[llm]
provider = "local"
model = "qwen3"
`, wantErr: true},
		{name: "cloud missing key", content: `[database]
path = "d"
[llm]
provider = "cloud"
model = "qwen3"
[api]
base_url = "http://x"
`, wantErr: true},
		{name: "empty file", content: ``, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempConfig(t, tt.content)
			_, err := Load(path)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Database.Path != "sylloge.db" {
		t.Errorf("Default Database.Path = %q, want sylloge.db", cfg.Database.Path)
	}
	if cfg.LLM.Provider != "local" {
		t.Errorf("Default LLM.Provider = %q, want local", cfg.LLM.Provider)
	}
	if cfg.API.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("Default API.BaseURL = %q", cfg.API.BaseURL)
	}
	if err := cfg.validate(); err != nil {
		t.Errorf("Default should validate cleanly, got %v", err)
	}
}
