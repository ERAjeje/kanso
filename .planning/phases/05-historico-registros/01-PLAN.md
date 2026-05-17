# Phase 5 — Histórico de Registros (Plan 01)

## Goal

Users can browse their emotional history chronologically and view full registration details inline.

## Tasks

### Task 1: Add `getRegistros()` to registros service

- **File**: `frontend/src/services/registros.ts`
- Add `getRegistros()` using `registrosDB.allDocs<RegistroDoc>({ include_docs: true })`
- Filter by `type === 'registro'`
- Sort by `dataHora` descending (most recent first)
- Export the function

### Task 2: Create `RegistroCard` component

- **File**: `frontend/src/components/RegistroCard.tsx`
- Props: `registro: RegistroDoc`
- Card shows:
  - **Header**: `sentimentoNome` if non-empty string; otherwise "Buscando sentimento" (italic, text-gray-400)
  - **Subtitle**: Formatted date/time from `dataHora` (pt-BR locale)
  - **Preview**: First ~80 chars of non-empty field (sensacoes || contexto || pensamentos), truncated with ellipsis
- Inline expand on click:
  - Shows all fields: Sensações, Sentimento, Contexto, Pensamentos
  - Each field with its label and full content
  - Toggle chevron/icon indicating collapsed/expanded state
- Styling consistent with Register page: `bg-white rounded-xl p-6 shadow-sm border border-gray-100`
- Transition/animation for expand/collapse

### Task 3: Rewrite `History.tsx` page

- **File**: `frontend/src/pages/History.tsx`
- Import `getRegistros` and `RegistroCard`
- States: `registros`, `loading`, `error`
- `useEffect` on mount to fetch all registros
- Layout: `p-8 max-w-lg mx-auto` (matching Register page)
- Header row with `SyncStatus` component (matching Register)
- **Loading state**: Skeleton/spinner
- **Empty state**: Friendly message "Nenhum registro ainda" with illustration/icon
- **Error state**: Error message with retry button
- **Data state**: List of `RegistroCard` components

### Task 4: Handle `sentimentoNome` fallback

- When `sentimentoNome` is empty string or undefined, display fallback text
- Fallback: "Buscando sentimento" — italic, text-gray-400
- This communicates the therapeutic process: the sentiment is still being discovered

### Task 5: Tests

- `frontend/src/components/RegistroCard.test.tsx`
  - Renders sentimentoNome when present
  - Renders fallback text when sentimentoNome is empty/null
  - Shows date/time formatted correctly
  - Shows content preview truncated
  - Expands inline on click
  - Collapses on second click
- `frontend/src/pages/History.test.tsx`
  - Mocks `getRegistros` from registros service
  - Renders loading state initially
  - Renders list of cards after data loads
  - Renders empty state when no registros
  - Renders error state on failure

### Task 6: Update STATE.md

- Mark Phase 5 as in progress
- Update session continuity

## Files Affected

| File | Action |
|------|--------|
| `frontend/src/services/registros.ts` | Add `getRegistros()` |
| `frontend/src/pages/History.tsx` | Rewrite from placeholder |
| `frontend/src/components/RegistroCard.tsx` | **New** |
| `frontend/src/components/RegistroCard.test.tsx` | **New** |
| `frontend/src/pages/History.test.tsx` | **New** |
| `.planning/STATE.md` | Update |

## Dependencies

- None (Phase 2 already provides PouchDB setup and RegistroDoc type)

## Verification

- `npm run test` — all existing + new tests pass
- `npm run lint` — no lint errors
- Manual: visual check of history page in browser
