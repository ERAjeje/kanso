# Pitfalls Research: Kanso

**Domain:** Offline-first PWA emotion diary with CouchDB sync, Go/Python backend, PDF generation, push notifications
**Researched:** 2026-05-16
**Overall Confidence:** HIGH (verified via official docs and well-known failure patterns)

## Critical Pitfalls

Mistakes that cause data loss, security breaches, or major rewrites.

---

### Pitfall 1: Silent Data Loss from Unhandled PouchDB Conflicts

**What goes wrong:**
Emotion entries are silently lost or overwritten when two devices (or the PWA and a server-side update) modify the same document concurrently. The user records a feeling, syncs later, and the entry disappears or shows stale data.

**Why it happens:**
CouchDB uses Multi-Version Concurrency Control (MVCC). Every update requires the correct `_rev`. PouchDB's live replication defaults to "last write wins" via deterministic algorithm, but conflicts accumulate in the revision tree. If the code never checks `_conflicts`, the losing revision's data is invisible — the user's entry is effectively lost.

The leading cause: calling `put()` without reading the latest `_rev` first, or not handling the 409 conflict error properly.

**How to avoid:**
1. Always fetch the latest doc (including `{conflicts: true}`) before updating
2. Use `pouchdb-upsert` plugin (or implement upsert pattern) for delta-based updates where possible
3. Design docs as immutable "append-only" entries (accountant pattern) to avoid conflicts entirely
4. Register a `changes` listener that monitors for `_conflicts` and resolves them automatically (merge fields or last-write-wins per field)

**Warning signs:**
- Random 409 errors in browser console during sync
- Users reporting "my entry from yesterday is gone" even though sync says "OK"
- Growing `_rev` chain length (>10 revisions per doc)

**Phase to address:**
Phase 1 (Core Sync Infrastructure) — must build conflict resolution into the sync layer from day one. Retrofitting is painful.

---

### Pitfall 2: CouchDB Exposed Without Proper Auth — Emotional Data Leak

**What goes wrong:**
The PWA syncs directly with CouchDB over the public internet. Without proper authentication, anyone who discovers the CouchDB URL can read, modify, or delete every emotional diary entry for every user.

**Why it happens:**
PouchDB's `sync()` requires a URL to CouchDB. The naive approach exposes CouchDB on a public port with only admin credentials. The apontamentos.md correctly identifies using Traefik proxy-auth with `X-Auth-CouchDB-User` header, but the security model is fragile:
- If CouchDB's `_users` database isn't properly configured, any authenticated user can read all data
- If `_security` per-database isn't set, users can access other users' databases
- If Traefik's JWT validation fails open, unauthenticated requests reach CouchDB

**How to avoid:**
1. **Never expose CouchDB directly to the internet.** Always proxy through Traefik.
2. Configure CouchDB with `require_valid_user = true` in `[couchdb]` section
3. Use per-user databases (`user-db-{userid}`) with security docs restricting access to that user only
4. On Traefik: validate JWT in a ForwardAuth middleware, inject `X-Auth-CouchDB-User` header, and configure CouchDB for proxy authentication
5. Enable CORS properly — only allow the PWA origin
6. Set `default_security = everyone` or explicitly configure each database's `_security`

**Warning signs:**
- CouchDB accessible on a public port (5984) without Traefik in front
- No `_security` document on databases
- `require_valid_user` is set to `false` (default)

**Phase to address:**
Phase 0 (Infrastructure/Architecture) — auth proxy must be designed before any data flows.

---

### Pitfall 3: Live Replication Dies Silently on Network Loss

**What goes wrong:**
The user opens the app, views their history, but it shows stale data because PouchDB live replication silently stopped. New entries made on another device never arrive. The user thinks their data is synced, but it's not — compromising therapeutic continuity.

**Why it happens:**
PouchDB's `sync({ live: true })` without `{ retry: true }` stops replication on the first network error. Mobile browsers regularly lose connectivity. Live replication fires an `error` event and then goes silent. Most developers only listen for `'change'` events and miss the `'error'` and `'paused'` events entirely.

**How to avoid:**
1. Always use `{ live: true, retry: true }` for mobile PWA sync
2. Implement a sync status indicator in the UI (connected/syncing/offline/error)
3. Listen for `'paused'` and `'active'` events to update UI state
4. Cancel and restart replication on `'denied'` errors (auth token expired)
5. Consider adding a periodic "force sync" on a timer as a fallback

