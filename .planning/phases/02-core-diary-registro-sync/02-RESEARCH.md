# Phase 2: Core Diary — Registro & Sync - Research

**Researched:** 2026-05-16
**Domain:** Emotion registration form, offline-first PouchDB storage, cloud sync with CouchDB
**Confidence:** HIGH

## Summary

Phase 2 builds the core user interaction of Kanso: the emotion registration form. The architecture follows the established PouchDB↔CouchDB direct sync pattern — registrations save immediately to local PouchDB (works offline), and live replication syncs to CouchDB when connectivity is available. No Go backend changes are needed for CRUD operations; the backend only handles authentication, and Phase 2 integration is purely frontend work plus a new Go endpoint for sentimentos retrieval if needed.

The primary technical decisions are: (1) use `@headlessui/react` v2 Combobox for the sentimento autocomplete — it's the official Tailwind-compatible unstyled component with full accessibility, (2) use `lucide-react` v1.16 for tab bar icons (Pencil, Clock, User), (3) enhance the existing `pouchdb.ts` sync setup to expose sync event handlers for the `SyncStatus` component, and (4) wrap the app with a `TabBar` layout shell once user is authenticated.

**Primary recommendation:** Implement in 3 plans — (1) Combobox + Form core, (2) SyncStatus enhancement + Toast, (3) TabBar layout with route refactoring.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Sentimento Combobox
- **D-01:** Autocomplete — type to filter existing sentiments, press Enter/blur to create new
- **D-02:** New sentiments auto-saved to sentimentos DB on blur (immediately, not only on form submit)
- **D-03:** Sorted alphabetically
- **D-04:** Display name only (no usage count or metadata)

#### Form Layout & Flow
- **D-05:** Single scrollable form (not guided sections)
- **D-06:** Field order: date/time → sentimento → sensações → contexto → pensamentos
- **D-07:** Free-text fields (sensações, contexto, pensamentos) are multi-line textareas
- **D-08:** Form resets to empty after successful submission

#### Backdating UX
- **D-09:** Native HTML datetime-local input (no custom picker, no quick-select buttons)
- **D-10:** Defaults to current date/time on form open
- **D-11:** Precision: date + time (hours:minutes)

#### Offline & Sync Feedback
- **D-12:** Global sync status only (no per-registration status icons)
- **D-13:** When submitting offline: save to PouchDB silently + show toast "Saved locally — will sync when online"
- **D-14:** SyncStatus component enhanced to show actual sync state (online/syncing/offline) via PouchDB change/active events

#### Tab Navigation
- **D-15:** Bottom tab bar with 3 tabs: Registrar (active), Histórico (placeholder), Perfil (placeholder)
- **D-16:** Icon + text labels using lucide-react (Pencil, Clock, User icons)
- **D-17:** Inactive tabs show "Em breve" placeholder content

### the agent's Discretion
- Specific PouchDB document schema for registros (follow apontamentos.md model)
- CSS/styling details for form, tab bar, and components
- Form validation approach (required fields, min/max lengths)
- Error state UX for failed saves (though main path is offline-first)
- Choice of autocomplete combobox library or custom implementation
- Animation/transition details for tab switching

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| **REG-01** | User can register emotion with date/time (defaults to now), sensações (texto livre), sentimento (combobox customizável), contexto (texto livre), pensamentos (texto livre) | Headless UI Combobox + controlled form with textareas + PouchDB.put() on submit |
| **REG-02** | User can retroactively set date/time to a past moment | Native `<input type="datetime-local">` with no min/max restrictions |
| **REG-03** | Sentimento combobox lists previously saved sentiments and allows typing new ones | Query `sentimentosDB` via `allDocs()` for seed data; auto-save new on blur to same DB. Combobox filters locally |
| **SYNC-01** | Registrations save immediately to PouchDB (offline) | `registrosDB.put(doc)` called on form submit — synchronous within JS event loop, persists to IndexedDB |
| **SYNC-02** | PouchDB automatically syncs to CouchDB when connectivity is available | Already configured: `pouchdb.ts` creates live sync with `{live: true, retry: true}`. Phase 2 just writes documents |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Registration form rendering | Browser (React) | — | All form state is local; no server needed for render |
| Form validation | Browser (React) | — | Validated before PouchDB write; instant feedback |
| Sentimento combobox data | Browser (PouchDB) | — | `sentimentosDB` is local-first; syncs to CouchDB on background |
| PouchDB write (offline save) | Browser (PouchDB) | — | `put()` writes to IndexedDB; synchronous from user perspective |
| PouchDB → CouchDB sync | PouchDB replication engine | CouchDB | Already configured in `pouchdb.ts`; automatic with `live: true, retry: true` |
| Sync status indication | Browser (React) | PouchDB events | React component subscribes to PouchDB sync events |
| Tab bar layout | Browser (React) | — | Client-side routing with react-router-dom |
| Sentimento CRUD (backend) | — | — | **No Go backend needed for Phase 2.** PouchDB syncs sentimentos directly |
| Authentication (session) | Go backend (chi) | CouchDB JWT | Already built in Phase 1 — JWT in localStorage, sent in PouchDB fetch |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `@headlessui/react` | ^2.2.10 | Combobox autocomplete (sentimento field) | Official Tailwind-labs accessible, unstyled component; v2 supports Combobox with filtering and custom input. Perfect for D-01/D-03/D-04. [VERIFIED: npm registry] |
| `lucide-react` | ^1.16.0 | Tab bar icons (Pencil, Clock, User) | Lightweight icon library (tree-shakeable), works with React 19, most popular icon set for Tailwind projects. [VERIFIED: npm registry] |
| `pouchdb-browser` | ^9.0.0 | Local IndexedDB storage + CouchDB sync | Already installed in Phase 1. Core offline-first engine. [VERIFIED: npm registry] |
| `react-router-dom` | ^7.15.1 | Client-side routing + NavLink for tab bar | Already installed in Phase 1. NavLink provides `isActive` for tab highlighting. [VERIFIED: npm registry] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `date-fns` | ^4.1.0 | Date formatting for display and PouchDB queries | For formatting `dataHora` in the form display and sorting. Strongly recommended for timezone-safe operations. [VERIFIED: npm registry] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `@headlessui/react` Combobox | Custom combobox (input + dropdown) | More control but no accessibility guarantees, no keyboard navigation, more code to maintain. Headless UI v2 is the best choice for D-01/D-03. |
| `@headlessui/react` Combobox | `react-select` | react-select is heavier, harder to style with Tailwind, and has a different API. Headless UI is the Tailwind ecosystem's native choice. |
| `lucide-react` | `@heroicons/react` | Heroicons has fewer icons, specific to Tailwind design. lucide-react has the exact icons needed (Pencil, Clock, User) and is more widely used. |
| `date-fns` | Native `Intl.DateTimeFormat` | Native API works but date-fns provides cleaner formatting and manipulation APIs for `dataHora` display and query construction. |

