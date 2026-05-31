# PWA — Prompt de Instalação Mobile

**Slug:** pwa-install-prompt
**Type:** quick
**Status:** approved
**Approved at:** 2026-05-31

## Tasks

1. **Generate PWA icons** — Create icon-192.png, icon-512.png, badge-72.png via Python/Pillow
2. **Hook: useInstallPrompt** — Capture `beforeinstallprompt`, expose `canInstall`, `triggerInstall()`, `userDismissed`
3. **Component: InstallBanner** — Banner "Instalar Kanso" / "Agora não" with animation
4. **Integrate in App.tsx** — Render InstallBanner inside layout
5. **Unify manifest** — Move icons to VitePWA config, remove public/manifest.json duplicate
6. **Tests** — Hook + component unit tests

## Behavior

- Captures `beforeinstallprompt` silently
- Shows banner when available
- "Instalar App" → triggers native `prompt()`
- "Agora não" → dismisses permanently via sessionStorage
- After install → banner hides