**Warning signs:**
- No sync status indicator in the UI
- Console shows "PouchDB replication error" but app continues as if nothing happened
- User sees different data on different devices

**Phase to address:**
Phase 1 (Core Sync Infrastructure) — retry logic and sync status UI are core UX features, not afterthoughts.

---

### Pitfall 4: chromedp Chrome Instance Memory Exhaustion in Docker

**What goes wrong:**
When the user requests a PDF report, the Go backend launches a headless Chrome instance via chromedp. Under concurrent requests, each Chrome process consumes 200–500 MB of RAM. The container runs out of memory, the OOM killer terminates processes, and the PDF generation fails — or worse, the entire backend container is killed.

**Why it happens:**
- chromedp launches a full headless Chrome process per context (not lightweight)
- Default Docker memory limits aren't set, or are set too low
- Chrome crashes with `BUS_ADRERR` when `/dev/shm` is too small in Docker (default 64MB)
- No request queuing — concurrent PDF requests launch multiple Chrome instances simultaneously

**How to avoid:**
1. Use the `chromedp/headless-shell` Docker image (smaller, headless-only, ~140MB vs full Chrome)
2. Set Docker memory limit: `--memory="1g"` and `--memory-reservation="512m"`
3. Increase shared memory: `--shm-size=2g` (critical — BUS_ADRERR without this)
4. Implement a PDF generation queue with max 1 concurrent Chrome instance (channel-based semaphore in Go)
5. Set a generous but firm context timeout: `context.WithTimeout(ctx, 60*time.Second)`
6. Reuse a single Chrome allocator context rather than creating a new one per request

**Warning signs:**
- PDF generation fails intermittently under load
- Container logs show "OOMKilled"
- Chrome crashes with "Bus error" or "BUS_ADRERR"
- PDF generation takes increasingly longer time

**Phase to address:**
Phase 2 (Reporting/PDF) — Chrome resource management must be designed alongside PDF generation, not patched in later.

---

### Pitfall 5: Push Notifications Show Chrome's Default "This site has been updated" Message

**What goes wrong:**
Instead of the app's reminder notification ("Time to record your feelings"), the user sees Chrome's generic "This site has been updated in the background" notification. This confuses the user, reduces trust, and defeats the purpose of reminders.

**Why it happens:**
Chrome shows the default notification when the push event in the service worker doesn't call `self.registration.showNotification()` within the `event.waitUntil()` promise chain. The most common cause:
- Calling `showNotification()` but not returning the promise from `event.waitUntil()`
- Async operations (fetch, cache) that fail before `showNotification()` is called
- Promise chain that breaks silently

**How to avoid:**
1. Always return the `showNotification()` promise inside `event.waitUntil()`:
   ```js
   self.addEventListener('push', (event) => {
     event.waitUntil(
       self.registration.showNotification('Kanso', { body: '...' })
     );
   });
   ```
2. Never branch without showing a notification in every code path
3. Handle errors explicitly — if data fetch fails, show a default reminder anyway
4. Test push notifications on real devices (Chrome on Android, not just desktop)

**Warning signs:**
- Notifications say "This site has been updated" instead of app content
- Promise returned from `waitUntil` doesn't include `showNotification` return value
- Uncaught exceptions in push event handler

**Phase to address:**
Phase 4 (Push Notifications) — Service worker push handler is the most error-prone code path in the entire app. Must be tested with real FCM.

---

### Pitfall 6: JWT Expiration Kills Active Sessions Without Graceful Refresh

**What goes wrong:**
The user is typing a detailed emotional entry. They hit save. The sync fails with a 401. The JWT expired while they were writing. The entry is lost. The user must re-authenticate and loses their train of thought — terrible for a therapy tool where capturing the moment matters.

**Why it happens:**
- JWT has a short expiration (e.g., 1 hour) for security
- No refresh token flow — the app doesn't detect expiration before making requests
- PouchDB sync doesn't know about JWT — it just makes HTTP requests that fail with 401
- No retry mechanism that first refreshes the token, then replays the failed request

**How to avoid:**
1. Implement a token refresh mechanism alongside Google Sign-In:
   - Store refresh token securely (httpOnly cookie or encrypted localStorage)
   - On 401 during sync or API call: attempt refresh, retry original request