**Installation:**
```bash
npm install @headlessui/react@^2.2.10 lucide-react@^1.16.0 date-fns@^4.1.0
```

## Architecture Patterns

### System Architecture Diagram

```
┌──────────────────────────────────────────────────────────┐
│                    Browser (PWA)                          │
│                                                           │
│  ┌────────────────────────────────────────────────┐       │
│  │  React (TabBarLayout)                          │       │
│  │  ┌─────────────┐  ┌──────────┐  ┌──────────┐  │       │
│  │  │  Registrar   │  │ Histórico│  │  Perfil   │  │       │
│  │  │  (active)    │  │(placeholder)│(placeholder)│  │       │
│  │  └──────┬──────┘  └──────────┘  └──────────┘  │       │
│  │         │                                       │       │
│  │  ┌──────▼──────────────────────────────────┐    │       │
│  │  │  RegistrationForm                        │    │       │
│  │  │  ┌─────────┐ ┌──────────────────────┐   │    │       │
│  │  │  │datetime │ │ Sentimento Combobox   │   │    │       │
│  │  │  │-local   │ │ (Headless UI v2)      │   │    │       │
│  │  │  └─────────┘ └──────────────────────┘   │    │       │
│  │  │  ┌────────────────────────────────────┐ │    │       │
│  │  │  │ Sensações (textarea)               │ │    │       │
│  │  │  │ Contexto (textarea)                │ │    │       │
│  │  │  │ Pensamentos (textarea)             │ │    │       │
│  │  │  └────────────────────────────────────┘ │    │       │
│  │  └─────────────────────────────────────────┘    │       │
│  └────────────────────────────────────────────────┘       │
│                    │                                       │
│  ┌─────────────────▼──────────────────────────────┐       │
│  │           PouchDB (IndexedDB)                   │       │
│  │  registrosDB ───── put(doc) ───→ local save      │       │
│  │  sentimentosDB ─── put(doc) ───→ local save      │       │
│  │  Sync: {live: true, retry: true}                  │       │
│  │  Events: change ▶ active ▶ paused ▶ error       │       │
│  └─────────────────┬──────────────────────────────┘       │
│                    │                                       │
│  ┌─────────────────▼──────────────────────────────┐       │
│  │  SyncStatus + ToastNotification                 │       │
│  │  (subscribes to PouchDB sync events)            │       │
│  └────────────────────────────────────────────────┘       │
└──────────────────────────┬───────────────────────────────┘
                           │ HTTPS + JWT in Authorization
                           ▼
┌──────────────────────────────────────────────────────────┐
│                    Traefik Gateway                         │
│  /db/* ──► CouchDB  (JWT validated natively)               │
│  /api/* ──► Go Backend (for auth only)                     │
└──────────────────────────────────────────────────────────┘
```

**Data flow (registration):**
1. User fills form → React state
2. Submit → `registrosDB.put(doc)` → localStorage immediately
3. PouchDB live sync picks up new doc → replicates to CouchDB
4. If offline: document stays in PouchDB; syncs when online (retry: true)
5. Sentimento: typed in Combobox → on blur → `sentimentosDB.put()` → single source of truth
6. SyncStatus: React subscribes to PouchDB `active`/`paused` events → shows online/syncing/offline

### Recommended Project Structure (Phase 2 additions)
```
frontend/src/
├── components/
│   ├── AuthGuard.tsx              # (existing)
│   ├── SyncStatus.tsx             # ENHANCE — add PouchDB sync event listeners
│   ├── Toast.tsx                  # NEW — simple toast component for offline saves
│   ├── TabBar.tsx                 # NEW — bottom tab bar layout + placeholder pages
│   ├── SentimentoCombobox.tsx     # NEW — Headless UI Combobox for sentimento field
│   └── RegistrationForm.tsx       # NEW — the full form component
├── hooks/
│   ├── useAuth.tsx                # (existing)
│   └── usePouchSync.ts            # NEW — hook exposing PouchDB sync state
├── services/
│   ├── auth.ts                    # (existing)
│   ├── pouchdb.ts                 # ENHANCE — expose sync handler for event subscription
│   └── registros.ts               # NEW — CRUD functions for registros (put, query)
├── pages/
│   ├── Login.tsx                  # (existing)
│   ├── Register.tsx               # REPLACE — integrate RegistrationForm + SentimentoCombobox
│   ├── History.tsx                # NEW — placeholder ("Em breve")
│   └── Profile.tsx                # NEW — placeholder ("Em breve")
├── App.tsx                        # MODIFY — add TabBar layout wrapper for authenticated routes
├── main.tsx                       # (existing)
└── types/
    └── index.ts                   # NEW — TypeScript types for registro and sentimento docs
```

