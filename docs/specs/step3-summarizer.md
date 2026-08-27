# Spec: Step 3 — Summarizer interface + Ollama impl (env config)

Status: ready for `build`
Roadmap step: 3

## Context for the builder

`internal/summarize/` already has:
- `summarizer.go` — the `Summarizer` interface:
  ```go
  type Summarizer interface {
      Summarize(notes string) (string, error)
  }
  ```
  Keep this unchanged.
- `ollama.go` — `OllamaSummarizer` that POSTs to `/api/generate`. **Problem**: it
  hardcodes defaults `http://OLLAMA_HOST:11434` and `llama3` when env vars are
  missing. Per AGENTS.md, "config via env vars — never hardcode host/model".

This step removes the hardcoded defaults, moves config loading behind an
env-based constructor, and adds table-driven tests using an `httptest` server
(no real API calls).

## Task 1 — Make the constructor config-driven

Rewrite `NewOllamaSummarizer` in `internal/summarize/ollama.go`:

```go
type OllamaConfig struct {
    BaseURL      string // required, e.g. http://host:11434
    Model        string // required, e.g. my-report-model
    SystemPrompt string // optional; empty allowed
}
```

- `func NewOllamaSummarizer(cfg OllamaConfig) (*OllamaSummarizer, error)` —
  validate: `cfg.BaseURL` and `cfg.Model` must be non-empty; otherwise return a
  descriptive error. Store the config on the struct.
- `func OllamaSummarizerFromEnv() (*OllamaSummarizer, error)` — reads
  `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, and `OLLAMA_SYSTEM_PROMPT` (optional) from
  the environment and calls `NewOllamaSummarizer`. No hardcoded fallbacks.

The `Summarize` method body (POST `/api/generate`, parse `{"response": "..."}`)
stays as-is. Only the constructor/validation changes.

Acceptance: no `OLLAMA_HOST` or `llama3` string literal remains in any source
file. `go build ./...` passes.

## Task 2 — Keep HTTP client injectable

The struct already has a `client *http.Client`. Add it to `OllamaConfig`:

```go
type OllamaConfig struct {
    BaseURL      string
    Model        string
    SystemPrompt string
    Client       *http.Client // optional; defaults to &http.Client{}
}
```

If `cfg.Client == nil`, use `&http.Client{}`. This lets tests point at an
`httptest.Server`.

## Task 3 — Table-driven tests

Create `internal/summarize/ollama_test.go` with a test helper:

```go
func testServer(t *testing.T, status int, body string) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // assert method POST, path /api/generate
        w.WriteHeader(status)
        w.Write([]byte(body))
    }))
}
```

Tests (table-driven where sensible):

1. **TestSummarizeSuccess** — server returns `{"response": "Great work"}` with
   200. `Summarize("notes")` returns `"Great work"`, nil error. Also verify the
   request body contains the model name and the notes text (read `r.Body` in the
   handler and record it).

2. **TestSummarizeHTTPError** — server returns 500 with body `"boom"`. Expect a
   non-nil error containing "boom".

3. **TestSummarizeMalformedJSON** — server returns 200 with body `not-json`.
   Expect a non-nil error.

4. **TestNewOllamaSummarizerValidation** — table-driven: missing BaseURL,
   missing Model, both present → expect error / success respectively.

5. **TestOllamaSummarizerFromEnv** — set `OLLAMA_BASE_URL`/`OLLAMA_MODEL`,
   call, expect non-nil; then clear env (use `t.Setenv`), expect error when
   required vars missing.

Acceptance: `go test ./internal/summarize/... -v` passes. Every test asserts
something (no `t.Log`-only stubs).

## Task 4 — Verify the tree is clean

From repo root:
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all pass

## Explicitly out of scope

- `internal/server`, `internal/web` — step 4.
- `internal/store` — already done in step 2.
- Wiring the summarizer into the HTTP handler — step 4/5.

## Handoff to build

1. Rework `NewOllamaSummarizer` to take `OllamaConfig`, validate required
   fields, default the client.
2. Add `OllamaSummarizerFromEnv` reading `OLLAMA_BASE_URL`/`OLLAMA_MODEL`/
   `OLLAMA_SYSTEM_PROMPT`, no hardcoded fallbacks.
3. Delete the hardcoded-default constructor logic.
4. Write `internal/summarize/ollama_test.go` with the 5 tests above.
5. Verify: `go build ./...`, `go vet ./...`, `go test ./...`.