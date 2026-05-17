# Phase 6 — Push Notifications (Plan 01)

## Goal

Users receive push notification reminders at 12h, 18h, e 23h (defaults, configurable) to record their emotions. Tapping a notification opens the registration screen with current date/time.

## Architecture

```
┌──────────────┐   POST /api/push/subscribe   ┌──────────────┐
│  PWA/React   │ ────────────────────────────▶ │  Go Backend  │
│  + SW        │   {fcmToken, timezone}        │  (kanso-api) │
│              │ ◀──────────────────────────── │              │
│  Notificação │      FCM push message         │  pushSvc     │
│              │                               │  (FCM SDK)   │
└──────────────┘                               └──────┬───────┘
                                                      │ POST /api/push/send
                                                      │ (internal)
                                             ┌────────▼───────┐
                                             │  Scheduler      │
                                             │  (kanso-sched)  │
                                             │  Go ticker loop │
                                             │                 │
                                             │  Lê push_prefs  │
                                             │  do CouchDB     │
                                             └─────────────────┘
```

## Tasks

### Task 1: Add push preference model and repository methods

**File**: `backend/internal/repository/couchdb.go`

Add types and methods for push preferences:

```go
type PushPrefsDoc struct {
    ID        string   `json:"_id,omitempty"`
    Rev       string   `json:"_rev,omitempty"`
    Type      string   `json:"type"`
    UserSub   string   `json:"userSub"`
    Enabled   bool     `json:"enabled"`
    Times     []string `json:"times"`
    Timezone  string   `json:"timezone"`
    FCMToken  string   `json:"fcmToken"`
    CreatedAt string   `json:"createdAt,omitempty"`
    UpdatedAt string   `json:"updatedAt,omitempty"`
}
```

Methods:
- `GetPushPrefs(sub string) (*PushPrefsDoc, error)` — GET `push_prefs:{sub}` from `usuarios` DB
- `SavePushPrefs(doc *PushPrefsDoc) error` — PUT to `usuarios/push_prefs:{sub}`
- `GetAllPushPrefs() ([]PushPrefsDoc, error)` — Mango query `{type: "push_prefs", enabled: true}` on `usuarios` DB (for scheduler)

<read_first>
- `backend/internal/repository/couchdb.go` — existing patterns (UserDoc, putDoc, mangoQuery)
</read_first>

<acceptance_criteria>
- `PushPrefsDoc` struct defined with all 9 fields
- `GetPushPrefs(sub)` returns `*PushPrefsDoc` or `nil` (not found)
- `SavePushPrefs(doc)` uses `c.putDoc("usuarios", doc.ID, doc)`
- `GetAllPushPrefs()` returns only docs with `type: "push_prefs"` and `enabled: true`
- All methods follow existing error handling patterns
</acceptance_criteria>

### Task 2: Add FCM config and push service

**File**: `backend/internal/config/config.go`
Add field: `FCMServerKey string` (env: `FCM_SERVER_KEY`, default: `""`)

**File**: `backend/internal/service/push.go` (new)
Create `PushService`:

```go
type PushService struct {
    couchRepo     *repository.CouchDB
    fcmServerKey  string
    httpClient    *http.Client
}

func NewPushService(couchRepo *repository.CouchDB, fcmServerKey string) *PushService
```

Methods:
- `Subscribe(sub, fcmToken, timezone string) error` — Saves/updates push preferences document
- `GetPreferences(sub string) (*repository.PushPrefsDoc, error)` — Delegates to repo
- `UpdatePreferences(sub string, enabled bool, times []string) error` — Updates prefs
- `Send(sub string) error` — Sends FCM push via Firebase HTTP API:
  - POST to `https://fcm.googleapis.com/fcm/send`
  - Header: `Authorization: key={fcmServerKey}`
  - Header: `Content-Type: application/json`
  - Body: `{"to": "{fcmToken}", "notification": {"title": "Kanso", "body": "Como você está se sentindo agora?", "click_action": "/register"}}`
  - Reads user's FCM token from CouchDB prefs

