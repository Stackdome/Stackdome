# Context Map

This repo has multiple bounded contexts. Each has its own `CONTEXT.md` glossary.

| Context | Glossary | Scope |
|---|---|---|
| backend | `./CONTEXT.md` | API-server hub: `cmd/`, `pkg/` |
| frontend | `./frontend/CONTEXT.md` | React SPA: `frontend/src/` |
| agent | _deferred_ | spoke cluster-agent operator — no source in this repo; add lazily when agent terms are resolved |

System-wide ADRs: `docs/adr/`. Frontend-specific ADRs: `frontend/docs/adr/`.