### Pattern 1: PouchDB Document Schema (Immutable Append-Only)
**What:** Each emotion registration is written once with a UUID `_id` and never updated. This avoids MVCC conflicts entirely.

**Document schemas:**

```typescript
// frontend/src/types/index.ts

/** Schema for a registro document saved to registros PouchDB/CouchDB */
export interface RegistroDoc {
  _id: string                    // crypto.randomUUID()
  type: 'registro'
  userId: string                 // from JWT sub claim
  dataHora: string               // ISO 8601: "2026-05-16T14:30:00"
  sensacoes: string              // free text
  sentimentoId: string | null    // UUID from sentimentos DB, or null for new
  sentimentoNome: string         // denormalized for display
  contexto: string               // free text
  pensamentos: string            // free text
  createdAt: string              // ISO 8601: when doc was created in PouchDB
  updatedAt: string              // ISO 8601: same as createdAt for entry (immutable)
}

/** Schema for a sentimento document saved to sentimentos PouchDB/CouchDB */
export interface SentimentoDoc {
  _id: string                    // crypto.randomUUID()
  type: 'sentimento'
  userId: string                 // from JWT sub claim
  nome: string                   // the emotion name, normalized (lowercased, trimmed)
  criadoEm: string               // ISO 8601
}
```

**Why schema is defined in frontend/types:** The document is created by the frontend (PouchDB), synced to CouchDB, and optionally read by the Go backend (for reports). TypeScript types ensure consistency.

### Pattern 2: PouchDB CRUD Service
**What:** Encapsulated functions for saving registros and querying sentimentos, used by React components.

```typescript
// frontend/src/services/registros.ts
import { registrosDB, sentimentosDB, getUserId } from './pouchdb'
import type { RegistroDoc, SentimentoDoc } from '../types'

export async function saveRegistro(data: Omit<RegistroDoc, '_id' | 'type' | 'userId' | 'createdAt' | 'updatedAt'>): Promise<RegistroDoc> {
  const now = new Date().toISOString()
  const doc: RegistroDoc = {
    _id: crypto.randomUUID(),
    type: 'registro',
    userId: getUserId(),
    ...data,
    createdAt: now,
    updatedAt: now,
  }
  await registrosDB.put(doc)
  return doc
}

export async function saveSentimento(nome: string): Promise<SentimentoDoc> {
  const doc: SentimentoDoc = {
    _id: crypto.randomUUID(),
    type: 'sentimento',
    userId: getUserId(),
    nome: nome.trim().toLowerCase(),
    criadoEm: new Date().toISOString(),
  }
  await sentimentosDB.put(doc)
  return doc
}

/** Fetch all sentiments for the current user, sorted alphabetically */
export async function getSentimentos(): Promise<SentimentoDoc[]> {
  const result = await sentimentosDB.allDocs<SentimentoDoc>({ include_docs: true })
  return result.rows
    .map(row => row.doc!)
    .filter(doc => doc.type === 'sentimento')
    .sort((a, b) => a.nome.localeCompare(b.nome, 'pt-BR'))
}
```

### Pattern 3: PouchDB Sync Events for SyncStatus
**What:** Expose the sync handler from `pouchdb.ts` so React components can subscribe to sync state changes.

**Why this pattern:** The existing `pouchdb.ts` creates sync but doesn't expose the handler. To implement D-12/D-14, React needs access to the `Replication` object's events.

```typescript
// frontend/src/services/pouchdb.ts — ENHANCED
import PouchDB from 'pouchdb-browser'
import { authService } from './auth'
import type { SyncState } from '../types'

const COUCHDB_URL = 'https://kanso.local/db'

// Type for sync state listeners
type SyncListener = (state: SyncState) => void
const syncListeners = new Set<SyncListener>()

export function onSyncChange(listener: SyncListener): () => void {
  syncListeners.add(listener)
  return () => syncListeners.delete(listener)
}

function notifySyncState(state: SyncState) {
  syncListeners.forEach(fn => fn(state))
}

function createSyncedDB(dbName: string): { local: PouchDB.Database; remote: PouchDB.Database } {
  const local = new PouchDB(`kanso_${dbName}`)
  const remote = new PouchDB(`${COUCHDB_URL}/${dbName}`, {
    fetch: (url, opts = {}) => {
      const jwt = authService.getStoredJWT()
      const headers = new Headers(opts.headers)
      if (jwt) headers.set('Authorization', `Bearer ${jwt}`)
      return PouchDB.fetch(url, { ...opts, headers })
    },
  })

  local.sync(remote, { live: true, retry: true })
    .on('change', () => notifySyncState('syncing'))
    .on('paused', () => notifySyncState(navigator.onLine ? 'online' : 'offline'))
    .on('active', () => notifySyncState('syncing'))
    .on('error', (err) => {
      console.error(`PouchDB sync error (${dbName}):`, err)
      notifySyncState('offline')
    })

  return { local, remote }
}

// Export both DB and the remote for potential cancellation
export const { local: registrosDB, remote: registrosRemote } = createSyncedDB('registros')
export const { local: sentimentosDB, remote: sentimentosRemote } = createSyncedDB('sentimentos')

// Get userId from stored JWT (for writing to documents)
export function getUserId(): string {
  const jwt = authService.getStoredJWT()
  if (!jwt) return ''
  try {
    const payload = JSON.parse(atob(jwt.split('.')[1]))
    return payload.sub || ''
  } catch {
    return ''
  }
}
```

**Important design decision:** Both `registros` and `sentimentos` DB instances emit sync events. To avoid duplicate state updates (two events per sync cycle), the `SyncStatus` component should listen to only one DB's events (e.g., `registros`), or the `notifySyncState` function should debounce rapid state transitions. The recommendation is: listen to one DB and accept slightly imprecise state for the other. Since both DBs sync to the same CouchDB instance, their `paused`/`active` transitions are correlated.

