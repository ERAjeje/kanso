---
phase: fix
plan_id: 01
wave: 1
depends_on: []
files_modified:
  - frontend/index.html
  - frontend/src/pages/Login.tsx
  - frontend/src/pages/Login.test.tsx
autonomous: false
requirements:
  - "BUG-001: Botão Google Sign-In não renderiza porque script GIS não está no index.html"
  - "BUG-002: Nenhum fallback visível se o Google Identity Services falhar"
---

# Plan 01: Fix Google Sign-In button not rendering

## Objective

Fix the Login page so the Google Sign-In button actually renders and is clickable.

**Root cause:** `Login.tsx` depends on `window.google.accounts.id` (Google Identity Services) but the required script `https://accounts.google.com/gsi/client` is not loaded in `index.html`. The `<div id="google-signin-btn">` renders empty — the user sees only decorative text with no actionable button.

## Tasks

### Task 1: Add GIS script to index.html and add fallback button

**type**: execute
**wave**: 1
**files_modified**:
  - frontend/index.html
  - frontend/src/pages/Login.tsx
  - frontend/src/pages/Login.test.tsx

<read_first>
- frontend/index.html
- frontend/src/pages/Login.tsx
</read_first>

<acceptance_criteria>
- `frontend/index.html` includes `<script src="https://accounts.google.com/gsi/client" async defer>` in `<head>`
- `Login.tsx` renders a styled fallback "Entrar com Google" button if `window.google` is not available after timeout
- Fallback button calls `window.google.accounts.id.prompt()` or redirects to Google OAuth URL
- No TypeScript errors (tsc --noEmit passes)
- Tests pass (pnpm test)
</acceptance_criteria>

<action>
1. Add GIS script to `frontend/index.html`:
   - Insert `<script src="https://accounts.google.com/gsi/client" async defer></script>` before `</head>`

2. Update `frontend/src/pages/Login.tsx`:
   - Add a fallback timer: if `window.google` is not available after 5 seconds, show a custom styled button
   - The fallback button mimics Google Sign-In styling (white background, Google logo SVG, "Entrar com Google")
   - Fallback button onClick calls `window.google?.accounts.id.prompt()` or opens Google OAuth directly
   - Keep existing GIS `renderButton` logic as primary path
   - Add a link "Criar conta" that navigates to `/register` (the diary registration is the next step after login)

3. Create `frontend/src/pages/Login.test.tsx`:
   - Test that the page renders without crashing
   - Test that the prompt text is displayed
</action>

---

## Verification

1. `cd frontend && npx tsc --noEmit` passes
2. `cd frontend && pnpm test` passes
3. Login page renders with visible Google Sign-In button (or styled fallback)
4. Clicking button triggers Google OAuth flow
