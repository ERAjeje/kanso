---
status: resolved
trigger: "Push notification permission not requested on login; switch toggle silent failure"
created: 2026-05-17
updated: 2026-05-17
---

## Symptoms

- **Expected**: On login, notification permission should be requested OR show feedback if VAPID key is missing
- **Actual**: Nothing happens — silent return when VITE_VAPID_PUBLIC_KEY is not set
- **Expected**: Toggling push notification switch should show visual feedback (success/error)
- **Actual**: Switch does nothing when API calls fail — silent catch block
- **Error messages**: None shown to user (caught silently)
- **Reproduction**: Login without VITE_VAPID_PUBLIC_KEY set; toggle switch without backend running

## Diagnosis

### Root Cause 1: Silent VAPID key check
`frontend/src/hooks/useAuth.tsx` — after sign-in, tries to register push:
```typescript
const vapidKey = import.meta.env.VITE_VAPID_PUBLIC_KEY
if (!vapidKey) return  // silent — no feedback
```

### Root Cause 2: Silent catch on switch toggle
`frontend/src/pages/Profile.tsx` — `toggleEnabled` and `updateTime`:
```typescript
try {
  await updatePreferences(next)
  setPrefs(prev => prev ? { ...prev, enabled: next.enabled } : null)
} catch {
  // revert on error — but no visual feedback
}
```

## Fix Plan

1. **Add Toast component** — create a simple toast system for user feedback
2. **Show feedback when VAPID key missing** — toast message: "Notificações push não configuradas — chave VAPID ausente"
3. **Show error on switch toggle failure** — toast message when API call fails
4. **Show success on switch toggle** — brief confirmation on successful toggle

## Files to modify

- `frontend/src/components/Toast.tsx` — already exists (check current implementation)
- `frontend/src/hooks/useAuth.tsx` — add toast when vapid key missing
- `frontend/src/pages/Profile.tsx` — add toast feedback for toggle + update

## Evidence

- [x] useAuth.tsx line ~50-60: silent return when !vapidKey
- [x] Profile.tsx toggleEnabled: silent catch block
- [x] Profile.tsx updateTime: silent catch block

## Resolution

**Root cause confirmed:** Missing visual feedback on silent failures.

**Fix applied:**
1. `frontend/src/hooks/useAuth.tsx` — Toast quando VAPID key ausente: "Notificações push não configuradas"
2. `frontend/src/pages/Profile.tsx` — Toast de erro em `toggleEnabled` e `updateTime` quando API falha
3. `frontend/src/pages/Profile.tsx` — Toast de sucesso quando preferências salvas

**Files changed:** `useAuth.tsx`, `Profile.tsx`

**Verification:** `npm run test` — 61 tests pass
**Approved:** 2026-05-17
**Executed:** 2026-05-17