Use Firebase Legacy HTTP API (simpler than Admin SDK, same as recommended in ARCHITECTURE.md).

<read_first>
- `backend/internal/repository/couchdb.go` — PushPrefsDoc, existing patterns
- `backend/internal/service/report.go` — service struct pattern (repo dependency)
- `backend/internal/config/config.go` — env config pattern
</read_first>

<acceptance_criteria>
- `config.go` has `FCMServerKey string` field with env var reading
- `push.go` has `PushService` with `Subscribe`, `GetPreferences`, `UpdatePreferences`, `Send`
- `Send` makes HTTP POST to FCM API with correct headers and body payload
- `Send` returns error if FCM token not found for user
- `Subscribe` auto-sets `createdAt` on new, `updatedAt` always
</acceptance_criteria>

### Task 3: Create push handler and routes

**File**: `backend/internal/handler/push.go` (new)

Create `PushHandler`:

```go
type PushHandler struct {
    pushSvc *service.PushService
}
```

Endpoints:
- `HandleSubscribe` — POST `/api/push/subscribe`
  - Body: `{fcmToken: string, timezone: string}`
  - Calls `pushSvc.Subscribe(sub, fcmToken, timezone)`
  - Returns `200 {"status": "ok"}`
- `HandleGetPreferences` — GET `/api/push/preferences`
  - Returns push prefs document (or default: `{enabled: true, times: ["12:00", "18:00", "23:00"]}`)
- `HandleUpdatePreferences` — PUT `/api/push/preferences`
  - Body: `{enabled: bool, times: string[]}`
  - Calls `pushSvc.UpdatePreferences(sub, enabled, times)`
- `HandleSend` — POST `/api/push/send` (internal — also protected by JWT)
  - Body: `{userId: string}` (sent by scheduler)
  - Calls `pushSvc.Send(userId)`
  - Returns `200 {"status": "ok"}`

**File**: `backend/cmd/kanso-api/main.go`
- Add imports for new handler/service
- Instantiate `pushSvc` and `pushHandler`
- Add routes under protected group:
  ```go
  r.Post("/api/push/subscribe", pushHandler.HandleSubscribe)
  r.Get("/api/push/preferences", pushHandler.HandleGetPreferences)
  r.Put("/api/push/preferences", pushHandler.HandleUpdatePreferences)
  r.Post("/api/push/send", pushHandler.HandleSend)
  ```

<read_first>
- `backend/internal/handler/report.go` — handler pattern (struct, constructor, JWT from context)
- `backend/cmd/kanso-api/main.go` — route registration pattern
</read_first>

<acceptance_criteria>
- `push.go` handler has all 4 methods
- `main.go` instantiates `pushSvc` and `pushHandler` and registers all 4 routes under JWT group
- Endpoints use `Middleware.UserContextKey` for user identification
- `POST /api/push/send` uses body param `userId` (not JWT claims — scheduler calls it)
</acceptance_criteria>

### Task 4: Frontend service worker + push notification hook

**File**: `frontend/public/sw.js` (new — service worker)

```javascript
self.addEventListener('push', event => {
  const data = event.data?.json() ?? { title: 'Kanso', body: 'Como você está se sentindo agora?' }
  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body,
      icon: '/icon-192.png',
      badge: '/badge-72.png',
      vibrate: [200, 100, 200],
      data: { url: '/register' }
    })
  )
})

self.addEventListener('notificationclick', event => {
  event.notification.close()
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then(clientList => {
      if (clientList.length > 0) {
        clientList[0].focus()
        clientList[0].navigate(event.notification.data?.url || '/register')
      } else {
        clients.openWindow(event.notification.data?.url || '/register')
      }
    })
  )
})
```

**File**: `frontend/src/services/push.ts` (new)

