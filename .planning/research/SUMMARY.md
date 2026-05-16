# Project Research Summary

**Project:** Kanso
**Domain:** Therapeutic emotion diary PWA (Portuguese, offline-first, NLP-enhanced)
**Researched:** 2026-05-16
**Confidence:** HIGH (stack/pitfalls verified via official docs; features MEDIUM based on competitive analysis)

## Executive Summary

Kanso is an offline-first therapeutic emotion diary PWA built around a unique therapist-as-primary-consumer model. Unlike mood trackers that are self-help tools with optional sharing (Daylio, Bearable, Journey), Kanso positions the therapist as the central consumer of structured emotional data delivered via automated WhatsApp PDF reports. The core technical challenge is maintaining reliable offline data collection while orchestrating server-side NLP classification, PDF generation, and WhatsApp delivery — all as a single-binary Go backend with Python inference.

The recommended architecture is a **PouchDB→CouchDB direct sync model** with a stateless Go API layer for authentication and side-effects only. This is the established pattern for offline-first browser apps: PouchDB in the IndexedDB-backed browser syncs bidirectionally with CouchDB via live replication (`{live: true, retry: true}`), while the Go backend (chi router) handles Google OAuth validation, JWT issuance, NLP orchestration, chromedp PDF rendering, FCM push dispatch, and Twilio WhatsApp integration. The Python NLP service runs on an internal Docker network — never publicly exposed — and processes emotion classification on a multilingual XLM-RoBERTa model. All services are fronted by Traefik v3 with automatic Let's Encrypt TLS.

**Key risks and mitigations:** (1) Silent data loss from unhandled PouchDB conflicts — solved by immutable append-only document design with UUID `_id`s, avoiding concurrent writes entirely. (2) NLP as a hard dependency that blocks user flow — solved by making analysis fully async: entries save immediately to local PouchDB, NLP enriches the CouchDB document later. (3) chromedp memory exhaustion in Docker — solved by single-concurrent PDF generation via Go channel semaphore, pinned `chromedp/headless-shell` image, and `--shm-size=2g`. (4) Portuguese emotion model accuracy risk — `tabularisai/multilingual-emotion-classification` (XLM-RoBERTa) is the recommended model; must be validated against Portuguese emotional phrases before production.

## Key Findings

### Recommended Stack

The stack follows the well-established **PouchDB ↔ CouchDB offline-first PWA pattern** with a Go service layer for auth and side-effects. This is not novel architecture — it's the standard approach for browser-based offline-first apps, used by hundreds of projects documented in PouchDB/CouchDB guides. The novelty is in the feature composition (structured diary + NLP + WhatsApp PDF).

**Core technologies:**
- **React 19.2 + Vite 8 + TypeScript 5.8+** — PWA frontend with fast HMR, compiler-optimized renders, and Rolldown-based bundling. Vite SPA chosen over Next.js because SSR adds complexity with no benefit for offline-first PWA.
- **Go 1.26 + chi v5** — Backend API with single-binary deployment, excellent concurrency for parallel JWT validation/PDF generation/FCM dispatch. Chi chosen over Express/Fastify for its stdlib-native router and minimal dependency tree.
- **CouchDB 3.5 + PouchDB 9** — Document database and client DB with native bidirectional replication protocol. The only viable stack for offline-first browser sync — no alternative exists for PouchDB-compatible sync outside CouchDB.
- **FastAPI + transformers + PyTorch (CPU)** — NLP service with HuggingFace pipeline API. Python is the only viable ecosystem for transformer inference in 2026; Go has no equivalent.
- **Traefik v3** — Reverse proxy with automatic Docker service discovery, Let's Encrypt TLS, and middleware pipeline. Chosen over Caddy for multi-service Docker Compose routing.
- **Emotion model:** `tabularisai/multilingual-emotion-classification` (XLM-RoBERTa, 11 emotions, MIT license, ~1.1GB)

**Supporting services:** chromedp/headless-shell for PDF generation, Firebase Admin SDK (Go) for FCM push, Twilio Go SDK for WhatsApp, Zustand for lightweight client state.

**What NOT to use:** Redux (overkill for 3 screens), Next.js (breaks offline assumptions), MongoDB/Firestore (no browser sync protocol), Redis (no pub/sub needs), celery (FastAPI BackgroundTasks sufficient), GPU PyTorch (unnecessary for single-user inference).

