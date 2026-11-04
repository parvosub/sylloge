# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Read `AGENTS.md` first: project scope, the v1 "Do not" list (no auth, no external
access, no multi-student batching, no LLM settings in the UI, no JS framework or
build pipeline), and the spec → build → review workflow (`docs/specs/`) all live there.

## Commands

```sh
go build ./...                      # CI runs build, vet, test on every push
go vet ./...
go test ./...
go test ./internal/server -run TestGenerateSummary   # single test
cp sylloge.toml.example sylloge.toml && go run ./cmd/sylloge   # http://localhost:8080
go run ./cmd/sylloge --version
```

`mattn/go-sqlite3` needs cgo, so builds and tests need `CGO_ENABLED=1` and a C compiler.

## Architecture

Request flow: `cmd/sylloge/main.go` loads `internal/config` (TOML) → opens
`internal/store` (SQLite) → parses `internal/web` templates → builds the
`summarize.Summarizer` → hands all three to `internal/server`.

- **Config**: `sylloge.toml` (path override: `SYLLOGE_CONFIG`) is the only
  source for DB path, LLM provider/model/system prompt, and API base_url/key.
  `SYLLOGE_ADDR` is the only other env var the code reads (`SYLLOGE_DB` in the
  Dockerfile/compose is not read). `config.validate` enforces
  provider ∈ {local, cloud}, and that cloud requires api_key.
- **Store**: schema is created with `CREATE TABLE IF NOT EXISTS` in
  `createTables`; there is no migration system. Notes are append-only;
  every generated summary is a new row (latest = highest id), which is
  what powers summary history.
- **Summarizer**: one-method interface (`Summarize(notes string)`);
  `OpenAICompatibleSummarizer` POSTs to `{base_url}/chat/completions`.
  The server concatenates *all* of a student's notes and sends them in one call.
- **Templates**: `web.LoadTemplates` builds one template set per page
  (base.html + every partial + that page) so each page can define `main`.
  New pages must be added to the `pages` list; new partials to `partials`.
  Full pages render `base.html`; HTMX endpoints render a named block via
  `renderPartial(page, block)`. For example, `POST /students/{id}/summary` → the `summary`
  block in `summary_partial.html`, swapped into `#summary-section`.
- **Markdown → HTML**: `internal/server/markdown.go` (`summaryToHTML`) is a
  small, model-output-specific converter for the `contenteditable` summary box.
  It flattens bullets into prose and keeps only `**bold**`/headings; everything else
  is HTML-escaped. Covered by a table-driven test in `markdown_test.go`.
- **Tests**: server tests inject a `fakeSummarizer` and use a real SQLite
  file in `t.TempDir()`. Summarize tests stand up an `httptest` server as
  the OpenAI-compatible endpoint. No external services are needed.

## Release

Pushing a `v*` tag runs `.github/workflows/release.yml`: it builds linux/darwin ×
amd64/arm64 binaries with `-X main.version=<tag>` plus `.sha256` files, and pushes
a GHCR image (`:latest` and `:<tag>`, linux/amd64 only).
