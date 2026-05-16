# Kanso — Project Guide

## Orchestrator (Default Behavior)

⚠️ **This section defines your default operating mode. Follow it for every interaction.**

**At session start, load the orchestrator skill:**
```
skill name="gsd-orchestrator"
```

The orchestrator is your **default behavior** for EVERY user prompt. Before responding,
always classify the user's intent and route to the appropriate GSD command.

### Immutable Rules (Always Enforced)

1. **APPROVAL GATE** — No code execution without explicit `/approve` command
2. **WORKFLOW SEQUENCE** — Always: discuss → plan → approve → execute
3. **TDD** — Tests written before implementation code (Red → Green → Refactor)
4. **TRACEABILITY** — Every execution links to an approved plan

### Key Files
- `.planning/APPROVALS.md` — Tracks approval state (pending, approved, rejected)
- `.planning/STATE.md` — Project state, progress, pending todos
- `.planning/ROADMAP.md` — Phase definitions and status

### Routing Summary
| Intent | Route |
|--------|-------|
| "continuar", "próximo passo", "resumir" | `/gsd-resume-work` |
| "status", "progresso", "como estamos" | `/gsd-progress` |
| "implementar", "criar", "fazer", "resolver" | → discuss → plan → **/approve** → TDD → execute |
| "bug", "erro", "crash", "não funciona" | `/gsd-debug` |
| "testes", "testar", "cobertura" | `gsd-add-tests` |
| "planejar", "plano", "planejamento" | `/gsd-plan-phase` |
| "revisar", "review", "qualidade" | `/gsd-code-review` |
| "verificar", "validar" | `/gsd-verify-work` |
| "limpar contexto", "nova sessão", "reset", "nova funcionalidade" | `/new` + auto handoff |
| "ajuda", "o que posso fazer" | `/gsd-help` or routing table |
| **`/approve`** | Authorizes execution of pending plan |

**Reminder:** "sim", "ok", "pode fazer" ≠ `/approve`. Only `/approve` unlocks execution.

---

## GSD Workflow Commands

### Project Planning
- `/gsd-progress` — Check progress, advance workflow, or dispatch freeform intent
- `/gsd-discuss-phase N` — Gather context and clarify approach for a phase
- `/gsd-plan-phase N` — Create detailed plan for a phase
- `/gsd-execute-phase N` — Execute all plans in a phase

### Quality Gates
- `/gsd-code-review` — Review source files for bugs, security issues
- `/gsd-verify-work` — Validate built features through conversational UAT
- `/gsd-secure-phase` — Verify threat mitigations

### Project Management
- `/gsd-stats` — Display project statistics and timeline
- `/gsd-workstreams` — Manage parallel workstreams
- `/gsd-capture` — Capture ideas, tasks, notes
- `/gsd-settings` — Update workflow preferences

## Phase Structure

1. **Foundation & Authentication** — Infra (Docker, Traefik, CouchDB) + Google OAuth + JWT
2. **Core Diary: Registro & Sync** — Emotion form + PouchDB offline + CouchDB sync
3. **Reports** — PDF generation via chromedp + async job + download

## Stack

- **Frontend:** React 19 / Vite 8 / TypeScript / Tailwind / PouchDB
- **Backend:** Go 1.26 (chi, JWT, FCM, Twilio)
- **NLP:** Python 3.12 (FastAPI, transformers) — deferred to v2
- **Database:** CouchDB 3.5
- **Infra:** Traefik v3 / Docker / Docker Compose

## Current State

See [.planning/STATE.md](.planning/STATE.md) for up-to-date project status, progress, pending todos, and session continuity.
