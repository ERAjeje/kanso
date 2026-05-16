# Phase 2 — Pattern Map

> Code patterns and analogs for Phase 2 files, mapped from existing codebase.

---

## Component Patterns

### Existing: `AuthGuard.tsx`
- **Pattern:** Simple wrapper component using React Context for auth state
- **Key conventions:** Named export, Tailwind classes inline, no CSS modules, no styled-components
- **Analog for:** `AuthGuard` wrapping pattern → `TabBar` layout wrapper

### Existing: `SyncStatus.tsx`
- **Pattern:** State-driven UI via `useState` + `useEffect`, inline color map object
- **Key conventions:** Colors as object map, `flex items-center gap-2` layout, `rounded-full` for dots
- **Analog for:** Enhanced `SyncStatus` (same component, +PouchDB hook), `Toast` (state + conditional render)

---

## Page Patterns

### Existing: `Register.tsx`
- **Pattern:** Component with `useAuth` hook, contained within scrollable div, `p-8 max-w-lg mx-auto` page wrapper
- **Key conventions:** `h1` page title + optional subtitle at top, `bg-white rounded-xl p-6 shadow-sm border border-gray-100` card container
- **Analog for:** `History.tsx`, `Profile.tsx` (use same page wrapper pattern with "Em breve" content)

### Existing: `Login.tsx`
- **Pattern:** Centered card layout with brand header, uses `text-indigo-600` for brand accent
- **Key conventions:** `min-h-screen flex items-center justify-center` centering, `bg-indigo-50` page background with `bg-white` card

---

## Service Patterns

### Existing: `pouchdb.ts`
- **Pattern:** Singleton DB instances created at module load, `createSyncedDB()` factory function
- **Key conventions:** PouchDB local name = `kanso_{dbName}`, remote URL = `COUCHDB_URL/{dbName}`, JWT in fetch Authorization header via `authService.getStoredJWT()`
- **Analog for:** Enhanced `pouchdb.ts` (add `onSyncChange()`, `getUserId()`)

### Existing: `auth.ts`
- **Pattern:** Service module with exported functions, JWT stored in localStorage
- **Key conventions:** No classes — module-level functions, `authService` object, `authenticatedFetch` wrapper

---

## Hook Patterns

### Existing: `useAuth.tsx`
- **Pattern:** React Context + Provider pattern, exports context hook + provider component
- **Key conventions:** `createContext` + `useContext` + `useState`, provides `user`, `loading`, `signOut` to consumers
- **Analog for:** `usePouchSync.ts` (same hook pattern, simpler — no context, just `useState` + `useEffect`)

---

## App Structure Patterns

### Existing: `App.tsx`
- **Pattern:** React Router v7 routes, AuthProvider wrapper, AuthGuard on protected routes
- **Key conventions:** `<Routes>` at top level, `<Route path="/login">` unprotected, `<Route path="/register">` wrapped in `<AuthGuard>`, catch-all `*` → Login
- **Analog for:** TabBar layout restructure — `<Route element={<AuthGuard><TabBar/></AuthGuard>}>` with nested child routes

---

## TypeScript Types

### Existing: None (no `types/index.ts`)
- **Pattern to establish:** Shared interfaces for PouchDB document schemas
- **Key conventions from go backend:** `_id` + `_rev` pattern, snake_case JSON fields in Go (camelCase in TS), ISO 8601 dates

---

## UI Conventions

- Colors: gray-50 bg, indigo-600 accent, gray-800 text, gray-500 secondary text, gray-400 muted
- Borders: border-gray-200/300 for inputs, border-gray-100 for cards
- Focus: `focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500`
- Buttons: `bg-indigo-600 text-white rounded-lg py-3 font-medium hover:bg-indigo-700 disabled:opacity-50`
- Page wrapper: `p-8 max-w-lg mx-auto`
- Card: `bg-white rounded-xl p-6 shadow-sm border border-gray-100`
