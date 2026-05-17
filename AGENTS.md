# Kanso — Project Guide

## 🔴 HARD RULES — NEVER VIOLATE (Zero Tolerance)

These rules are ABSOLUTE. Any violation is a workflow failure and must be reported.

### Rule A: No code execution without `/approve`
- **No file edits, no test writes, no implementation code** before explicit `/approve`
- "sim", "ok", "pode fazer" are NOT approval — only `/approve` counts
- Exception: reading files, asking questions, planning, discussing are allowed

### Rule B: Workflow sequence is MANDATORY
The sequence is **LOCKED**: discuss → plan → approve → execute (TDD inside execute)
- NEVER skip a step
- NEVER start implementation during discussion
- NEVER start execution during planning
- NEVER execute before `/approve`

### Rule C: Bugs follow `/gsd-debug` workflow
- Bug report → MUST route to `/gsd-debug`
- `/gsd-debug` follows its own: capture → diagnose → plan fix → **/approve** → fix + test
- NEVER jump directly to code changes for a bug report

### Rule D: Verify before acting — S.T.O.P.
Before writing ANY code, ask yourself:
1. **S** — Is there a plan? (PLAN.md exists?)
2. **T** — Is it approved? (APPROVALS.md shows approved?)
3. **O** — What step am I on? (discuss/plan/approve/execute?)
4. **P** — Proceed only if the answer to all is correct.

If ANY answer is no → **STOP AND REPORT IMMEDIATELY**.

---

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
5. **BUG WORKFLOW** — All bugs must pass through `/gsd-debug` before any code change

#### Approval Gate — Enforced Checks (do not skip)

When `/approve` is received, the orchestrator MUST:

a) **Verify pending plan exists** in `.planning/APPROVALS.md` under `## Pending Approval` section.
b) **Verify PLAN.md exists** at `.planning/phases/<plan-slug>/01-PLAN.md` matching the pending plan ID.
c) If either is missing → **BLOCK** execution and report:
   ```
   ⛔ APPROVAL REJECTED — Pre-condition not met.
      Missing: [APPROVALS.md pending entry / PLAN.md]
      Action required: Complete planning docs first.
   ```
d) Only after both exist: update APPROVALS.md → mark as approved → proceed to TDD gate → execute.

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