### Pattern 4: usePouchSync Hook
**What:** React hook that wraps the PouchDB sync event subscription, providing a clean `SyncState` to components.

```typescript
// frontend/src/hooks/usePouchSync.ts
import { useState, useEffect } from 'react'
import { onSyncChange } from '../services/pouchdb'
import type { SyncState } from '../types'

export function usePouchSync(): SyncState {
  const [state, setState] = useState<SyncState>(
    navigator.onLine ? 'online' : 'offline'
  )

  useEffect(() => {
    const unsub = onSyncChange(setState)
    return unsub
  }, [])

  return state
}
```

### Pattern 5: Enhanced SyncStatus Component
**What:** Consumes `usePouchSync` hook instead of relying solely on browser `online`/`offline` events.

```typescript
// frontend/src/components/SyncStatus.tsx — ENHANCED
import { SyncIcon, WifiIcon, WifiOffIcon, LoaderIcon } from 'lucide-react' // or simple colored dots
import { usePouchSync } from '../hooks/usePouchSync'

export function SyncStatus() {
  const syncState = usePouchSync()
  
  const config = {
    online: { color: 'bg-green-500', label: 'Sincronizado', icon: WifiIcon },
    syncing: { color: 'bg-yellow-500', label: 'Sincronizando...', icon: LoaderIcon },
    offline: { color: 'bg-red-500', label: 'Offline', icon: WifiOffIcon },
    // Could add 'error' state
  } as const

  const { color, label, icon: Icon } = config[syncState]

  return (
    <div className="flex items-center gap-1.5 text-xs text-gray-400">
      <span className={`w-2 h-2 rounded-full ${color}`} />
      <span>{label}</span>
    </div>
  )
}
```

### Pattern 6: SentimentoCombobox with Headless UI v2
**What:** Autocomplete combobox that queries local `sentimentosDB` for suggestions and auto-saves new sentiments on blur.

**When to use:** Only for the sentimento field on the registration form (D-01 through D-04).

```tsx
// frontend/src/components/SentimentoCombobox.tsx
import { useState, useEffect, useRef, type KeyboardEvent } from 'react'
import { Combobox } from '@headlessui/react'
import { getSentimentos, saveSentimento } from '../services/registros'
import type { SentimentoDoc } from '../types'
import { ChevronsUpDown } from 'lucide-react' // or use ▼

interface Props {
  value: string
  onChange: (nome: string, id: string | null) => void
}

export function SentimentoCombobox({ value, onChange }: Props) {
  const [query, setQuery] = useState('')
  const [sentimentos, setSentimentos] = useState<SentimentoDoc[]>([])
  const inputRef = useRef<HTMLInputElement>(null)
  const [selected, setSelected] = useState<string | null>(value || null)

  // Load sentiments on mount
  useEffect(() => {
    getSentimentos().then(setSentimentos)
  }, [])

  const filtered = query === ''
    ? sentimentos
    : sentimentos.filter(s =>
        s.nome.toLowerCase().includes(query.toLowerCase())
      )

  const handleBlur = async () => {
    const trimmed = query.trim()
    if (!trimmed) return
    
    // Check if it already exists
    const exists = sentimentos.some(s => s.nome.toLowerCase() === trimmed.toLowerCase())
    if (!exists && trimmed.length > 0) {
      const newDoc = await saveSentimento(trimmed)
      setSentimentos(prev => [...prev, newDoc].sort((a, b) => a.nome.localeCompare(b.nome, 'pt-BR')))
      onChange(newDoc.nome, newDoc._id)
      setSelected(newDoc.nome)
    }
  }

  const handleSelect = (nome: string | null) => {
    if (!nome) return
    const sent = sentimentos.find(s => s.nome === nome)
    onChange(nome, sent?._id || null)
    setSelected(nome)
    setQuery('')
  }

  // Allow pressing Enter to create if no selection made
  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && query.trim() && filtered.length === 0) {
      handleBlur()
    }
  }

  return (
    <Combobox value={selected} onChange={handleSelect} onClose={() => setQuery('')}>
      <div className="relative">
        <Combobox.Input
          ref={inputRef}
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm
                     focus:outline-none focus:ring-2 focus:ring-indigo-500"
          displayValue={(nome: string | null) => nome ?? ''}
          onChange={(e) => setQuery(e.target.value)}
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
          placeholder="Digite ou selecione um sentimento..."
          autoComplete="off"
        />
        <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-3">
          <ChevronsUpDown className="h-4 w-4 text-gray-400" />
        </Combobox.Button>
      </div>

      <Combobox.Options
        className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-lg
                   border border-gray-200 bg-white shadow-lg text-sm"
      >
        {filtered.length === 0 && query !== '' ? (
          <div className="px-4 py-2 text-gray-500">
            Pressione Enter para criar "{query}"
          </div>
        ) : (
          filtered.map((sentimento) => (
            <Combobox.Option
              key={sentimento._id}
              value={sentimento.nome}
              className={({ active }) =>
                `cursor-pointer px-4 py-2 ${active ? 'bg-indigo-50 text-indigo-700' : 'text-gray-900'}`
              }
            >
              {sentimento.nome}
            </Combobox.Option>
          ))
        )}
      </Combobox.Options>
    </Combobox>
  )
}
```

### Pattern 7: TabBar Layout with NavLink
**What:** Bottom-anchored tab bar using `react-router-dom` NavLink with icons from `lucide-react`.

**Rationale:** React Router v7 `NavLink` provides `isActive` for styling. The tab bar is rendered as a fixed-bottom `nav` element with flexbox. Active tab uses `text-indigo-600`, inactive uses `text-gray-400`.

