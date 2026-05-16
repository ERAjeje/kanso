---
phase: 2
slug: core-diary-registro-sync
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-16
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Vitest (aligned with Vite 8 ecosystem) |
| **Config file** | none — Wave 0 installs `vitest.config.ts` |
| **Quick run command** | `npx vitest run --reporter=verbose --changed` |
| **Full suite command** | `npx vitest run --reporter=verbose` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `npx vitest run --reporter=verbose --changed`
- **After every plan wave:** Run `npx vitest run --reporter=verbose`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 1 | REG-01 | — | N/A | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "renders all fields"` | ❌ W0 | ⬜ pending |
| 2-01-02 | 01 | 1 | REG-01 | — | N/A | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "calls saveRegistro on submit"` | ❌ W0 | ⬜ pending |
| 2-01-03 | 01 | 1 | REG-02 | — | N/A | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "datetime-local has no min/max"` | ❌ W0 | ⬜ pending |
| 2-01-04 | 01 | 1 | REG-01 | — | N/A | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "form resets after save"` | ❌ W0 | ⬜ pending |
| 2-02-01 | 02 | 1 | REG-03 | — | N/A | unit | `npx vitest run src/components/SentimentoCombobox.test.tsx -t "shows all sentiments"` | ❌ W0 | ⬜ pending |
| 2-02-02 | 02 | 1 | REG-03 | — | N/A | unit | `npx vitest run src/components/SentimentoCombobox.test.tsx -t "filters list on type"` | ❌ W0 | ⬜ pending |
| 2-02-03 | 02 | 1 | REG-03 | — | N/A | unit | `npx vitest run src/components/SentimentoCombobox.test.tsx -t "creates sentiment on blur"` | ❌ W0 | ⬜ pending |
| 2-03-01 | 03 | 2 | SYNC-01 | — | N/A | integration | `npx vitest run src/services/registros.test.ts -t "saves to PouchDB"` | ❌ W0 | ⬜ pending |
| 2-03-02 | 03 | 2 | SYNC-02 | — | N/A | unit | `npx vitest run src/services/pouchdb.test.ts -t "creates live sync with retry"` | ❌ W0 | ⬜ pending |
| 2-03-03 | 03 | 2 | D-15/D-16 | — | N/A | unit | `npx vitest run src/components/TabBar.test.tsx -t "renders 3 tabs"` | ❌ W0 | ⬜ pending |
| 2-03-04 | 03 | 2 | D-17 | — | N/A | unit | `npx vitest run src/components/TabBar.test.tsx -t "placeholder shows Em breve"` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `vitest` + `@testing-library/react` — install with `npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom`
- [ ] `src/components/RegistrationForm.test.tsx` — form rendering + submit + reset behavior
- [ ] `src/components/SentimentoCombobox.test.tsx` — filtering, create on blur, loading sentiments
- [ ] `src/components/TabBar.test.tsx` — 3 tabs, active state, placeholder content
- [ ] `src/services/registros.test.ts` — PouchDB put and query
- [ ] `src/services/pouchdb.test.ts` — sync configuration verification
- [ ] `src/hooks/usePouchSync.test.ts` — hook subscription and cleanup
- [ ] `src/hooks/__mocks__/pouchdb.ts` — mock PouchDB for tests (or use `pouchdb-memory` adapter)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Offline toast appears when submitting without connectivity | D-13 | Requires browser network condition simulation | Open DevTools → Network tab → Offline. Submit form. Verify toast "Salvo localmente — será sincronizado quando você estiver online" appears. |
| Sync status transitions from offline → syncing → online | D-14 | Requires live PouchDB sync with CouchDB | Start app offline. Verify red "Offline" dot. Go online. Verify yellow "Sincronizando..." then green "Sincronizado". |
| Sentimento auto-save on blur creates document in CouchDB | D-02 | Requires full PouchDB↔CouchDB sync | Submit form with new sentiment. Check CouchDB `sentimentos` database via Fauxton for the new doc. |

---

## Validation Sign-Off

- [ ] All tasks have automated verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
