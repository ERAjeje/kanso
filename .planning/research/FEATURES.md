# Feature Research: Therapeutic Emotion Diary

**Domain:** Therapeutic emotion diary / mood tracker PWA (Portuguese, offline-first)
**Researched:** 2026-05-16
**Confidence:** MEDIUM — based on official app pages and comparison analysis; no primary user research conducted

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Quick mood/emotion entry** | Core action — Daylio's 2-tap entry is the gold standard. Users expect logging to take <30s. | LOW | Kanso's combobox + free text approach is more involved than Daylio but matches Bearable's model. Must be fast. |
| **Cronological history view** | Every app (Daylio, Journey, Bearable) has a timeline/feed of past entries. Fundamental for reflection. | LOW | List view + calendar view (togglable) expected. |
| **Reminders / notifications** | Daylio, Journey, Bearable all have configurable reminders. Users forget to log without them. | MEDIUM | Kanso has this (12, 18, 23h defaults). Need to support multiple reminders per day. |
| **Search / filter past entries** | Daylio and Journey both support searching entries by text or filtering by mood/activity. | MEDIUM | Search by text, filter by sentiment/date range. |
| **Data privacy** | All apps emphasize privacy. Daylio keeps data device-only; Bearable encrypts. This is non-negotiable for emotional data. | MEDIUM | Kanso's PouchDB/CouchDB sync means data lives on server. Must communicate privacy clearly and encrypt at rest. |
| **Dark mode** | Daylio, Journey, Bearable all support it. Expected in 2026 for any app used in low-light emotional moments. | LOW | CSS variable swap — trivial with Tailwind. |
| **Password/biometric lock** | Daylio PIN/Face ID, Journey passcode/biometric, Bearable has it. Users journal intimate feelings. | LOW | PWA limitation: no native biometric API (WebAuthn partially available). PIN fallback needed. |
| **Charts / mood trends** | Daylio's mood line + correlations, Bearable's advanced reports. Users expect visual patterns. | MEDIUM | Line chart over time, mood distribution pie/bar. Should be exportable for therapist. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Therapist report via WhatsApp** | Unique bridge between personal journaling and clinical therapy. No major app does this natively. Daylio/Bearable require manual PDF export + email. | HIGH | Core Kanso differentiator. Requires Twilio integration + PDF generation. **This is the killer feature.** |
| **NLP emotion analysis in Portuguese** | English NLP is table stakes (Youper, Journey Odyssey AI). Portuguese emotion analysis with decent accuracy is rare. Daylio has no NLP. Bearable no NLP. Journey Odyssey AI is English-only. | HIGH | Requires fine-tuned Portuguese model or transformer pipeline. This is technically risky — accuracy for emotional NLP in Portuguese is lower than English. |
| **Retroactive entry (backfill)** | Most apps only log "right now." Supporting backfill (with "I felt this earlier today") acknowledges real life. Daylio only does current. Bearable does current and time adjustment. | LOW | Kanso already has this requirement. Simple datetime picker. |
| **Custom sentiment labeling (discovered in therapy)** | Bearable lets you create custom feelings. Kanso's sentiment combobox is intentionally customizable so users discover emotion names with their therapist. | LOW | Already planned. This is stronger than Daylio's fixed emoji scale. Matches Bearable's model. |
| **Offline-first with automatic sync** | Many apps require internet. PouchDB/CouchDB means entries save immediately and sync when connectivity returns. Journey offers cloud sync but pures offline journaling apps (Daylio) store only locally. | HIGH | Already planned. Requires handling sync conflicts gracefully. |
| **Periodic report (manual generation, covers period since last)** | Bearable has similar "sharing with doctor" but you must select date range manually. Kanso's "since last report" is more focused on therapy pacing. | MEDIUM | Unique approach — less choice but more therapeutic structure: covers exactly the period since your last session. |
| **Structured entry fields (sensation, feeling, context, thought)** | Daylio is mood + activities + optional note. Kanso breaks the entry into 4 targeted fields that map to CBT/journaling techniques. | LOW | A differentiator for clinical usefulness — but risk of being slower than Daylio's 2-tap model. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Social features / community** | Users want validation. Sanvello has community groups. | Emotional data is deeply personal. Social features create privacy risk, moderation burden, and can trivialize serious emotions. Users compare "who's more depressed" — toxic. | Keep it 1:1 with therapist only. No feeds, no likes, no comments. |
| **Gamification (streaks, badges, points)** | Daylio has achievements/goals. Users feel motivated by streaks. | Emotional health is not a game. Streaks create guilt when users skip logging during genuinely bad days. This undermines the therapeutic purpose — you should rest, not feel bad about breaking a streak. | Optional personal goals only (like Bearable's approach). No badges, no leaderboards, no "leveling up." |
| **AI chatbot therapy** | Youper built their whole product around this. Users want 24/7 emotional support. | Dangerous in therapeutic context. AI cannot replace a therapist. Youper now has extensive crisis/suicide prevention disclaimers. Liability risk. False sense of support. | Kanso's NLP is analytical only (pattern detection, not conversation). Keep it that way. |
| **Multiple therapists** | Scalability mindset. Users might change therapists. | Adds therapist management UI, therapist-side auth, multi-Twilio routing. Bloat for MVP. Only 1 therapist needed per user. | Single therapist per user (already decided). If needed later: simple "change therapist" action, not full multi-management. |
| **Real-time chat with therapist** | Users want immediate feedback. Sanvello-like coaching. | Blurs boundary between journaling and therapy. Creates expectation of immediate response. Adds complexity (websockets, moderation). | Scheduled reports only. WhatsApp is not real-time chat with therapist — it's a file delivery channel. Keep boundary clear. |

## Feature Dependencies

```
Quick mood entry
    └──requires──> Authentication (Google OAuth)
    └──requires──> Offline storage (PouchDB)

History view
    └──requires──> Quick mood entry (data to show)
    └──enhances──> Charts / mood trends

Reminders
    └──requires──> Quick mood entry (what to remind about)
    └──requires──> Push notification infrastructure (FCM)

Search / filter
    └──requires──> History view (UI to search within)
    └──enhances──> Charts / mood trends (filtered analysis)

NLP emotion analysis
    └──requires──> Quick mood entry (text to analyze)
    └──enhances──> Charts / mood trends (with emotion breakdown)
    └──enhances──> Therapist report (enriched report)

Therapist report PDF
    └──requires──> Quick mood entry (data to include)
    └──requires──> PDF generation library
    └──enhances──> NLP emotion analysis (included in report)

WhatsApp send
    └──requires──> Therapist report PDF (what to send)
    └──requires──> Therapist phone number on profile
    └──requires──> Twilio API integration

Dark mode, Biometric lock, Custom sentiment
    └─── no hard dependencies — can be added at any time
```

### Dependency Notes

- **Therapist report requires entries:** Duh. But also means the "period since last report" logic needs a stored `lastReportDate` per user.
- **NLP enriches everything:** NLP analysis should be fire-and-forget (async) so entries save immediately and NLP results arrive later. This prevents a slow NLP pipeline from blocking the entry flow.
- **WhatsApp dependency chain is deep:** PDF → entries + NLP → NLP → entries. This means v0.x must have entries and NLP before WhatsApp reports work.
- **Reminders require push infrastructure:** Push notifications in PWAs need FCM service worker + permission flow. User must grant notification permission before reminders work.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **Google OAuth authentication** — single sign-on, prepares for scale
- [x] **Quick mood entry** — combobox sentiment + free text fields (sensação, contexto, pensamento)
- [x] **Retroactive entry** — datetime picker for backdating
- [x] **Offline-first storage** — PouchDB with auto-sync to CouchDB
- [x] **History view** — chronological list of entries
- [x] **Tab navigation** — Register (default), History, Profile/Config

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **Push notification reminders** — configurable times (12, 18, 23h defaults)
- [ ] **Charts / mood trends** — basic line chart and distribution
- [ ] **Dark mode** — toggleable
- [ ] **Custom sentiment** — editable combobox entries
- [ ] **Search / filter** — by text, date range, sentiment

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **NLP emotion analysis** — Portuguese transformer model. Defer because: model accuracy risk, infrastructure cost, and the app works without it
- [ ] **Therapist PDF report** — requires NLP to be truly valuable (otherwise it's just a data dump)
- [ ] **WhatsApp integration** — depends on PDF report existing
- [ ] **Biometric/PIN lock** — PWA limitations; WebAuthn is still evolving on mobile browsers

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority | Rationale |
|---------|------------|---------------------|----------|-----------|
| Quick mood entry | HIGH | LOW | **P1** | Core action. Without this, no app exists. |
| Offline-first storage | HIGH | HIGH | **P1** | Without it, mobile users lose data on subway. |
| History view | HIGH | LOW | **P1** | Users need to see what they logged. |
| Google OAuth | HIGH | MEDIUM | **P1** | Required for sync and scale. |
| Retroactive entry | MEDIUM | LOW | **P1** | Planned — low cost, high practical value. |
| Push reminders | HIGH | MEDIUM | **P2** | Users forget. But app works without them. |
| Charts / mood trends | HIGH | MEDIUM | **P2** | Key for pattern recognition. |
| Dark mode | MEDIUM | LOW | **P2** | Expected but app works without. |
| Custom sentiment | HIGH | LOW | **P2** | Therapeutic value but not blocking launch. |
| Search / filter | MEDIUM | MEDIUM | **P2** | Nice when history grows. |
| NLP emotion analysis | HIGH | HIGH | **P3** | Key differentiator but expensive and risky. Validate core first. |
| PDF report | HIGH | MEDIUM | **P3** | Requires NLP for real value. |
| WhatsApp send | HIGH | HIGH | **P3** | Requires PDF. Complex dependency chain. |
| Biometric/PIN lock | MEDIUM | MEDIUM | **P3** | PWA limitation. Wait for better WebAuthn support. |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Strategic Feature Decisions for Kanso

### What Kanso Does Differently from Competitors

1. **Therapist-as-primary-consumer**: Most apps are self-help or share-with-doctor-as-optional. Kanso is built around the therapist relationship from day one. The report + WhatsApp flow is the central value proposition, not an add-on.

2. **Structured fields over freeform**: Daylio is mood + activity only. Journey is freeform journal. Kanso's 4-field structure (sensation, feeling, context, thought) mirrors therapeutic techniques (CBT, naming emotions). This is more rigid but more clinically useful.

3. **Portuguese-first**: No major emotion diary handles Portuguese NLP well. This is a genuine gap.

### What Kanso Deliberately Avoids

- **Not a replacement for therapy** — it's an augmentation tool. Youper-style AI therapy is dangerous to attempt.
- **Not a social network** — no community, no sharing (except with therapist), no likes.
- **Not a gamified habit tracker** — no streaks, no badges, no competition. Bearable and Daylio have goals which can be useful, but emotional health gamification is a fine line.
- **Not a multi-therapist platform** — single therapist, simple relationship.

### When to Reconsider Anti-Features

| Anti-Feature | Reconsider If |
|--------------|---------------|
| Optional goals | User research shows people want gentle personal targets (not streaks). Bearable-style. |
| Change therapist | User changes therapists or takes a break. Simple "transfer" option, not multi-management. |

## Competitor Feature Analysis

| Feature | Daylio | Journey | Bearable | Youper | Kanso (planned) |
|---------|--------|---------|----------|-------|-----------------|
| Mood entry speed | 2 taps (gold standard) | Several taps + typing | Several taps + customization | Chat-based | Combobox + free text (medium) |
| Offline support | Yes (local only, no cloud) | Cloud sync (paid) | Cloud sync | Requires internet | **Offline-first (PouchDB + CouchDB)** |
| NLP / AI analysis | **None** | Odyssey AI (English, journal Q&A) | **None** (correlations only) | AI chatbot (English, therapeutic) | **Portuguese emotion NLP** (planned) |
| Therapist sharing | Manual export → share | Manual PDF export | Doctor report (paid) | **None** | **Auto PDF → WhatsApp** |
| Emotion granularity | 5-point emoji scale | Mood picker | Customizable feelings + severity | Chat-based inference | **Customizable combobox** |
| Privacy model | Device-only | E2E encryption (paid) | Encrypted at rest | Server-side | Offline-first + server sync |
| Reminders | Yes | Yes | Yes | In-app | Push notifications |
| Multimedia entries | Photos (paid) | **Photos, video, audio** | **None** | Text only | Text only |
| Portuguese support | Interface only | Interface only | Interface only | Interface + limited | **Full (interface + NLP)** |
| Active user base | 20M+ | Millions | ~500K (est.) | ~1M (est.) | 1 (developer) |

## Sources

- [Daylio official site](https://daylio.net/) — HIGH confidence (official marketing)
- [Journey official site](https://journey.cloud/) — HIGH confidence (official marketing)
- [Journey vs Daylio comparison](https://journey.cloud/daylio-alternative/) — HIGH confidence (official comparison, likely biased)
- [Bearable mood tracker page](https://bearable.app/mood-tracker-app-journal/) — HIGH confidence (official site)
- [Youper official site](https://www.youper.ai/) — HIGH confidence (official site, primarily legal/medical disclaimers)
- PROJECT.md requirements — HIGH confidence (project-specific)

---
*Feature research for: Kanso — therapeutic emotion diary PWA*
*Researched: 2026-05-16*
