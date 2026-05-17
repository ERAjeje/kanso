# Phase 6 — Push Notifications (Plan 02: PouchDB Sync)

## Goal

Migrate push preferences from direct API calls to offline-first PouchDB sync, matching the architecture of registros and sentimentos.

## Tasks

### Task 1: Add `preferenciasDB` to pouchdb.ts

**File**: `frontend/src/services/pouchdb.ts`
- Add `export const { local: preferenciasDB } = createSyncedDB('preferencias')`

### Task 2: Rewrite push.ts to use PouchDB

**File**: `frontend/src/services/push.ts`
- `savePreferences(prefs)` — `preferenciasDB.put()` with upsert pattern
- `getPreferences()` — `preferenciasDB.get()` with fallback defaults
- `saveSubscription(userId, fcmToken, timezone)` — saves to PouchDB

### Task 3: Update Profile.tsx

**File**: `frontend/src/pages/Profile.tsx`
- Remove API error handling (PouchDB save is local)
- Remove `saving` state since writes are instant

### Task 4: Update useAuth.tsx + usePushNotifications.ts

**File**: `frontend/src/hooks/useAuth.tsx`
- Save push subscription to PouchDB instead of API

**File**: `frontend/src/hooks/usePushNotifications.ts`
- Use `preferenciasDB` for subscription

### Task 5: Update backend to read from `preferencias` DB

**File**: `backend/internal/repository/couchdb.go`
- Change `GetPushPrefs`, `SavePushPrefs`, `GetAllPushPrefs` → target `preferencias` db

**File**: `backend/internal/handler/push.go`
- Remove `HandleGetPreferences`, `HandleUpdatePreferences`
- Update `HandleSubscribe` to use `preferencias` db

**File**: `backend/cmd/kanso-api/main.go`
- Remove get/update routes
- Add `preferencias` database creation on startup

### Task 6: Update scheduler

**File**: `scheduler/main.go`
- Change Mango query to target `preferencias` db

## Files Affected

| File | Action |
|------|--------|
| `frontend/src/services/pouchdb.ts` | Add `preferenciasDB` |
| `frontend/src/services/push.ts` | Rewrite to PouchDB |
| `frontend/src/pages/Profile.tsx` | Remove API calls |
| `frontend/src/hooks/useAuth.tsx` | Use PouchDB for subscribe |
| `frontend/src/hooks/usePushNotifications.ts` | Use PouchDB |
| `backend/internal/repository/couchdb.go` | Target `preferencias` db |
| `backend/internal/handler/push.go` | Remove get/update handlers |
| `backend/cmd/kanso-api/main.go` | Remove routes + create db |
| `scheduler/main.go` | Target `preferencias` db |
