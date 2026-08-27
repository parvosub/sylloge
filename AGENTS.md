# AGENTS.md

## Project

Greenfield project (repo is currently empty). Product: a grade and comment website for teachers.

Core flow:
1. Teacher enters free-form, unstructured notes about a student's work through the web front end.
2. An AI model transforms those notes into a coherent summary of the student's work.

## Status: design committed (v1)

Design session complete. Stack and scope below are committed for v1.

## Users & context

- **Single user**: one art teacher, not tech savvy. UI must be dead simple or she won't use it.
- **Scale**: 1 teacher, 4 classes (entered by teacher), multiple students per class (entered by teacher).
- **Environment**: deployed on a local server in a homelab, accessed via a web browser on the LAN.
- **Workflow**: teacher adds notes over time (days/months); at end of semester/year she copies the
  generated summary out and pastes it into her school's website.
- **Owner**: cloud systems admin (Operations), learning C, comfortable with Docker/Terraform/Ansible/IaC,
  reads Python but does not write it. Project doubles as a GitHub portfolio piece.

## v1 scope

- Flow: teacher types/pastes free-form notes for **one student** → AI returns a coherent summary →
  summary lands in an **editable** text box → teacher copies it out.
- **Per-student**: one student, one summary at a time.
- **Persistence**: classes, students, notes, and summaries are saved in SQLite and persist across
  days/months so she can revisit and edit.
- **No auth** for v1 (single user on trusted LAN).
- Teacher never touches LLM settings.

## Committed stack

- **Backend**: Go. Single static binary, trivial Docker deploy, career-aligned for the owner, adjacent
  to C. Renders HTML server-side via `html/template`.
- **Frontend**: server-rendered HTML + **HTMX** for the dynamic submit→summary interaction. No SPA,
  no JS build pipeline.
- **Storage**: **SQLite** (single file, backed up to NAS). Schema: `classes` → `students` →
  `notes` / `summaries`.
- **AI**: local **Ollama** on the dedicated Arch LLM box (i7 10700k, 64GB RAM, RTX 3090 24GB), reached
  over the LAN via its HTTP API (`/api/generate`). Wrapped behind a small Go `Summarizer` interface so
  the provider is swappable. Model + temperature/top_p + system prompt live in a **Modelfile** on the
  Ollama side — never exposed to the teacher.
- **Deploy**: app in a small Docker container on Proxmox; Ollama stays on the dedicated GPU box.

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
- **AI provider**: keep the `Summarizer` interface small and provider-agnostic; the
  Ollama impl POSTs to `/api/generate`. Config via env vars — never hardcode host/model.
- **Scope guard**: obey the "Do not" list above (no auth, no external access, no
  multi-student batching, no LLM settings in the UI for v1). If a task implies any of
  these, stop and flag it rather than implementing it.
- **Verification**: after each task, ensure `go build ./...` and `go vet ./...` pass;
  add a table-driven test when you add a package with real logic.
- **Edits**: make minimal, surgical changes; do not reformat unrelated code. Prefer
  editing existing files over creating new ones.

## Next session — start here

All 7 roadmap steps implemented. v1 is end-to-end verified locally with real
Ollama (`qwen3.5:27b-q4_K_M`). `go build ./...`, `go vet ./...`, `go test ./...`
all pass (store, summarize, server packages have real table-driven tests).

**Git**: initialized, initial commit on `main`. No remote configured yet.

**Build roadmap (v1) status:**
1. Project skeleton — DONE
2. Data layer — DONE (SQLite schema + `internal/store` CRUD)
3. Summarizer interface — DONE (`internal/summarize`, Ollama impl, env config)
4. HTTP server + templates — DONE (routes, per-page `html/template` sets, HTMX)
5. Core UI flow — DONE (full flow: class → student → notes → summary → save)
6. Packaging — DONE on disk: Dockerfile + docker-compose + README env table.
   **Not yet verified**: no Docker on dev machine; `docker compose up -d --build`
   on the Proxmox host.
7. Modelfile — DONE on disk (`Modelfile`, temperature 0.9 / top_p 0.95).
   **Not yet verified**: `ollama create sylloge -f Modelfile` on the Ollama box.

**Running locally (no Docker):**
```sh
cp sylloge.toml.example sylloge.toml   # edit to point llm.api.base_url at your Ollama host
go run ./cmd/sylloge
# open http://localhost:8080
```
Database: `sylloge.db` in project root (gitignored).

**What the customer is reviewing now:**
- The app is running at `localhost:8080` for browser testing.
- Customer feedback is pending — wait for their input before next steps.

**Unverified (needs Proxmox host):**
- `docker compose up -d --build` — verify Docker image builds and runs.
- `ollama create sylloge -f Modelfile` on the Ollama box.

**Future versions (deferred — do not build in v1):**
- Auth + reverse proxy for outside access.
- Multi-student batch input.
- LLM parameter sliders in UI.
- PII anonymization before sending to model.