2. Use a PouchDB HTTP plugin or wrapper that:
   - Captures 401 responses
   - Refreshes the JWT
   - Updates the `Authorization` header
   - Retries the failed replication request
3. Keep Traefik and CouchDB proxy-auth in sync — ensure the new JWT is immediately valid
4. Set JWT expiration to a reasonable window (e.g., 24h for a therapy app where sessions are infrequent but long)

**Warning signs:**
- 401 errors in PouchDB sync logs
- Users reporting "I got logged out while writing"
- No refresh token mechanism in auth flow

**Phase to address:**
Phase 1 (Auth Infrastructure) — JWT refresh must be designed alongside initial auth, not added later.

---

### Pitfall 7: NLP Service as a Hard Dependency — PDF Generation Fails When NLP Is Down

**What goes wrong:**
The user requests a PDF report. The Go backend calls the NLP service to enrich emotion data. The NLP service is slow or down (container restarting, model loading, OOM during inference). The PDF generation fails entirely — even though all the data is already in CouchDB and the NLP enrichment is just a nice-to-have.

**Why it happens:**
- The PDF generation flow treats NLP analysis as a mandatory step
- No fallback to "unanalyzed" data when NLP is unavailable
- Go HTTP client default timeout (none!) causes hangs forever
- No circuit breaker — repeated failures cascade

**How to avoid:**
1. **Graceful degradation:** NLP enrichment is optional in PDF. If NLP is unavailable, generate the PDF with unanalyzed data but add a note: "Emotion analysis unavailable"
2. Always set HTTP client timeout on calls to NLP service: `&http.Client{Timeout: 10 * time.Second}`
3. Implement a circuit breaker pattern: after N failures, stop calling NLP for a cooldown period
4. Store NLP results in the CouchDB document (already planned in `analiseEmocoes` field) so PDF can use cached results even if NLP is down
5. Make NLP analysis async — don't block the report generation on real-time analysis. Analyze entries in background after they're created.

**Warning signs:**
- PDF generation depends on NLP service being available
- Go HTTP client uses default zero timeout (blocks forever)
- No cached NLP results in CouchDB documents
- Synchronous NLP analysis blocking user-facing features

**Phase to address:**
Phase 3 (NLP Analysis) — design NLP as an enhancement, not a dependency. Implement caching from the start.

---

### Pitfall 8: WhatsApp PDF Sharing Leaks Emotional Data

**What goes wrong:**
The PDF report — containing intimate emotional data, thoughts, and context — is sent via Twilio's WhatsApp API. The PDF URL (if stored on a server without auth) is accessible to anyone who has the link. Or the WhatsApp message is delivered to the wrong recipient. Or the PDF contains identifiable information in metadata.

**Why it happens:**
- PDF stored on a temporary URL without authentication
- No encryption of the PDF content
- PDF metadata (author name, creation date, software) leaks information
- Phone number of the psychologist is incorrect or changes
- No confirmation that the recipient actually received it
- Twilio logs (if not managed) retain message content

**How to avoid:**
1. Serve PDFs only through authenticated endpoints (JWT required), never public URLs
2. Generate PDF URLs with short expiration (e.g., 5 minutes) using signed tokens
3. Strip PDF metadata: no author, no creator field, no software identification
4. Notify the user in-app when the report is sent, with delivery confirmation
5. Use Twilio's webhook to confirm delivery status and surface errors to the user
6. Store psychologist's phone number encrypted in CouchDB (field-level encryption or at least encrypted document)
7. Add a "This report contains sensitive emotional data — please treat as confidential" header in the PDF

**Warning signs:**
- PDF URLs without authentication (/download/{jobId} without token)
- PDF metadata contains identifiable information
- No delivery confirmation mechanism
- Psychologist phone stored in plain text

**Phase to address:**
Phase 3 (WhatsApp/Reporting) — privacy must be designed into PDF sharing, not bolted on.

---

### Pitfall 9: Docker Compose Dev/Prod Config Drift

**What goes wrong:**
The app works perfectly in `docker compose up` on the developer's machine but fails in production. The CouchDB data doesn't persist across restarts. The NLP service crashes because it needs more memory. Push notifications fail because FCM credentials aren't set. The backend and CouchDB start before CouchDB is ready to accept connections.

