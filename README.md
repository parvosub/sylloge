# Sylloge

> A grade and comment website for teachers. Turn free-form classroom notes into
> polished report-card comments with a local or cloud AI model.

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/parvosub/sylloge/actions/workflows/ci.yml/badge.svg)](https://github.com/parvosub/sylloge/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Overview

Sylloge is a small, self-hosted web app for a single teacher. The teacher types
free-form, unstructured observations about a student's work over days or months.
At the end of the term, one click asks an AI model to transform those notes into
a coherent, professional report-card comment. The teacher edits, copies, and
pastes the result into the school's website.

Screenshot placeholder: `docs/screenshot.png` (coming soon).

## Why this stack

| Choice | Rationale |
| --- | --- |
| **Go** | Single static binary, trivial to deploy, great for small self-hosted tools. |
| **HTMX + server-rendered HTML** | No JS framework, no build pipeline. The dynamic submit→summary flow stays lightweight and fast. |
| **SQLite** | Zero-config, single-file persistence — ideal for a one-teacher app backed up to a NAS. |
| **Ollama / OpenAI-compatible API** | One standard API works with a local Ollama server *or* any cloud provider (OpenAI, Together, Groq, etc.). |

## Features

- **Dashboard** — card grid of classes with an inline "Add Class" form.
- **Class & student management** — add classes, add students per class.
- **Notes** — append free-form observations over time; the three most recent stay
  visible, older ones collapse into a toggle.
- **AI summaries** — one-click generation from the notes, editable in place.
- **Copy / download** — copy formatted text or download as a file.
- **Summary history** — every generated version is kept for review.
- **Local or cloud LLM** — configure a local Ollama server or a cloud API key.

## Install

### Option 1 — Docker

Pull the prebuilt image from GitHub Container Registry (no source needed):

```sh
cp sylloge.toml.example sylloge.toml   # edit for your LLM
docker compose up -d
# open http://localhost:8080
```

The included `docker-compose.yml` pulls `ghcr.io/parvosub/sylloge:latest`. To
build from source instead, edit it and replace `image:` with `build: .`.

The database is persisted in a named volume (`sylloge-data`) and your config is
mounted read-only.

### Option 2 — curl installer (prebuilt binary)

```sh
curl -fsSL https://raw.githubusercontent.com/parvosub/sylloge/main/install.sh | bash
```

This downloads the latest binary for your OS/architecture from GitHub Releases,
verifies its checksum, installs it to `/usr/local/bin`, and writes a default
config to `~/.sylloge/sylloge.toml`.

To build from source instead:

```sh
go build -o sylloge ./cmd/sylloge
SYLLOGE_CONFIG=~/sylloge.toml ./sylloge
```

## Configuration

Configuration is a single TOML file (default `sylloge.toml`, override with the
`SYLLOGE_CONFIG` environment variable). Copy `sylloge.toml.example` and edit it.

```toml
[database]
path = "sylloge.db"                 # SQLite database file

[llm]
provider = "local"                  # "local" (Ollama) or "cloud"
model = "qwen3:8b"                  # model to use
system_prompt = "You are a supportive teaching assistant writing a report-card comment."

[api]
base_url = "http://localhost:11434/v1"   # OpenAI-compatible endpoint
api_key = ""                             # empty for local; required for cloud
```

### Local (Ollama)

1. Install [Ollama](https://ollama.com) on the machine that will serve the model.
2. Pull a model, e.g. `ollama pull qwen3:8b`.
3. Point `base_url` at `http://<ollama-host>:11434/v1` and set the model name.
4. Leave `api_key` empty.

### Cloud

1. Set `provider = "cloud"`.
2. Set `base_url` to your provider's OpenAI-compatible endpoint.
3. Set `api_key` to your key. The app sends it as a `Bearer` token.

## Running

```sh
./sylloge            # reads ./sylloge.toml by default
SYLLOGE_CONFIG=/path/to/sylloge.toml ./sylloge

./sylloge --version  # print version and exit
```

## Project structure

```
cmd/sylloge          Entry point: config load, flags, wiring
internal/config      TOML configuration loading and validation
internal/store       SQLite data layer (classes, students, notes, summaries)
internal/summarize   AI summarizer — OpenAI-compatible implementation
internal/server      HTTP handlers, routing, markdown→HTML
internal/web         HTML templates (embedded) and helpers
```

## Testing

```sh
go test ./...
```

The `store`, `server`, `summarize`, and `config` packages have table-driven
tests. CI runs `go build`, `go vet`, and `go test` on every push.

## Releasing

Tag a version and CI builds binaries for Linux and macOS (amd64 + arm64),
generates checksums, and publishes a GitHub Release:

```sh
git tag v1.0.0
git push origin v1.0.0
```

## License

[MIT](LICENSE)