### Expected Features

**Must have (P1 — table stakes):**
- Quick mood entry with structured 4-field form (sensação, sentimento, contexto, pensamento) — core action
- Offline-first storage with PouchDB auto-sync to CouchDB — non-negotiable for mobile use
- Chronological history view (list + calendar) — users need to see past entries
- Google OAuth authentication — required for sync and scale
- Retroactive entry (datetime picker for backdating) — low cost, high practical value
- Tab navigation (Register, History, Profile)

**Should have (P2 — differentiators):**
- Push notification reminders (12/18/23h defaults, configurable) — users forget without them
- Charts / mood trends (line chart + distribution) — key for therapeutic pattern recognition
- Dark mode — expected in 2026
- Custom sentiment labeling — discovered in therapy, therapeutic value
- Search / filter by text, date range, sentiment

**Defer (P3 — v2+):**
- NLP emotion analysis in Portuguese — technically risky (model accuracy for Portuguese), expensive (1-2GB model), and the app works without it
- PDF report generation — requires NLP to be truly valuable (otherwise data dump)
- WhatsApp therapist delivery — deep dependency chain: entries → NLP → PDF → Twilio
- Biometric/PIN lock — PWA limitation; WebAuthn still evolving on mobile

**Anti-features (deliberately avoid):** Social features/community (privacy risk, moderation burden), gamification/streaks (creates guilt on bad days), AI chatbot therapy (liability, cannot replace therapist), multi-therapist (bloat for MVP), real-time chat (blurs therapy boundary).

### Architecture Approach

Kanso uses a **direct PouchDB↔CouchDB sync architecture** where the Go backend handles only authentication and side-effects — never CRUD. The browser saves entries to local PouchDB immediately (works offline), and live replication syncs bidirectionally with CouchDB through Traefik. JWT validation happens at the CouchDB level natively (CouchDB 3.x `jwt_authentication_handler`), eliminating the need for proxy header injection. Go orchestrates: Google OAuth → JWT issuance, NLP analysis (reads from CouchDB → calls internal Python service → writes results back), PDF generation (chromedp headless Chrome), WhatsApp delivery (Twilio), and push notifications (FCM). The Python NLP service is isolated on an internal Docker network with no public route.