**Why it happens:**
- Single `docker-compose.yml` used for both dev and prod with no overrides
- Volume mounts for hot-reload in dev are also used in production (or missing entirely)
- No `restart: always` on services
- No health checks — services start in random order
- `latest` image tags instead of pinned versions
- Environment variables for secrets are hardcoded or from `.env` that doesn't exist in production
- No resource limits (memory/cpu) for containers

**How to avoid:**
1. Use 3 Compose files:
   - `compose.yml` — shared base config
   - `compose.dev.yml` — dev overrides (volume mounts, debug ports, hot-reload)
   - `compose.prod.yml` — prod overrides (resource limits, restart policies, no bind mounts)
2. Run dev: `docker compose -f compose.yml -f compose.dev.yml up`
3. Run prod: `docker compose -f compose.yml -f compose.prod.yml up -d`
4. Add health checks to every service
5. Pin all images to specific versions (not `:latest`)
6. Use Docker secrets or `.env.production` for environment-specific config
7. Set memory limits: `deploy.resources.limits.memory: "512M"` for NLP, `"1G"` for backend+Chrome

**Warning signs:**
- Single `docker-compose.yml` with no override files
- `:latest` tags on services
- No `restart:` policy
- No `healthcheck` on database services
- Volume mounts for source code in what should be a production setup
- Missing `.dockerignore` — bloated images with node_modules

**Phase to address:**
Phase 0 (Infrastructure) — multi-file Compose setup from day one prevents drift.

---

### Pitfall 10: User Fatigue from Poorly Timed Notifications

**What goes wrong:**
The user downloads Kanso genuinely wanting to track their emotions. But notifications come at fixed times (12, 18, 23h) regardless of the user's sleep schedule, work hours, or emotional state. The user starts ignoring notifications, then disables them entirely, then forgets about the app. The therapeutic benefit is lost.

**Why it happens:**
- Notification times are hardcoded or set once during onboarding
- No capability to "snooze" a notification
- No intelligent scheduling (e.g., skip if user already registered today)
- No option to temporarily pause notifications (vacation, sick days, therapy break)
- The app doesn't adapt to user behavior patterns

**How to avoid:**
1. Make notification schedule fully configurable in-app with granular time picker
2. Add a "snooze for 1 hour" action button on the notification itself
3. Implement smart suppression: don't notify if user has already registered in the last 4 hours
4. Add a "pause notifications" toggle (1 day, 1 week, custom)
5. Allow notification payload to include custom message based on time of day
6. Show positive reinforcement stats ("You've tracked 15 days in a row!") to maintain motivation

**Warning signs:**
- Users disable notifications within the first week
- App uninstalled after 2-3 weeks
- Single fixed notification schedule with no customization
- No snooze or pause functionality

**Phase to address:**
Phase 4 (Push Notifications) — notification UX design is equally important as the technical implementation.

---

### Pitfall 11: Go Backend Assumes NLP Service Is Always Available

**What goes wrong:**
The Go backend calls the Python NLP service synchronously during entry creation. The NLP container is still starting up (or crashed, or running out of memory loading the model). The Go request hangs, the user sees a spinner that never resolves, and the entry they just typed is lost. Every subsequent entry creation also fails, cascading the outage.

**Why it happens:**
- Go's default `http.Client` has **no timeout** (it will wait forever)
- No circuit breaker pattern — the backend keeps calling a failing service
- No retry with backoff for transient failures
- NLP model loading on first request (cold start can take 10-30 seconds)
- Sequential request handling — one slow NLP call blocks other requests

**How to avoid:**
1. **Always** set explicit timeout on the Go HTTP client:
   ```go
   client := &http.Client{Timeout: 10 * time.Second}
   ```
2. Implement circuit breaker using `sony/gobreaker` or similar:
   - After 3 failures in 60 seconds: open circuit for 30 seconds
   - Return cached or "analysis unavailable" response during open circuit
3. Make NLP analysis async from entry creation:
   - Save entry to CouchDB immediately (no NLP)
   - Push a background job to analyze after save
   - Store results in `analiseEmocoes` field when done
