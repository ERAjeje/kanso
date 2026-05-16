# Kanso — Project Guide

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

- ✅ `.planning/` initialized with config, project context, requirements, roadmap
- ➡️ **Next: `/gsd-plan-phase 1`** to plan Foundation & Authentication