**Major components:**
1. **Frontend PWA (React + Vite + PouchDB + Tailwind)** — UI rendering, local-first data with IndexedDB-backed PouchDB, JWT management, offline queue
2. **CouchDB** — Document storage for registros/sentimentos/usuarios/relatorios databases, replication target for PouchDB, native JWT auth, Mango indexes, `validate_doc_update` for data integrity
3. **Go Backend (chi)** — Google OAuth validation, JWT signing, PDF generation (chromedp), WhatsApp sending (Twilio), FCM push dispatch, NLP orchestration (admin CouchDB access)
4. **Python NLP Service (FastAPI)** — Emotion analysis via HuggingFace transformers pipeline, POST /analyze endpoint on internal network only
5. **Traefik v3** — TLS termination, path-based routing (/api/* → Go, /db/* → CouchDB), CORS, rate limiting

### Critical Pitfalls

1. **Silent data loss from PouchDB conflicts** — Entries lost when concurrent writes produce unresolved revision trees. **Prevention:** Append-only immutable entries with UUID `_id`s (no concurrent updates possible). Listen for `_conflicts` in changes feed. Register `navigator.storage.persist()` at startup to prevent IndexedDB eviction.

2. **CouchDB exposed without auth (emotional data leak)** — PouchDB sync URL requires CouchDB on the public internet. **Prevention:** Always proxy through Traefik; never expose CouchDB directly. Configure `require_valid_user = true`, per-database `_security` docs, and `validate_doc_update` functions enforcing `userId === userCtx.name`.

3. **Live sync dies silently on network loss** — Without `{retry: true}`, PouchDB replication stops on first network error with no recovery. **Prevention:** Always use `{live: true, retry: true}`. Implement sync status UI indicators. Listen for `'paused'`/`'active'`/`'error'` events. Cancel and restart replication on auth token expiry.

4. **NLP as hard dependency blocks user flow** — Synchronous NLP analysis on entry creation causes timeouts when NLP service is slow/down. **Prevention:** Make NLP fire-and-forget async. Save entries to CouchDB immediately. Analyze in background via changes feed or scheduled job. Always set `http.Client{Timeout: 10s}`. Implement circuit breaker (3 failures → 30s cooldown). Cache results in `analiseEmocoes` field preemptively.

5. **chromedp OOM in Docker** — Concurrent headless Chrome instances exhaust container memory. **Prevention:** Use `chromedp/headless-shell` Docker image (~140MB). Set `--shm-size=2g`. Implement channel-based semaphore for max 1 concurrent PDF generation. Reuse allocator context. Set 30s context timeout.

## Implications for Roadmap

Based on dependency chains from FEATURES.md and build order from ARCHITECTURE.md:

### Phase 1: Foundation & Authentication
**Rationale:** Everything depends on infrastructure and auth. CouchDB, Traefik, and Google OAuth/JWT are prerequisites for offline sync. Docker Compose with multi-file setup prevents config drift from day one.
**Delivers:** Running Docker Compose stack (Traefik + CouchDB + Go skeleton + React scaffold), Google OAuth login flow, JWT issuance, CouchDB JWT-native authentication, empty chi handler stubs.
**Addresses:** P1 feature: Google OAuth authentication.
**Avoids:** Pitfall 2 (CouchDB auth leak — proxy through Traefik from start), Pitfall 9 (Docker Compose config drift — multi-file setup from day one).
**Flags:** Standard patterns. CouchDB JWT auth, Google OAuth, Traefik routing are well-documented. Skip research-phase.

### Phase 2: Core Diary — Entry Crud & Offline Sync
**Rationale:** Quick mood entry and offline-first storage are the heart of the app. Without this, there is no product. PouchDB CRUD, live sync, and history view form the user-visible output. This phase delivers the vertical slice that makes Kanso work.
**Delivers:** Register tab with structured 4-field form + retroactive datetime picker, History tab with chronological list, PouchDB init with `{live: true, retry: true}` sync, CouchDB Mango indexes + `validate_doc_update`, sync status indicator, `navigator.storage.persist()` at startup.
**Addresses:** P1 features: quick mood entry, offline-first storage, history view, retroactive entry, tab navigation.
**Avoids:** Pitfall 1 (conflict data loss — append-only UUID docs + conflict listener), Pitfall 3 (sync dies silently — retry:true + sync UI), Pitfall 13 (IndexedDB eviction — persist() at startup).
**Flags:** Standard patterns. PouchDB/CouchDB sync is well-documented. Skip research-phase.

### Phase 3: Enhanced User Features
**Rationale:** Push reminders, charts, search, dark mode, and custom sentiment are independent of NLP and PDF. They can be built immediately after the core diary is working, providing user engagement and retention before the more complex NLP pipeline.
**Delivers:** Configurable push notification reminders via FCM (12/18/23h defaults, snooze, smart suppression), mood trends line chart + distribution view, text/date range/sentiment search, dark mode toggle, editable sentiment combo box, pause notifications toggle.
**Addresses:** P2 features: push reminders, charts, dark mode, custom sentiment, search/filter.
**Avoids:** Pitfall 5 (default Chrome notifications — proper promise chain in push handler), Pitfall 10 (notification fatigue — snooze + pause + smart suppression), Pitfall 12 (notification click fails — proper notificationclick handler opening Register tab).
**Flags:** Push notification service worker patterns need careful implementation but are well-documented. Chart library selection (e.g., Chart.js vs Recharts vs SVG) needs a brief research check during planning.

### Phase 4: NLP Emotion Analysis (Async Pipeline)
**Rationale:** NLP enriches all downstream features (charts, reports) but the app works without it. Delaying it defers the highest-risk technology choice (Portuguese emotion model accuracy) until the core product is validated. The async architecture prevents NLP from blocking user flow.
**Delivers:** Python FastAPI NLP service with XLM-RoBERTa model on internal Docker network, Go NLP client with circuit breaker (`sony/gobreaker`) and explicit timeout, background analysis via CouchDB changes feed or scheduled job, NLP results stored in `analiseEmocoes` field, CouchDB rich queries by emotion probability.
**Addresses:** P3 feature: NLP emotion analysis in Portuguese. Enriches: charts (emotion breakdown), reports (enriched data).
**Avoids:** Pitfall 4 (NLP as hard dependency — async analysis, never blocks save), Pitfall 7 (NLP hard dependency for PDF — graceful degradation), Pitfall 11 (Go assumes NLP always available — circuit breaker + timeout), Pitfall 14 (Portuguese model fails — test against 20+ Portuguese emotional phrases before deployment).
**Flags:** NEEDS RESEARCH — Portuguese emotion model accuracy must be validated during planning. The recommended model (`tabularisai/multilingual-emotion-classification`) needs eval against Brazilian Portuguese emotional text. If accuracy <60%, consider fine-tuning or fallback models.

### Phase 5: Reports & WhatsApp Delivery
**Rationale:** The killer feature — therapist PDF report via WhatsApp — has the deepest dependency chain in the app. It requires: entries (Phase 2) + NLP enrichment (Phase 4) + PDF generation + Twilio integration. Building it last ensures all dependencies are in place and validated.
**Delivers:** Async PDF report generation via chromedp/headless-shell with HTML templating + SVG bar charts, single-concurrent Chrome semaphore, job polling pattern (202 Accepted → poll status), authenticated PDF download with short-expiry signed URLs, Twilio WhatsApp message with PDF attachment, delivery confirmation via Twilio webhook, therapist phone number management (encrypted in CouchDB).
**Addresses:** P3 features: PDF report generation, WhatsApp therapist delivery. The core Kanso differentiator.
**Avoids:** Pitfall 4 (Chrome OOM — semaphore + shm-size + headless-shell image), Pitfall 8 (WhatsApp data leak — authenticated PDF URLs, stripped metadata, encrypted therapist phone, delivery confirmation), Pitfall 9 (Docker config drift — compose overrides).
**Flags:** NEEDS RESEARCH — chromedp PDF rendering quality needs testing with Portuguese text + Unicode. Twilio WhatsApp media attachment format needs API verification. Brief research during planning.

### Phase 6: Polish & Production Hardening
**Rationale:** Everything else depends on the product being deployed and reliable. Backup strategy, monitoring, error tracking, and performance optimization come after the full feature set exists.
**Delivers:** CI/CD pipeline (GitHub Actions), health checks on all Docker services, CouchDB backup replication to remote, error tracking integration, performance optimization (lazy loading, code splitting, CouchDB view tuning), rate limiting fine-tuning.
**Addresses:** Production readiness for all features.
**Avoids:** Pitfall 9 (Docker config drift — verified), general production failure modes.
**Flags:** Standard patterns. Skip research-phase.

### Phase Ordering Rationale

- **Dependency-driven:** P1 (infra + auth) → P2 (core diary) → P3 (enhancements) → P4 (NLP) → P5 (reports) follows the strict dependency chain: entries need auth and storage, reports need entries and NLP, WhatsApp needs reports.
- **Risk deferral:** NLP (highest technical risk — Portuguese model accuracy) is deferred to Phase 4, after the core product is validated. This prevents model uncertainty from blocking the entire product.
- **Value acceleration:** P3 (push, charts, dark mode) is inserted between core diary and NLP to deliver user engagement features quickly while the NLP pipeline is being built.
- **Architecture alignment:** The async job pattern for PDF generation (P5) mirrors the async NLP pattern (P4), creating a consistent background processing architecture. Both use the CouchDB changes feed for job dispatch.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 4 (NLP):** Portuguese emotion model accuracy needs empirical validation. Must test `tabularisai/multilingual-emotion-classification` against a curated set of 20-30 Brazilian Portuguese emotional phrases before committing to the model. If accuracy is insufficient, fallback options include fine-tuning or English-only analysis with Portuguese interface.
- **Phase 5 (Reports):** chromedp PDF rendering quality with Portuguese text (Unicode, special characters) needs testing. Twilio WhatsApp Media API for PDF attachment delivery needs API verification — specifically the temporary URL format and size limits.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Foundation):** Google OAuth, CouchDB JWT auth, Traefik routing — all well-documented patterns with official guides
- **Phase 2 (Core Diary):** PouchDB sync with CouchDB — well-documented with multiple production examples
- **Phase 3 (Enhanced Features):** FCM push notifications — well-documented by Google, though service worker patterns need careful testing
- **Phase 6 (Polish):** Standard CI/CD, monitoring, backup patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified via npm/GitHub/PyPI/Docker Hub. Go+chi+CouchDB+PouchDB is a well-proven stack. FastAPI+transformers is standard for ML inference. |
| Features | MEDIUM | Based on competitive analysis of Daylio/Journey/Bearable/Youper official sites and PROJECT.md requirements. No primary user research conducted. Anti-features are opinionated but reasonable. |
| Architecture | HIGH | PouchDB↔CouchDB direct sync, Go chi layered architecture, Traefik gateway, NLP isolation — all follow established patterns documented in official sources. Anti-patterns identified have clear prevention. |
| Pitfalls | HIGH | Each pitfall documented with cause, prevention, and phase to address. Sources include official PouchDB/CouchDB docs, chromedp GitHub, web.dev, MDN, Docker Compose guide. UX pitfalls synthesized from multiple HCI/therapy app sources. |

**Overall confidence:** HIGH

### Gaps to Address

- **Portuguese NLP model accuracy:** The most significant unknown. The research recommends `tabularisai/multilingual-emotion-classification` as the best available option, but real-world accuracy on Brazilian Portuguese emotional text has not been empirically tested. This must be validated in Phase 4 planning with a benchmark evaluation.
- **No primary user research:** Feature prioritization is based on competitor analysis and project requirements, not validated user testing. The therapy-specific assumptions (structured 4-field form, periodic reports via WhatsApp) should ideally be validated with actual therapists/users before Phase 5.
- **Push notification mobile behavior:** PWA push notification behavior varies significantly across browsers (Chrome on Android, Safari on iOS, Firefox). The research documents patterns for Chrome but may need adjustments for iOS support when the PWA is added to the home screen.
- **CouchDB at scale:** The architecture assumes shared databases with `userId` document ownership. This works for single-user MVP but may need per-user databases at scale. The transition path is noted but not deeply researched.
- **chromedp font rendering:** PDF generation with Portuguese text including accented characters and special Unicode (emoticons) in the emotion combobox needs visual validation. chromedp may need system font packages installed in the Docker image.

## Sources

### Primary (HIGH confidence)
- **CouchDB 3.x JWT Authentication** — docs.couchdb.org (official docs)
- **PouchDB Replication Guide** — pouchdb.com/guides/replication.html (official docs)
- **PouchDB Conflicts Guide** — pouchdb.com/guides/conflicts.html (official docs)
- **Chi v5 REST Example** — github.com/go-chi/chi/_examples/rest/main.go (official example)
- **Chromedp README** — github.com/chromedp/chromedp (official repo, v0.15.1)
- **Chromedp/headless-shell Docker** — hub.docker.com/r/chromedp/headless-shell (official image)
- **Traefik v3 Middleware Docs** — doc.traefik.io/traefik (official docs)
- **Web.dev Push Notifications** — web.dev/articles/push-notifications-handling-messages (Google Chrome team)
- **Docker Compose Production Guide** — docs.docker.com/compose/production/ (official docs)
- **navigator.storage.persist()** — MDN Web Docs (official reference)
- **React 19.2, Vite 8, Tailwind 4.3, PouchDB 9.0** — npm/PyPI/Docker Hub version verification
- **Go 1.26** — go.dev/VERSION (official release)
- **tabularisai/multilingual-emotion-classification** — HuggingFace Hub (model config verified)

### Secondary (MEDIUM confidence)
- **Daylio, Journey, Bearable, Youper** — official marketing sites (official but biased toward positive presentation)
- **CouchDB Security** — official CouchDB docs (intro/security.html)
- **Traefik ForwardAuth Middleware** — Traefik v3 routing configuration docs
- **Emotion tracking app UX pitfalls** — synthesized from multiple HCI/therapy app UX research sources

### Tertiary (LOW confidence)
- **User behavior patterns for mood tracking** — inferred from competitor feature sets, not from primary research
- **Portuguese emotion model accuracy** — no benchmark data for the recommended model on Portuguese text exists in reviewed sources

---
*Research completed: 2026-05-16*
*Ready for roadmap: yes*