```tsx
// frontend/src/components/TabBar.tsx
import { NavLink, Outlet } from 'react-router-dom'
import { Pencil, Clock, User } from 'lucide-react'

const tabs = [
  { path: '/register', icon: Pencil, label: 'Registrar' },
  { path: '/history', icon: Clock, label: 'Histórico' },
  { path: '/profile', icon: User, label: 'Perfil' },
]

export function TabBar() {
  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-1 overflow-y-auto pb-16">
        <Outlet />
      </main>
      
      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-50">
        <div className="flex justify-around items-center h-16 max-w-lg mx-auto">
          {tabs.map(({ path, icon: Icon, label }) => (
            <NavLink
              key={path}
              to={path}
              className={({ isActive }) =>
                `flex flex-col items-center gap-0.5 px-4 py-2 text-xs font-medium transition-colors
                ${isActive ? 'text-indigo-600' : 'text-gray-400 hover:text-gray-600'}`
              }
            >
              <Icon className="h-5 w-5" />
              <span>{label}</span>
            </NavLink>
          ))}
        </div>
      </nav>
    </div>
  )
}
```

The `App.tsx` routes are restructured to use the TabBar layout for authenticated routes:

```tsx
// App.tsx — MODIFIED
import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './hooks/useAuth'
import { AuthGuard } from './components/AuthGuard'
import { TabBar } from './components/TabBar'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { History } from './pages/History'
import { Profile } from './pages/Profile'

function App() {
  return (
    <AuthProvider>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={
            <AuthGuard>
              <TabBar />
            </AuthGuard>
          }>
            <Route path="/register" element={<Register />} />
            <Route path="/history" element={<History />} />
            <Route path="/profile" element={<Profile />} />
          </Route>
          <Route path="*" element={<Login />} />
        </Routes>
      </div>
    </AuthProvider>
  )
}
```

**Key routing change:** Authenticated routes now use a **layout route** pattern. The `<AuthGuard>` wraps `<TabBar />`, which uses `<Outlet />` to render child routes. This means the tab bar is persistent across tab navigation with zero re-renders.

### Anti-Patterns to Avoid
- **Go backend as CRUD proxy:** Do NOT create Go endpoints for creating/fetching registros. PouchDB syncs directly to CouchDB. The Go backend only handles auth and future side-effects (reports, NLP).
- **Sequential document IDs:** Never use auto-incrementing IDs for PouchDB documents. Always use `crypto.randomUUID()` to prevent sync conflicts when multiple devices write offline.
- **Sentimento pre-population from backend:** Do NOT create a Go API to fetch the list of sentiments. The `sentimentosDB` syncs from CouchDB — users' sentiments come from local PouchDB, which seeded itself from the user's previous sessions synced via CouchDB.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Autocomplete combobox | Custom input + dropdown + keyboard navigation + accessibility | `@headlessui/react` v2 Combobox | Accessibility (ARIA), keyboard navigation (arrow keys, escape, blur), portal support, clean Tailwind integration. Writing this from scratch with proper a11y is ~500 lines. |
| Tab bar icons | SVG icons manually | `lucide-react` | Pencil, Clock, User — exact icons needed, tree-shakeable, 24px default, consistent stroke design. |
| Date/time picker | Custom date picker | Native `<input type="datetime-local">` | D-09 explicitly mandates native input. Native pickers on mobile (Chrome Android, iOS Safari) are optimized for touch. No library needed. |
| Sync state machine | Custom event system for PouchDB sync | `usePouchSync` hook wrapping `onSyncChange` subscription | PouchDB's replication object already emits all events needed. The hook is ~15 lines with proper cleanup. |
| Toast notifications | Full toast library (react-hot-toast, sonner) | Simple `<Toast>` component + CSS animation | D-13 only needs one type of toast ("Saved locally..."). A full library is overkill for a single-purpose toast. ~30 lines of component code. |

**Key insight:** The offline-first PouchDB pattern already handles most of the hard problems (live sync, auto-retry, conflict resolution). Phase 2 benefits from what's already built in Phase 1. The form is "just" a React form that calls `PouchDB.put()`. The complexity is in the UX details: combobox behavior, sync state indication, and tab navigation.

## Common Pitfalls

### Pitfall 1: Sync Event Storm (Rapid active/paused transitions)
**What goes wrong:** The `paused` and `active` events fire rapidly during normal sync operation (every batch of changes). The UI flickers between "syncing" and "online" constantly.
**Why it happens:** PouchDB fires `active` when starting a batch, then `paused` when the batch completes. With live sync, this happens every few seconds as the replication catches up.
**How to avoid:** Debounce the state transitions. Only transition from `online` to `syncing` if the sync stays `active` for >200ms. Only transition from `syncing` to `online` if `paused` fires without an `active` following within 1 second. Use a simple `useRef` timer.
**Warning signs:** The sync status indicator flashes "syncing...online...syncing...online..."

### Pitfall 2: Duplicate Sentimento Entries
**What goes wrong:** The `onBlur` handler fires when the user clicks a dropdown option (because focus moves from input to option), creating a duplicate sentiment.
**Why it happens:** Headless UI Combobox closes on option select, which causes `blur` on the input. The `onBlur` handler sees a non-empty query and saves it.
**How to avoid:** Check if the blur resulted from selecting an existing option (i.e., if `selected` was set). Use a flag or compare the query with the selected value. The pattern: in `handleSelect`, set a flag `justSelected = true`, and in `handleBlur`, skip save if `justSelected` is true (clear it after a short delay).
**Warning signs:** Duplicate entries in `sentimentosDB` with the same `nome` but different `_id`.

