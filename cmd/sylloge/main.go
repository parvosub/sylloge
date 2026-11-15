package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/parvosub/sylloge/internal/config"
	"github.com/parvosub/sylloge/internal/server"
	"github.com/parvosub/sylloge/internal/store"
	"github.com/parvosub/sylloge/internal/summarize"
	"github.com/parvosub/sylloge/internal/web"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println("sylloge " + version)
		return
	}

	configPath := os.Getenv("SYLLOGE_CONFIG")
	if configPath == "" {
		configPath = "sylloge.toml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	dbPath, err := filepath.Abs(cfg.Database.Path)
	if err != nil {
		log.Fatalf("resolving database path %q: %v", cfg.Database.Path, err)
	}
	log.Printf("database: %s", dbPath)
	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	tmpl, err := web.LoadTemplates()
	if err != nil {
		log.Fatal(err)
	}

	summarizer, err := summarize.NewOpenAICompatibleSummarizer(summarize.OpenAIConfig{
		BaseURL:      cfg.API.BaseURL,
		Model:        cfg.LLM.Model,
		SystemPrompt: cfg.LLM.SystemPrompt,
		APIKey:       cfg.API.APIKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("SYLLOGE_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := server.NewServer(st, tmpl, summarizer)
	if err := srv.Run(addr); err != nil {
		log.Fatal(err)
	}
}