```typescript
import { authService } from './auth'

const API = import.meta.env.VITE_API_URL || ''

export interface PushPreferences {
  enabled: boolean
  times: string[]
  timezone: string
}

export async function subscribe(fcmToken: string): Promise<void> {
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const token = authService.getJWT()
  await fetch(`${API}/api/push/subscribe`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ fcmToken, timezone })
  })
}

export async function getPreferences(): Promise<PushPreferences> {
  const token = authService.getJWT()
  const res = await fetch(`${API}/api/push/preferences`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  return res.json()
}

export async function updatePreferences(prefs: { enabled: boolean; times: string[] }): Promise<void> {
  const token = authService.getJWT()
  await fetch(`${API}/api/push/preferences`, {
    method: 'PUT',
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify(prefs)
  })
}
```

**File**: `frontend/src/hooks/usePushNotifications.ts` (new)

```typescript
import { useEffect, useState } from 'react'
import { useAuth } from './useAuth'
import { subscribe } from '../services/push'

export function usePushNotifications() {
  const { user } = useAuth()
  const [permission, setPermission] = useState<NotificationPermission | 'unavailable'>('default')
  const [subscribed, setSubscribed] = useState(false)

  useEffect(() => {
    if (!user) return
    if (!('Notification' in window)) {
      setPermission('unavailable')
      return
    }
    setPermission(Notification.permission)
  }, [user])

  const requestPermission = async () => {
    if (permission === 'unavailable') return
    const result = await Notification.requestPermission()
    setPermission(result)

    if (result === 'granted') {
      // Register service worker and subscribe
      const reg = await navigator.serviceWorker.register('/sw.js')
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(import.meta.env.VITE_VAPID_PUBLIC_KEY)
      })
      // Send subscription to backend
      const token = sub.toJSON().endpoint // Simplified — in production use FCM
      await subscribe(token)
      setSubscribed(true)
    }
  }

  return { permission, subscribed, requestPermission }
}

// Utility: VAPID key converter
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - base64String.length % 4) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; i++) outputArray[i] = rawData.charCodeAt(i)
  return outputArray
}
```

**File**: `frontend/index.html` — Register service worker on app load
Add to `<head>`:
```html
<link rel="manifest" href="/manifest.json" />
```

