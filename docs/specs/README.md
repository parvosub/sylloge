# Design Specs

This directory records the design decisions and step-by-step specs that guided
the development of Sylloge. Each file is scoped to a discrete piece of work and
states its in-scope / out-of-scope boundaries and acceptance criteria.

## Contents

- `step2-data-layer.md` — SQLite schema and the `internal/store` CRUD layer.
- `step3-summarizer.md` — the `Summarizer` interface and Ollama provider.
- `step4-server-templates.md` — HTTP server, routes, and server-rendered templates.
- `step5-core-ui.md` — the core class → student → notes → summary flow.
- `phase1-fix.md` — hardening pass (env-based config, no hardcoded defaults).
- `web-enhancements-phase1.md` — dashboard, DORFic theme, summary history,
  editable summaries.
- `remove-timestamps-from-cards.md` — tightening the UI.

These are kept as a historical record of how the project was planned and built.