### Pitfall 3: Form Data Lost on Navigation or Refresh
**What goes wrong:** User types a long entry, navigates to another tab accidentally, and all form data is lost.
**Why it happens:** React component unmounts on route change, losing all local state. React Router's layout route with `<Outlet />` preserves the tab content.
**How to avoid:** The TabBar layout pattern (Pattern 7) uses nested routes with `<Outlet />` — this means the Register component stays mounted when switching between tabs only if you use a state-based tab panel approach instead of routes. For the route-based approach, consider whether data preservation across tab switches matters (D-05 says the form resets on submit, not on tab switch). If needed, persist form state to `sessionStorage` on beforeunload to prevent data loss on accidental browser close. For MVP, accept that tab navigation resets the form (consistent with D-08's reset-on-submit principle).
**Warning signs:** Users complain about losing partially filled forms when switching tabs.

### Pitfall 4: DateTime-Local Input Timezone Confusion
**What goes wrong:** The `datetime-local` input returns a local time string without timezone info. When the user backdates to a past moment, the timezone offset at that past date may differ from the current offset (DST).
**Why it happens:** `datetime-local` gives you `"2026-05-16T14:30"` — no timezone. If stored as-is, display assumes the user's current timezone, which may differ.
**How to avoid:** Always convert the `datetime-local` value to an ISO 8601 string with explicit offset or UTC before saving. Use `date-fns` `format()` with `z` token, or append the user's timezone offset at save time. For MVP, storing local time as-is and displaying it back works for a single-user app in a single timezone.
**Warning signs:** Times shift by an hour when DST changes.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Headless UI v1 (manual state management) | Headless UI v2 (simplified Combobox API with `displayValue`, `immediate`) | 2024-2025 | v2 Combobox has cleaner v-model pattern, built-in nullable support, and `ComboboxOptions anchor` prop |
| `react-router-dom` v6 | `react-router-dom` v7 | 2025 | v7 has layout routes (`<Route element={<Layout/>}>` pattern), stable NavLink API. Our pattern follows v7 conventions |
| Tailwind v3 (config file required) | Tailwind v4 (CSS-first, `@import "tailwindcss"`) | 2025 | No config file needed. Class names same. Compatible with Headless UI v2 |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `crypto.randomUUID()` is available in all modern browsers the PWA targets | SentimentoCombobox / RegistrationForm | Low — `crypto.randomUUID()` is available in Chrome 63+, Firefox 95+, Safari 15.4+. Fallback: simple UUID v4 generator. |
| A2 | PouchDB `allDocs()` returns documents sorted by `_id` (which can be used for sorting) | PouchDB CRUD Service | Low — we sort by `nome` locally anyway. `allDocs()` default sort is by `_id` but we use explicit localeCompare. |
| A3 | The `sentimentosDB` starts empty for new users and populates as they type sentiments | SentimentoCombobox | Medium — first-time user sees an empty combobox with no suggestions. UX is still valid (type a sentiment, press Enter). Consider seeding with 3-5 common sentiment examples as a nice-to-have. |
| A4 | Go backend does not need changes for Phase 2 | Backend Integration | Medium — verified via CONTEXT.md and apontamentos.md: PouchDB syncs directly with CouchDB. Backend only handles auth. If CouchDB `_security` or `validate_doc_update` rejects incoming docs from PouchDB sync, might need backend fix — but Phase 1 should have configured this. |

## Code Examples

### RegistrationForm — Full Pattern

```tsx
// frontend/src/components/RegistrationForm.tsx
import { useState, type FormEvent } from 'react'
import { usePouchSync } from '../hooks/usePouchSync'
import { SentimentoCombobox } from './SentimentoCombobox'
import { saveRegistro } from '../services/registros'

interface Props {
  onSaved: () => void
  onShowToast: (message: string) => void
}

export function RegistrationForm({ onSaved, onShowToast }: Props) {
  const syncState = usePouchSync()
  const [dataHora, setDataHora] = useState(() => {
    const now = new Date()
    // Format as YYYY-MM-DDTHH:MM for datetime-local default
    return now.toISOString().slice(0, 16)
  })
  const [sentimento, setSentimento] = useState({ nome: '', id: null as string | null })
  const [sensacoes, setSensacoes] = useState('')
  const [contexto, setContexto] = useState('')
  const [pensamentos, setPensamentos] = useState('')
  const [saving, setSaving] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (saving) return
    setSaving(true)

    try {
      await saveRegistro({
        dataHora: new Date(dataHora).toISOString(),
        sensacoes,
        sentimentoId: sentimento.id,
        sentimentoNome: sentimento.nome,
        contexto,
        pensamentos,
      })

      // Reset form (D-08)
      setDataHora(new Date().toISOString().slice(0, 16))
      setSentimento({ nome: '', id: null })
      setSensacoes('')
      setContexto('')
      setPensamentos('')

      // Show offline toast if needed (D-13)
      if (syncState !== 'online') {
        onShowToast('Salvo localmente — será sincronizado quando você estiver online')
      }

      onSaved()
    } catch (err) {
      console.error('Failed to save:', err)
      // Error is unlikely since PouchDB.put() saves locally
      // but handle network errors if DB initialization failed
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Date/Time — D-09, D-10, D-11 */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Data e Hora
        </label>
        <input
          type="datetime-local"
          value={dataHora}
          onChange={(e) => setDataHora(e.target.value)}
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm
                     focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      {/* Sentimento — D-01, D-02, D-03, D-04 */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Sentimento
        </label>
        <SentimentoCombobox
          value={sentimento.nome}
          onChange={(nome, id) => setSentimento({ nome, id })}
        />
      </div>

      {/* Sensações — D-07 */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Sensações
        </label>
        <textarea
          value={sensacoes}
          onChange={(e) => setSensacoes(e.target.value)}
          rows={3}
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm
                     focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          placeholder="O que você está sentindo no corpo?"
        />
      </div>

      {/* Contexto — D-07 */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Contexto
        </label>
        <textarea
          value={contexto}
          onChange={(e) => setContexto(e.target.value)}
          rows={3}
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm
                     focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          placeholder="O que estava acontecendo?"
        />
      </div>

      {/* Pensamentos — D-07 */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Pensamentos
        </label>
        <textarea
          value={pensamentos}
          onChange={(e) => setPensamentos(e.target.value)}
          rows={3}
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm
                     focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          placeholder="O que está passando pela sua cabeça?"
        />
      </div>

      <button
        type="submit"
        disabled={saving || !sentimento.nome.trim()}
        className="w-full bg-indigo-600 text-white rounded-lg py-3 font-medium
                   hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed
                   transition-colors"
      >
        {saving ? 'Salvando...' : 'Registrar'}
      </button>
    </form>
  )
}
```

## Backend Integration Analysis

### What Backend Work Does Phase 2 Need?

**Answer: None, for the basic MVP flow.**

The PouchDB-to-CouchDB direct sync architecture means:
- `registrosDB.put(doc)` → syncs to CouchDB automatically
- `sentimentosDB.put(doc)` → syncs to CouchDB automatically
- The existing Traefik+CouchDB JWT auth from Phase 1 handles authentication

**If the Go backend needs to query sentimentos (for future report generation):**
The existing `repository/couchdb.go` pattern can be extended:

```go
// backend/internal/repository/couchdb.go — potential addition for future use
type SentimentoDoc struct {
    ID     string `json:"_id,omitempty"`
    Rev    string `json:"_rev,omitempty"`
    Type   string `json:"type"`
    UserID string `json:"userId"`
    Nome   string `json:"nome"`
}

func (c *CouchDB) GetSentimentosByUser(userID string) ([]SentimentoDoc, error) {
    // Query to sentimentos DB using Mango or allDocs
    url := fmt.Sprintf("%s/sentimentos/_all_docs?include_docs=true&startkey=%q&endkey=%q",
        c.baseURL, "_", "_")
    // ... existing HTTP pattern
}
```

**NOT needed in Phase 2:** No new Go endpoint is required. The only Phase 2 Go change (if at all) would be ensuring `validate_doc_update` functions on `registros` and `sentimentos` databases accept the PouchDB writes. This was set up in Phase 1.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest (aligned with Vite 8 ecosystem) |
| Config file | Run in existing Vite config or `vitest.config.ts` |
| Quick run command | `npx vitest run --reporter=verbose` |
| Full suite command | Same — single-suite for frontend |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REG-01 | Form renders all 5 fields (datetime, sentimento, sensações, contexto, pensamentos) | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "renders all fields"` | ❌ Wave 0 |
| REG-01 | Submit calls registrosDB.put with correct document shape | unit (mock DB) | `npx vitest run src/components/RegistrationForm.test.tsx -t "calls saveRegistro on submit"` | ❌ Wave 0 |
| REG-01 | Form resets after successful save | unit | Same file, separate test | ❌ Wave 0 |
| REG-02 | datetime-local input allows past dates (no min/max restriction) | unit | `npx vitest run src/components/RegistrationForm.test.tsx -t "datetime-local has no min/max"` | ❌ Wave 0 |
| REG-03 | Combobox shows all stored sentiments | unit (mock sentimentosDB) | `npx vitest run src/components/SentimentoCombobox.test.tsx -t "shows all sentiments"` | ❌ Wave 0 |
| REG-03 | Typing filters the list | unit | Same file | ❌ Wave 0 |
| REG-03 | Blur with new text creates a sentiment | unit | Same file | ❌ Wave 0 |
| SYNC-01 | saveRegistro stores doc locally via PouchDB.put | integration | `npx vitest run src/services/registros.test.ts -t "saves to PouchDB"` | ❌ Wave 0 |
| SYNC-02 | Sync is configured with live:true, retry:true | unit | `npx vitest run src/services/pouchdb.test.ts -t "creates live sync with retry"` | ❌ Wave 0 |
| D-15/D-16 | Tab bar renders 3 tabs with correct icons | unit | `npx vitest run src/components/TabBar.test.tsx -t "renders 3 tabs"` | ❌ Wave 0 |
| D-17 | Inactive tabs show "Em breve" | unit | Same file | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `npx vitest run --reporter=verbose --changed`
- **Per wave merge:** Full suite run
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `vitest` + `@testing-library/react` — install with `npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom`
- [ ] `src/components/RegistrationForm.test.tsx` — form rendering + submit + reset behavior
- [ ] `src/components/SentimentoCombobox.test.tsx` — filtering, create on blur, loading sentiments
- [ ] `src/components/TabBar.test.tsx` — 3 tabs, active state, placeholder content
- [ ] `src/services/registros.test.ts` — PouchDB put and query
- [ ] `src/services/pouchdb.test.ts` — sync configuration verification
- [ ] `src/hooks/usePouchSync.test.ts` — hook subscription and cleanup
- [ ] `src/hooks/__mocks__/pouchdb.ts` — mock PouchDB for tests (or use `pouchdb-memory` adapter)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Already covered in Phase 1 | Google OAuth + JWT |
| V5 Input Validation | Yes | Zod schema validation before PouchDB write |
| V6 Cryptography | No | No new crypto needed (JWT handled in Phase 1) |

### Known Threat Patterns for React + PouchDB

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS in textarea fields | Tampering | React's default JSX escaping prevents XSS. Input is rendered through React, not `dangerouslySetInnerHTML`. Ensure all display of registro data uses React text rendering. |
| IndexedDB data on compromised device | Information Disclosure | No built-in browser encryption. Rely on device-level encryption (full-disk encryption on mobile). Emotional diary data is sensitive — document this limitation. |

**Validation approach:** Use Zod for runtime validation of form data before writing to PouchDB:

```typescript
// frontend/src/services/validation.ts
import { z } from 'zod'

export const registroSchema = z.object({
  dataHora: z.string().datetime(),
  sensacoes: z.string().max(2000, "Máximo de 2000 caracteres"),
  sentimentoNome: z.string().min(1, "Sentimento é obrigatório").max(100),
  contexto: z.string().max(2000),
  pensamentos: z.string().max(2000),
})
// Mark: sentimentoId is nullable — it's null for brand new sentiments
```

Zod is the recommended validation library (lightweight, TypeScript-first, tree-shakeable). Install with `npm install zod`.

## File Modification Plan

| File | Action | Reason |
|------|--------|--------|
| `frontend/src/services/pouchdb.ts` | **ENHANCE** | Expose `onSyncChange()` for sync event subscription, export `getUserId()` helper |
| `frontend/src/types/index.ts` | **CREATE** | TypeScript types for `RegistroDoc`, `SentimentoDoc`, `SyncState` |
| `frontend/src/services/registros.ts` | **CREATE** | `saveRegistro()`, `saveSentimento()`, `getSentimentos()` functions |
| `frontend/src/services/validation.ts` | **CREATE** | Zod schemas for form validation |
| `frontend/src/hooks/usePouchSync.ts` | **CREATE** | React hook wrapping `onSyncChange()` |
| `frontend/src/components/SentimentoCombobox.tsx` | **CREATE** | Headless UI Combobox for sentimento field |
| `frontend/src/components/RegistrationForm.tsx` | **CREATE** | Full registration form with all 5 fields |
| `frontend/src/components/SyncStatus.tsx` | **MODIFY** | Replace browser event listener with `usePouchSync` hook |
| `frontend/src/components/Toast.tsx` | **CREATE** | Simple toast notification component |
| `frontend/src/components/TabBar.tsx` | **CREATE** | Bottom tab bar layout component |
| `frontend/src/pages/Register.tsx` | **REWRITE** | Replace placeholder with RegistrationForm integration |
| `frontend/src/pages/History.tsx` | **CREATE** | Placeholder "Em breve" page |
| `frontend/src/pages/Profile.tsx` | **CREATE** | Placeholder "Em breve" page |
| `frontend/src/App.tsx` | **MODIFY** | Add layout route with TabBar, add History/Profile routes |
| `frontend/package.json` | **MODIFY** | Add dependencies: `@headlessui/react`, `lucide-react`, `date-fns`, `zod`, `vitest`, `@testing-library/react` |
| `backend/...` | **NO CHANGES** | No Go backend changes needed for Phase 2 |

## Plan Splitting Recommendation

**Structure: 3 plans across 2 waves**

### Wave 1 (Foundation — independent, no dependencies between plans)

**Plan 2-1: Infrastructure and Data Layer**
- Files: `services/pouchdb.ts` (enhance), `types/index.ts` (create), `services/registros.ts` (create), `services/validation.ts` (create), `hooks/usePouchSync.ts` (create)
- Deliverables: TypeScript types, PouchDB sync events exposed, CRUD service functions, validation schemas, sync state hook
- Tests: Unit tests for registros service, validation schemas
- Dependencies: None within Phase 2

**Plan 2-2: Sentimento Combobox and Registration Form**
- Files: `components/SentimentoCombobox.tsx` (create), `components/RegistrationForm.tsx` (create), `components/Toast.tsx` (create)
- Deliverables: Fully functional registration form with all 5 fields, combobox autocomplete, auto-save on blur, form validation, toast for offline saves
- Tests: Combobox filtering + create, form render + submit + reset
- Dependencies: Plan 2-1 (types + services)

### Wave 2 (depends on form being complete)

**Plan 2-3: TabBar Layout and SyncStatus Enhancement**
- Files: `components/TabBar.tsx` (create), `pages/History.tsx` (create), `pages/Profile.tsx` (create), `pages/Register.tsx` (rewrite), `components/SyncStatus.tsx` (modify), `App.tsx` (modify)
- Deliverables: Bottom tab bar navigation, placeholder tabs, enhanced sync status indicator, route restructuring with layout routes
- Tests: TabBar rendering + navigation, SyncStatus event subscription
- Dependencies: Plan 2-1, Plan 2-2

**Rationale for plan split:**
- Plan 2-1 has zero UI dependencies — pure data layer work
- Plan 2-2 is the core deliverable that can be developed and tested independently
- Plan 2-3 is the "shell" that ties everything together — needs the form to be complete before the UX flow makes sense
- Wave 2 starts immediately after Wave 1 plans complete (they're not blocked waiting)

## Sources

### Primary (HIGH confidence)
- `/apache/pouchdb` — PouchDB sync events API (`change`, `paused`, `active`, `denied`, `error`) [VERIFIED: Context7]
- `/tailwindlabs/headlessui` v2 — Headless UI Combobox API, displayValue, onChange, onClose [VERIFIED: Context7]
- `/remix-run/react-router` v7 — NavLink isActive, layout routes, Outlet [VERIFIED: Context7]
- `npm view lucide-react version` → 1.16.0 [VERIFIED: npm registry]
- `npm view @headlessui/react version` → 2.2.10 [VERIFIED: npm registry]
- `npm view react-router-dom version` → 7.15.1 (existing dependency) [VERIFIED: npm registry]
- `npm view pouchdb-browser version` → 9.0.0 (existing dependency) [VERIFIED: npm registry]
- `apontamentos.md` — Document schemas for registros and sentimentos [CITED: project file]
- `02-CONTEXT.md` — All D-01 through D-17 decisions [CITED: project file]
- `frontend/src/services/pouchdb.ts` — Existing PouchDB sync pattern [CITED: source code]

### Secondary (MEDIUM confidence)
- Tailwind v4 + Headless UI v2 compatibility — Both from Tailwind Labs, designed to work together [ASSUMED]

### Tertiary (LOW confidence)
- Empty sentimentos DB first-time UX — No user testing data available [ASSUMED]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all version-verified via npm registry and Context7
- Architecture: HIGH — follows established PouchDB↔CouchDB patterns, existing code patterns
- Pitfalls: MEDIUM — sync event storm needs runtime verification; timezone handling depends on actual usage

**Research date:** 2026-05-16
**Valid until:** 2026-06-16 (stable libraries; Headless UI v2 and react-router-dom v7 are stable releases)