**File**: `frontend/public/manifest.json` (create or update)
```json
{
  "name": "Kanso",
  "short_name": "Kanso",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#4f46e5",
  "icons": [
    { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

<read_first>
- `frontend/src/services/push.ts` — will reference auth service for JWT
- `frontend/src/hooks/useAuth.tsx` — user state pattern
- `frontend/src/services/auth.ts` — need `getJWT()` method
</read_first>

<acceptance_criteria>
- `sw.js` handles `push` event with `showNotification` and `notificationclick` with navigation to `/register`
- `push.ts` exports `subscribe`, `getPreferences`, `updatePreferences` — all use Authorization header
- `usePushNotifications.ts` exports `permission`, `subscribed`, `requestPermission`
- `manifest.json` exists with `start_url`, `display: standalone`, icons
- Service worker registered on app load
</acceptance_criteria>

### Task 5: Integrate permission prompt on login

**File**: `frontend/src/hooks/useAuth.tsx`
- After successful sign-in (in `signIn`), trigger notification permission request
- Import `usePushNotifications` or call `requestPermission` directly

**Simpler approach**: Add permission request call after `setUser(u)` in `signIn`:

```typescript
// After auth success, prompt for notification permission
if ('Notification' in window && Notification.permission === 'default') {
  requestNotificationPermission()
}
```

**File**: `frontend/src/main.tsx` — Register service worker on startup

```typescript
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js')
  })
}
```

<read_first>
- `frontend/src/hooks/useAuth.tsx` — signIn flow
- `frontend/src/main.tsx` — app entry point
</read_first>

<acceptance_criteria>
- After Google sign-in completes, `Notification.requestPermission()` is called if permission is `default`
- Service worker `/sw.js` is registered on app load
- No errors if notification API is unavailable
</acceptance_criteria>

### Task 6: Create notification settings UI on Profile page

**File**: `frontend/src/pages/Profile.tsx`
- Add "Lembretes" section below existing content
- Toggle switch for enabled/disabled
- Time inputs (3 fields, default 12:00, 18:00, 23:00)
- Uses `getPreferences` and `updatePreferences` from push service
- Show permission status: "Permitido" / "Negado" / "Não solicitado"
- If denied, show instruction to enable in browser settings

**File**: `frontend/src/pages/Profile.test.tsx` (new)
- Renders notification settings section
- Toggle calls `updatePreferences`
- Shows correct permission state

<read_first>
- `frontend/src/pages/Profile.tsx` — current profile page layout
- `frontend/src/services/push.ts` — API functions
- `frontend/src/components/ReportSection.tsx` — example of section within Profile
</read_first>

<acceptance_criteria>
- Profile page has "Lembretes" section with toggle and 3 time inputs
- Toggle calls `updatePreferences({enabled, times})`
- Time inputs default to "12:00", "18:00", "23:00"
- Shows current notification permission state
- Tests render the settings section and verify toggle interaction
</acceptance_criteria>

### Task 7: Create scheduler microservice

**File**: `scheduler/main.go` (new)

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

type config struct {
    couchDBURL  string
    couchDBUser string
    couchDBPass string
    apiURL      string // URL of kanso-api (e.g., http://api:8080)
    interval    time.Duration
}

type pushPrefsDoc struct {
    ID      string   `json:"_id"`
    UserSub string   `json:"userSub"`
    Enabled bool     `json:"enabled"`
    Times   []string `json:"times"`
    Timezone string  `json:"timezone"`
}

type mangoQuery struct {
    Selector map[string]interface{} `json:"selector"`
}

type mangoResponse struct {
    Docs []json.RawMessage `json:"docs"`
}
```

Logic:
1. Every `interval` (default 1 minute), query CouchDB for all `type: push_prefs, enabled: true`
2. For each user, check if current time (in their timezone) matches any configured time
3. If match found, POST to `{apiURL}/api/push/send` with `{"userId": "{userSub}"}`
4. Log successes and errors (no retry — next tick will handle missed)

**File**: `.env` — Add `FCM_SERVER_KEY` entry

**File**: `infra/docker-compose.yml` — Add scheduler service:
```yaml
scheduler:
  build:
    context: ../scheduler
    dockerfile: Dockerfile
  container_name: kanso-scheduler
  env_file: ../.env
  environment:
    - COUCHDB_URL=http://couchdb:5984
    - COUCHDB_USER=admin
    - API_URL=http://api:8080
  depends_on:
    couchdb:
      condition: service_healthy
    api:
      condition: service_started
  restart: unless-stopped
```

**File**: `scheduler/Dockerfile` (new)
```dockerfile
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /kanso-scheduler .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /kanso-scheduler /usr/local/bin/
CMD ["/usr/local/bin/kanso-scheduler"]
```

**File**: `scheduler/go.mod` (new)
```
module github.com/edson/kanso-scheduler

go 1.26.2
```

<read_first>
- `infra/docker-compose.yml` — existing service definitions
- `backend/Dockerfile` — multi-stage build pattern
</read_first>

