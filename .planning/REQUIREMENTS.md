# Requirements: Kanso

**Defined:** 2026-05-16
**Core Value:** O usuário consegue nomear e registrar suas emoções no momento em que as sente, criando um histórico que torna o processo terapêutico mais concreto e orientado a dados.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Authentication

- [ ] **AUTH-01**: User can sign in with Google OAuth
- [ ] **AUTH-02**: User receives signed JWT after authentication for API access
- [ ] **AUTH-03**: User can log out, invalidating session

### Registro

- [ ] **REG-01**: User can register emotion with date/time (defaults to now), sensações (texto livre), sentimento (combobox customizável), contexto (texto livre), pensamentos (texto livre)
- [ ] **REG-02**: User can retroactively set date/time to a past moment
- [ ] **REG-03**: Sentimento combobox lists previously saved sentiments and allows typing new ones

### Sincronização

- [ ] **SYNC-01**: Registrations save immediately to PouchDB (offline)
- [ ] **SYNC-02**: PouchDB automatically syncs to CouchDB when connectivity is available
- [ ] **SYNC-03**: CouchDB is configured with per-user database isolation via proxy auth

### Relatórios

- [x] **REL-01**: User can generate a PDF report covering the period from the last report date to today
- [x] **REL-02**: Report includes all registered emotions in the period with their fields (data, sensações, sentimento, contexto, pensamentos)
- [x] **REL-03**: Report generation runs asynchronously (job-based), user can poll for completion

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Histórico

- **HIST-01**: User can view chronological list of past registrations
- **HIST-02**: User can filter/search registrations by date range, text, or sentiment

### Notificações

- **NOTF-01**: User receives push notifications at configurable times (default 12, 18, 23h)
- **NOTF-02**: User can enable/disable notifications and customize reminder times
- **NOTF-03**: Tapping notification opens the registration screen

### Contato Psicóloga

- **CONT-01**: User can save therapist name and WhatsApp number in profile
- **CONT-02**: User can update therapist contact information

### Análise NLP

- **NLP-01**: Backend analyzes registration text (sensações + contexto + pensamentos) for emotion patterns
- **NLP-02**: Detected emotions are stored alongside the registration
- **NLP-03**: Analysis runs asynchronously — does not block registration

### WhatsApp

- **WHAT-01**: User can send generated PDF report to therapist via WhatsApp
- **WHAT-02**: Report is sent to the saved therapist WhatsApp number via Twilio

### Perfil

- **PROF-01**: User can view their profile information (name, email from Google)
- **PROF-02**: Dark mode toggle
- **PROF-03**: Session persists across app opens (no re-login on every visit)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Social features / community | Emotional data is deeply personal; privacy risk |
| Gamification (streaks, badges) | Emotional health is not a game — guilt from broken streaks undermines purpose |
| AI chatbot therapy | Cannot replace a therapist; liability risk |
| Multiple therapists | Only one psychologist per user; simple "change" action later if needed |
| Real-time chat with therapist | Blurs journaling/therapy boundary; WhatsApp is file delivery only |
| App nativo (iOS/Android) | PWA suficiente para MVP e portfólio |
| Suporte offline sem PouchDB | Offline-first é core da arquitetura |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 — Foundation & Authentication | Pending |
| AUTH-02 | Phase 1 — Foundation & Authentication | Pending |
| AUTH-03 | Phase 1 — Foundation & Authentication | Pending |
| REG-01 | Phase 2 — Core Diary: Registro & Sync | Pending |
| REG-02 | Phase 2 — Core Diary: Registro & Sync | Pending |
| REG-03 | Phase 2 — Core Diary: Registro & Sync | Pending |
| SYNC-01 | Phase 2 — Core Diary: Registro & Sync | Pending |
| SYNC-02 | Phase 2 — Core Diary: Registro & Sync | Pending |
| SYNC-03 | Phase 1 — Foundation & Authentication | Pending |
| REL-01 | Phase 3 — Reports | Complete |
| REL-02 | Phase 3 — Reports | Complete |
| REL-03 | Phase 3 — Reports | Complete |

**Coverage:**
- v1 requirements: 12 total
- Mapped to phases: 12 ✓
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-16*
*Last updated: 2026-05-16 after initial definition*
