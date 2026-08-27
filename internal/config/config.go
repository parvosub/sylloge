package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// Database holds SQLite connection settings.
type Database struct {
	Path string `toml:"path"`
}

// LLM holds model selection and provider settings.
type LLM struct {
	Provider     string `toml:"provider"` // "local" (Ollama) or "cloud"
	Model        string `toml:"model"`
	SystemPrompt string `toml:"system_prompt"`
}

// API holds the OpenAI-compatible endpoint settings.
type API struct {
	BaseURL string `toml:"base_url"`
	APIKey  string `toml:"api_key"`
}

// Config is the application configuration loaded from a TOML file.
type Config struct {
	Database Database `toml:"database"`
	LLM      LLM      `toml:"llm"`
	API      API      `toml:"api"`
}

// Default returns a sensible default configuration for a local Ollama setup.
func Default() *Config {
	return &Config{
		Database: Database{Path: "sylloge.db"},
		LLM: LLM{
			Provider:     "local",
			Model:        "qwen3.5:27b-q4_K_M",
			SystemPrompt: "You are a supportive teaching assistant writing a report-card comment for one student.",
		},
		API: API{
			BaseURL: "http://localhost:11434/v1",
			APIKey:  "",
		},
	}
}

// Load reads and parses the TOML config file at path, then validates it.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Database.Path) == "" {
		return fmt.Errorf("config: database.path is required")
	}
	switch c.LLM.Provider {
	case "local", "cloud":
	default:
		return fmt.Errorf("config: llm.provider must be \"local\" or \"cloud\", got %q", c.LLM.Provider)
	}
	if strings.TrimSpace(c.LLM.Model) == "" {
		return fmt.Errorf("config: llm.model is required")
	}
	if strings.TrimSpace(c.API.BaseURL) == "" {
		return fmt.Errorf("config: api.base_url is required")
	}
	if c.LLM.Provider == "cloud" && strings.TrimSpace(c.API.APIKey) == "" {
		return fmt.Errorf("config: api.api_key is required when llm.provider is \"cloud\"")
	}
	return nil
}