<acceptance_criteria>
- `scheduler/main.go` exists with ticker loop checking every 1 minute
- Reads `push_prefs` docs from CouchDB with Mango query `{type: "push_prefs", enabled: true}`
- Matches current time (in user's timezone) against user's configured times
- Calls `POST {apiURL}/api/push/send` with correct body for matching users
- Dockerfile uses multi-stage build
- docker-compose has `scheduler` service with env vars and depends_on
- `.env` has `FCM_SERVER_KEY`
</acceptance_criteria>

### Task 8: Tests

**File**: `backend/internal/service/push_test.go` (new)
- `Subscribe` creates/updates push prefs document
- `Send` makes HTTP POST to FCM API (mock HTTP client)
- `Send` returns error when no FCM token exists
- `GetPreferences` returns prefs from repo

**File**: `backend/internal/handler/push_test.go` (new)
- `HandleSubscribe` returns 200 on valid request
- `HandleGetPreferences` returns default prefs
- `HandleUpdatePreferences` updates and returns 200

**File**: `frontend/src/services/push.test.ts` (new)
- `subscribe` sends POST with fcmToken and timezone
- `getPreferences` fetches and returns prefs
- `updatePreferences` sends PUT with enabled and times

**File**: `frontend/src/hooks/usePushNotifications.test.ts` (new)
- Returns correct permission state
- `requestPermission` registers service worker and subscribes

**File**: `frontend/src/pages/Profile.test.tsx` — Update or create to cover notification settings

<read_first>
- `backend/internal/service/report_test.go` — test patterns
- `backend/internal/handler/report_test.go` — handler test patterns
- `frontend/src/services/registros.test.ts` — frontend service test patterns
</read_first>

<acceptance_criteria>
- All test files compile and pass
- `npm run test` passes
- Backend tests cover subscribe, send, get/update preferences flows
- Frontend tests cover API calls and permission hook
</acceptance_criteria>

### Task 9: Update planning docs

**File**: `.planning/STATE.md`
- Mark Phase 6 as in progress
- Update session continuity

**File**: `.planning/ROADMAP.md`
- Update Phase 6 status to "In Progress"

<read_first>
- `.planning/STATE.md` — current state
- `.planning/ROADMAP.md` — roadmap format
</read_first>

<acceptance_criteria>
- STATE.md updated with Phase 6 status
- ROADMAP.md reflects current status
- Session continuity updated
</acceptance_criteria>

## Files Affected

| File | Action |
|------|--------|
| `backend/internal/repository/couchdb.go` | Add `PushPrefsDoc` + CRUD methods |
| `backend/internal/config/config.go` | Add `FCMServerKey` |
| `backend/internal/service/push.go` | **New** — PushService |
| `backend/internal/service/push_test.go` | **New** — service tests |
| `backend/internal/handler/push.go` | **New** — PushHandler |
| `backend/internal/handler/push_test.go` | **New** — handler tests |
| `backend/cmd/kanso-api/main.go` | Add push routes |
| `frontend/public/sw.js` | **New** — service worker |
| `frontend/public/manifest.json` | **New** — PWA manifest |
| `frontend/src/services/push.ts` | **New** — push API client |
| `frontend/src/services/push.test.ts` | **New** — push service tests |
| `frontend/src/hooks/usePushNotifications.ts` | **New** — permission/token hook |
| `frontend/src/hooks/usePushNotifications.test.ts` | **New** — hook tests |
| `frontend/src/hooks/useAuth.tsx` | Add permission prompt after login |
| `frontend/src/main.tsx` | Register service worker on load |
| `frontend/src/pages/Profile.tsx` | Add notification settings section |
| `frontend/src/pages/Profile.test.tsx` | Add notification settings tests |
| `scheduler/main.go` | **New** — scheduler microservice |
| `scheduler/go.mod` | **New** |
| `scheduler/Dockerfile` | **New** |
| `infra/docker-compose.yml` | Add scheduler service |
| `.env` | Add `FCM_SERVER_KEY` |
| `.planning/STATE.md` | Update |
| `.planning/ROADMAP.md` | Update |

## Dependencies

- Phase 1 (auth) — JWT middleware, user identification
- Phase 2 (registro) — registration screen to open on notification tap
- FCM server key from Firebase Console

## Verification

- `go build ./...` — backend compiles
- `npm run test` — all existing + new tests pass
- `npm run lint` — no lint errors
- `docker compose config` — docker-compose valid
- Manual: push notification appears on device and tap opens /register
