# AGENTS.md

## Project

Grade and comment website for teachers. Live public repo:
https://github.com/parvosub/sylloge (MIT, owner `parvosub`). v1 shipped as
release `v1.0.0` (binaries + GHCR Docker image).

Core flow:
1. Teacher enters free-form, unstructured notes about a student's work through the web front end.
2. An AI model transforms those notes into a coherent summary of the student's work.

## Status: v1 shipped

## Users & context

- **Single user**: one teacher, non-technical. The UI must stay dead simple.
- **Scale**: 1 teacher, a handful of classes (entered by teacher), multiple students per class (entered by teacher).
- **Environment**: deployed on a local server in a homelab, accessed via a web browser on the LAN.
- **Workflow**: the teacher adds notes over time (days/months); at end of semester/year the
  generated summary is copied out and pasted into the school's website.
- **Owner**: operations background. The repo doubles as a portfolio piece, so
  code quality and documentation matter.

## v1 scope

- Flow: teacher types/pastes free-form notes for **one student** → AI returns a coherent summary →
  summary lands in an **editable** text box → teacher copies it out.
- **Per-student**: one student, one summary at a time.
- **Persistence**: classes, students, notes, and summaries are saved in SQLite and persist across
  days/months so the teacher can revisit and edit.
- **No auth** for v1 (single user on trusted LAN).
- Teacher never touches LLM settings.

## Committed stack

- **Backend**: Go. Single static binary, trivial Docker deploy.
  Renders HTML server-side via `html/template`.
- **Frontend**: server-rendered HTML + **HTMX** for the dynamic submit→summary interaction. No SPA,
  no JS build pipeline.
- **Storage**: **SQLite** (single file, backed up to NAS). Schema: `classes` → `students` →
  `notes` / `summaries`.
- **AI**: any **OpenAI-compatible** endpoint — local Ollama (`/v1`) or a cloud
  provider. Wrapped behind a small Go `Summarizer` interface so the provider is
  swappable. Connection details live in `sylloge.toml` — never exposed to the
  teacher.
- **Deploy**: small Docker container (GHCR image) or a single binary on a LAN server;
  the LLM endpoint is wherever `sylloge.toml` points.

## Deferred (future versions)

- Anonymize student names / PII before sending to the model.
- Reverse proxy for secure access from outside the home network (will require auth).
- LLM parameter sliders (temp, top_p, etc.) so the teacher can tune output to taste.
- Multi-student batch input / splitting a class dump into per-student notes.
- Model selection is not pinned; owner will swap models in/out for testing.

## Do not

- Do not add auth, external access, or multi-student batching in v1.
- Do not expose LLM settings in the UI.
- Do not introduce a JS framework or build pipeline; keep it HTMX + server-rendered HTML.

## Conventions for the build agent (local coder)

This project is built through a hybrid opencode workflow: the frontier `plan` agent
writes a file-level spec, the local `build` agent (Qwen3-Coder-30B-A3B) executes it, and
the frontier `review` agent audits the diff. If you are the `build` agent, follow these:

- **Layout**: `cmd/sylloge/main.go` for the entrypoint; all application code under
  `internal/` (e.g. `internal/store`, `internal/summarize`, `internal/server`,
  `internal/web` for templates). Nothing importable outside this module.
- **Stack discipline**: Go stdlib first. Server-rendered `html/template` + HTMX only.
  No SPA, no JS framework, no JS build pipeline. SQLite via a single store package.
- **AI provider**: keep the `Summarizer` interface small and provider-agnostic;
  the impl POSTs to the OpenAI-compatible `/v1/chat/completions`. Config comes
  from the TOML file — never hardcode host/model/key.
- **Scope guard**: obey the "Do not" list above (no auth, no external access, no
  multi-student batching, no LLM settings in the UI for v1). If a task implies any of
  these, stop and flag it rather than implementing it.
- **Verification**: after each task, ensure `go build ./...` and `go vet ./...` pass;
  add a table-driven test when you add a package with real logic.
- **Edits**: make minimal, surgical changes; do not reformat unrelated code. Prefer
  editing existing files over creating new ones.

## Next session — start here

**v1 is shipped.** Live at https://github.com/parvosub/sylloge — release
`v1.0.0` (binaries + checksums on Releases, image on GHCR). CI green.
Read `docs/session-2026-09-01.md` for the full ship-session summary,
verified/unverified list, and workflow notes (permissions, security hook).

**Current architecture**: `cmd/sylloge` (entrypoint, `--version`) +
`internal/{config,store,summarize,server,web}`. TOML config
(`sylloge.toml`, copy from `sylloge.toml.example`) is the single source of
truth for database path, LLM provider/model/system prompt, and API
base_url/api_key. `SYLLOGE_ADDR` sets the listen address; `SYLLOGE_CONFIG`
overrides the config path.

**Running locally (no Docker):**
```sh
cp sylloge.toml.example sylloge.toml   # edit to point llm.api.base_url at your Ollama host
go run ./cmd/sylloge
# open http://localhost:8080
```
Database: `sylloge.db` in project root (gitignored).

**Unverified (needs other hosts):**
- `docker compose up -d` on the Proxmox host (pulls the GHCR image).
- The curl installer from a fresh machine against a real release.
- GHCR image is linux/amd64 only (arm64 needs the binary or source build).

**When adding a new feature:** start with a `docs/specs/` spec from the
`plan` agent, then hand it to the local `build` agent, then `review`.
Obey the "Do not" list above — items there need their own v2 spec first.

**Future versions (deferred — do not build without a new spec):**
- Auth + reverse proxy for outside access.
- Multi-student batch input.
- LLM parameter sliders in UI.
- PII anonymization before sending to model.
- Multi-arch (linux/arm64) Docker image.
- Auto-generate default config on first run.