4. Pre-warm the NLP model on container startup (don't lazy-load on first request)
5. Add proper health check for NLP service that accounts for model loading time (~15-30s)

**Warning signs:**
- Go HTTP client created with default zero timeout
- NLP analysis blocks the user-facing API response
- No cached/fallback responses when NLP is down
- Service starts getting timeouts after model reloads

**Phase to address:**
Phase 3 (NLP Integration) — async analysis with circuit breaker from day one.

---

### Pitfall 12: NotificationClick Opens Wrong Page or Fails on Mobile

**What goes wrong:**
The user taps a push notification reminder. Instead of opening the "Register" tab where they can record their feeling, it opens the app's generic home page or — worse — does nothing. The user, in the moment of feeling an emotion, has to navigate to find the entry form. The emotional moment passes. The entry is lost.

**Why it happens:**
- Service worker `notificationclick` event isn't handled
- The event handler doesn't use `event.waitUntil()` — the service worker may terminate before the page opens
- The click targets a generic URL instead of the specific entry form URL with parameters
- On some mobile browsers, `clients.openWindow()` navigates away from the current app state instead of focusing the existing window
- PWA scope isn't configured properly — clicks outside scope do nothing

**How to avoid:**
1. Always handle `notificationclick` in the service worker:
   ```js
   self.addEventListener('notificationclick', (event) => {
     event.notification.close();
     event.waitUntil(
       clients.matchAll({ type: 'window', includeUncontrolled: true })
         .then((clientList) => {
           if (clientList.length > 0) {
             return clientList[0].focus();
           }
           return clients.openWindow('/');
         })
     );
   });
   ```
2. Open the specific tab/form: include a `url` or `data` in the notification payload that directs to `/?tab=register`
3. Use `clients.matchAll` to find existing windows before opening new ones (prevents duplicate tabs)
4. Test on actual mobile devices — desktop Chrome behavior differs from mobile

**Warning signs:**
- No `notificationclick` handler in service worker
- Clicking a notification does nothing or opens the wrong page
- App opens as a new tab instead of focusing the existing one
- Only tested on desktop, never on mobile Chrome

**Phase to address:**
Phase 4 (Push Notifications) — notification interaction flow is as important as delivery.

---

### Pitfall 13: Emotion Entry Lost Due to PWA Cache Eviction

**What goes wrong:**
The user types a detailed emotional entry offline (on a bus, in a remote area). Before it syncs, the browser decides to reclaim disk space. PouchDB's IndexedDB storage is partially or fully evicted. The entry is gone. The user has no way to know this happened.

**Why it happens:**
- Browsers (especially Safari, Chrome on iOS, and Android under storage pressure) can delete IndexedDB data without warning
- PWA without proper storage persistence request is at the mercy of the browser's LRU eviction policy
- No feedback to the user that data was lost
- `navigator.storage.persist()` was never requested or denied

**How to avoid:**
1. Request persistent storage at app startup:
   ```js
   if (navigator.storage && navigator.storage.persist) {
     const isPersisted = await navigator.storage.persist();
     if (!isPersisted) {
       // Log and warn — data may be evicted
     }
   }
   ```
2. Check storage estimate periodically and warn user if space is low
3. Show sync status for each entry — green checkmark when confirmed synced to CouchDB
4. Consider using PouchDB with LevelDB adapter (in Node/Electron) or SQLite adapter for better persistence guarantees

**Warning signs:**
- Entries that were created but never confirmed as synced disappear
- Browser storage notification appears ("Site is using storage space")
- `navigator.storage.persisted()` returns `false`
- No sync confirmation indicator per entry

**Phase to address:**
Phase 1 (Core Sync Infrastructure) — storage persistence should be requested during initial PWA setup.

---

### Pitfall 14: Emotion Model Selection Fails for Portuguese

**What goes wrong:**
The NLP model selected (`nlptown/bert-base-multilingual-uncased-emotion` or `pysentimiento/robertuito-emotion-analysis`) performs poorly on Brazilian Portuguese emotional text. The user writes "aperto no peito e vontade de chorar" (chest tightness and urge to cry) and the model returns "neutral" or misclassifies the emotion. The analysis becomes useless, and the user loses trust.

**Why it happens:**
- Many emotion models are trained on English data (GoEmotions, etc.) or other languages
- Multilingual models often prioritize English, Spanish, or French over Portuguese
- Portuguese emotion-specific models are rare and may not cover the emotion taxonomy the app needs
- No testing with real Portuguese emotional text during model selection
- The model's confidence threshold is set too high, filtering out most results

**How to avoid:**
1. **Test before committing:** Run the selected model against a sample of 20-30 Portuguese emotional phrases before implementation
2. Prefer models with demonstrated Portuguese support:
   - `joaoalvarenga/bert-base-portuguese-cased-emotion` — trained on Portuguese text
   - `pysentimiento/robertuito-emotion-analysis` — at least has Spanish (closer than English)
   - Test multilingual models for actual Portuguese performance
3. Lower confidence thresholds for Portuguese (0.3-0.5 instead of 0.7+) since models are less confident
4. Cache and allow users to provide feedback on emotion classification
5. Consider fine-tuning a model on Portuguese emotional text if accuracy is insufficient

**Warning signs:**
- Model never returns emotion probabilities above 0.5 for Portuguese text
- Emotions returned don't make sense for the input text
- Only tested with English samples; no Portuguese evaluation set
- No mechanism for user feedback on classification accuracy

**Phase to address:**
Phase 3 (NLP Analysis) — model selection must include Portuguese evaluation before integration.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| **Single docker-compose.yml** | Simpler setup | Staging/prod drift, hard to reproduce issues | Never — use overrides from day one |
| **No sync status UI** | Faster initial build | Users don't know if data is synced; data loss invisible | MVP only (single user who knows the system) |
| **No conflict resolution** | Simpler initial sync | Silent data loss when sharing across devices | Never — even single user can have conflicts |
| `latest` **image tags** | Always "fresh" | Non-reproducible builds, unexpected breaking changes | Never — pin versions everywhere |
| **NLP analysis on entry creation (sync)** | Real-time enrichment | User-facing latency, cascading failures when NLP is down | Never — make it async |
| **Skipping health checks** | Faster docker-compose setup | Services start in wrong order, race conditions, hard to debug | Dev-only, never in production |
| **No timeout on Go HTTP client** | Simpler code (default is zero) | Hangs forever when services are down | Never — always set explicit timeouts |
| **Single `_users` database for all CouchDB users** | Easier setup | Security boundary issues, no per-user isolation | Never — use per-user databases or validate |
| **Default browser notifications without custom icon/sound** | Faster implementation | Lower user engagement, app feels generic | Only during initial notification setup |
| **Hardcoded psychologist WhatsApp number** | Simpler code | Cannot change therapist, no multi-therapist support | MVP only if single-user |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| **PouchDB → CouchDB sync** | No `{retry: true}` on live sync | Always use `{live: true, retry: true}` |
| **PouchDB → CouchDB auth** | Sending credentials in every request | Use Traefik proxy-auth with JWT, never expose CouchDB directly |
| **Go → NLP (Python)** | Default HTTP client (no timeout) | `&http.Client{Timeout: 10 * time.Second}` with circuit breaker |
| **Go → CouchDB (admin)** | Using admin credentials in app code | Use a restricted CouchDB user with per-database access only |
| **Go → Twilio WhatsApp** | Sending PDF as attachment in message | Use Twilio Media API; ensure PDF URL is authenticated and temporary |
| **PWA → FCM** | Not handling token refresh | Listen for `pushsubscriptionchange` event and re-register |
| **Traefik → CouchDB proxy** | Forwarding all headers blindly | Only forward `X-Auth-CouchDB-User` from validated JWT |
| **Go → chromedp** | Creating new allocator per request | Reuse allocator context; limit concurrent Chrome instances |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| **Full database sync for all users** | Slow initial load, high bandwidth | Per-user databases with filtered sync; use `_selector` for replication filtering | >100 entries per user |
| **Synchronous NLP on every entry** | Spinner on save, timeout failures | Async analysis with stored results in document | >5 concurrent users |
| **No pagination on history view** | UI freezes, slow rendering | Windowed/Iceberg queries with PouchDB Mango pagination | >500 entries |
| **Creating new Chrome process per PDF** | OOM crashes, slow PDF generation | Chrome allocator pooling, max 1 concurrent generation | >2 concurrent PDF requests |
| **Emotion analysis on full text** | Slow inference, memory spikes | Truncate to 512 tokens (BERT limit), batch if needed | >1000 chars per entry |
| **Unoptimized CouchDB views** | Slow queries, high disk I/O | Use Mango indexes for simple queries; design docs for complex ones | >10K documents per DB |
| **Default IndexedDB quota** | Data eviction without warning | Request `navigator.storage.persist()` at startup | After browser storage pressure |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| **CouchDB exposed to internet** | Anyone can read emotional diary entries | Always proxy through Traefik with JWT validation |
| **No per-user database isolation** | Users can access other users' entries | Per-user CouchDB databases with `_security` docs |
| **PDF download without auth** | Report of emotional data accessible by anyone | Signed, short-expiry URLs; JWT on download endpoint |
| **JWT without refresh mechanism** | Users lose data on token expiry | Refresh token flow; automatic retry on 401 during sync |
| **WhatsApp number stored in plaintext** | Psychologist contact exposed if DB compromised | Encrypt sensitive user data in CouchDB |
| **NLP service exposed internally without auth** | Attackers could submit arbitrary text for analysis | Internal Docker network only; no public NLP endpoint |
| **Psychologist phone not validated** | Wrong recipient receives emotional data | Validate phone number format; confirm before first send |
| **FCM API key in client code** | Attacker could send notifications to all users | FCM key on server only; client only has push subscription |
| **No rate limiting on auth endpoint** | Brute force Google auth callback | Rate limit at Traefik level on `/api/auth/*` |
| **IndexedDB data not encrypted at rest** | Emotional data readable if device compromised | No built-in encryption in browsers; rely on device-level encryption |

## UX Pitfalls

Common user experience mistakes in emotion diary apps.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| **No guidance on how to name emotions** | Users get stuck, don't know what to write | "What are you feeling?" prompt with examples ("worried, excited, heavy...") |
| **Too many required fields** | High friction, user skips entries | Make only datetime required; everything else optional |
| **Emotion combobox is overwhelming** | Choice paralysis | Start with 6-8 common emotions, allow adding more organically |
| **Notifications at fixed times only** | "I don't feel what I'm supposed to feel right now" | Allow on-demand "I want to record now" from notification |
| **Retroactive entry is confusing** | "Can I record yesterday's feeling?" confusion | UI: "Record now" vs "Record for a past moment" with clear visual split |
| **No streaks or progress** | Low motivation to continue | Show "You've recorded X days in a row" + gentle encouragement |
| **Report PDF is text-heavy** | Psychologist has to read through everything | Include emotion summary, timeline graph, most frequent feelings |
| **No search/filter on history** | Hard to find patterns | Filter by emotion, date range, free text search |
| **No context for what "context" means** | Users don't fill the context field | "What was happening? (a meeting, a conversation, alone, with friends...)" |
| **Form doesn't save on app close** | User types a long entry, app crashes, all lost | Save to PouchDB on every field change (debounced), not on form submit |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **PouchDB sync:** Looks set up, but missing `{retry: true}` and conflict resolution — verify the `error` event handler exists
- [ ] **Push notifications:** Notification appears but no `notificationclick` handler — verify clicking the notification opens the register tab
- [ ] **JWT auth:** Login works but no token refresh — verify what happens when the JWT expires mid-session
- [ ] **PDF generation:** Works locally but Chrome crashes in Docker — verify `--shm-size` is set and memory limits exist
- [ ] **NLP integration:** API works but Portuguese accuracy is untested — verify with a set of real Portuguese emotional phrases
- [ ] **CouchDB security:** Database works but no `_security` documents or `require_valid_user` — verify CouchDB rejects unauthenticated reads
- [ ] **Docker Compose:** Starts up but no health checks — verify services wait for dependencies before starting
- [ ] **WhatsApp sharing:** Sends the message but PDF URL is publicly accessible — verify the download URL requires JWT auth
- [ ] **Offline form:** Entries save locally but no indicator if sync succeeded — verify per-entry sync status in the UI
- [ ] **Storage persistence:** App runs offline but IndexedDB can be evicted — verify `navigator.storage.persist()` was called

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Silently unsynced entries | MEDIUM | Force manual sync, compare CouchDB doc count with IndexedDB, identify missing docs via `_changes` feed |
| OOM from Chrome in Docker | HIGH | Add `--shm-size=2g`, implement concurrent request limit, restart container — verify PDF generation works before user retries |
| Lost entries from IndexedDB eviction | VERY HIGH | Impossible once brower deletes data. Mitigation: prevent with `persist()`, detect with sync confirmation, warn user |
| Expired JWT kills sync session | MEDIUM | Retry with refresh token, preserve unsaved entries in localStorage before sync attempt |
| Default Chrome notification shown | LOW | Fix promise chain in push event handler, test on real Android device |
| Portuguese emotion model fails to classify | MEDIUM | Switch model, retrain on Portuguese data, or lower confidence threshold |
| Wrong WhatsApp recipient receives report | HIGH | Only fixable with delivery confirmation system + user verification before send |
| NLP service timeout cascades to all requests | MEDIUM | Circuit breaker opens, async analysis enabled, cached results used |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| PouchDB conflict data loss | Phase 1 (Sync) | Unit test: create concurrent revisions, verify merge resolution |
| CouchDB auth leak | Phase 0 (Infra) | Integration: attempt unauthenticated CouchDB read, verify 401 |
| Live sync dies silently | Phase 1 (Sync) | E2E: simulate offline, add entry, reconnect, verify entry synced |
| Chrome OOM in Docker | Phase 3 (Reports) | Load test: request 3 concurrent PDFs, verify no OOM |
| Default push notification | Phase 4 (Push) | E2E: send push, verify custom notification, not Chrome default |
| JWT expiration kills session | Phase 1 (Auth) | Integration: use expired JWT, verify refresh + retry flow |
| NLP as hard dependency | Phase 3 (NLP) | Integration: kill NLP service, request PDF, verify graceful fallback |
| WhatsApp data leak | Phase 3 (Reports) | Security: verify PDF URL without JWT returns 401 |
| Docker Compose config drift | Phase 0 (Infra) | Review: `compose.yml` + `compose.dev.yml` + `compose.prod.yml` exist |
| User notification fatigue | Phase 4 (Push) | UX review: snooze, pause, smart suppression features implemented |
| Go HTTP client no timeout | Phase 3 (NLP) | Code review: all `http.Client` have explicit `Timeout` set |
| Notification click fails | Phase 4 (Push) | E2E: send push, click notification, verify register tab opens |
| PWA cache eviction | Phase 1 (PWA) | E2E: verify `navigator.storage.persist()` called at startup |
| Portuguese emotion model | Phase 3 (NLP) | Test: run 20 Portuguese phrases, verify >60% classification accuracy |

## Sources

- [PouchDB Conflict Resolution Guide](https://pouchdb.com/guides/conflicts.html) — HIGH confidence, official docs
- [PouchDB Replication Guide](https://pouchdb.com/guides/replication.html) — HIGH confidence, official docs
- [CouchDB Security Introduction](https://docs.couchdb.org/en/stable/intro/security.html) — HIGH confidence, official docs
- [CouchDB Authentication Database](https://docs.couchdb.org/en/stable/intro/security.html#authentication-database) — HIGH confidence, official docs
- [CouchDB Replication Protocol](https://docs.couchdb.org/en/stable/replication/protocol.html) — HIGH confidence, official docs
- [chromedp README (headless-shell Docker)](https://github.com/chromedp/chromedp) — HIGH confidence, official repo
- [chromedp/headless-shell Docker Hub](https://hub.docker.com/r/chromedp/headless-shell) — HIGH confidence, shows `--shm-size 2G` requirement
- [Web.dev Push Events Handling](https://web.dev/articles/push-notifications-handling-messages) — HIGH confidence, Google Chrome team documentation
- [Web.dev Web Push Protocol](https://web.dev/articles/push-notifications-web-push-protocol) — HIGH confidence, official spec documentation
- [Traefik ForwardAuth Middleware Docs](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/) — HIGH confidence, official docs
- [Docker Compose Production Guide](https://docs.docker.com/compose/production/) — HIGH confidence, official docs
- [MDN Progressive Web Apps](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps) — HIGH confidence, MDN docs
- [navigator.storage.persist() MDN](https://developer.mozilla.org/en-US/docs/Web/API/StorageManager/persist) — HIGH confidence, MDN docs
- Emotion tracking app UX pitfalls — MEDIUM confidence, synthesized from multiple sources (HCI research on mood tracking, therapy app UX reviews)

---
*Pitfalls research for: Kanso (offline-first PWA emotion diary)*
*Researched: 2026-05-16*
