---
status: complete
executed_at: 2026-05-31
---

# PWA Install Prompt — Summary

## What was done

1. **Ícones PWA** — Gerados com Pillow (fundo indigo-600, letra "K"): `icon-192.png`, `icon-512.png`, `badge-72.png`
2. **Hook `useInstallPrompt`** — Captura `beforeinstallprompt`, expõe `canInstall`, `triggerInstall()`, `userDismissed()`. Persiste dispensa em `sessionStorage`.
3. **Componente `InstallBanner`** — Banner "Instale o Kanso para uma experiência melhor" com botões "Instalar App" / "Agora não". Posicionado no rodapé (`bottom-20`).
4. **Integração em `App.tsx`** — Banner renderizado em todas as rotas autenticadas.
5. **Manifest unificado** — Ícones adicionados ao `VitePWA` config, `public/manifest.json` removido, `index.html` link do manifest limpo. `injectRegister: false` para preservar `sw.js` manual de push.

## Tests

- `useInstallPrompt.test.ts` — 9 tests (event capture, preventDefault, trigger, accept, dismiss, sessionStorage check)
- `InstallBanner.test.tsx` — 4 tests (hidden when false, rendered when true, install click, dismiss click)
- All 82 existing tests pass ✅
- TypeScript compiles clean ✅

## Files changed/created

| File | Action |
|------|--------|
| `frontend/public/icon-192.png` | created |
| `frontend/public/icon-512.png` | created |
| `frontend/public/badge-72.png` | created |
| `frontend/src/hooks/useInstallPrompt.ts` | created |
| `frontend/src/hooks/useInstallPrompt.test.ts` | created |
| `frontend/src/components/InstallBanner.tsx` | created |
| `frontend/src/components/InstallBanner.test.tsx` | created |
| `frontend/src/App.tsx` | edited |
| `frontend/vite.config.ts` | edited |
| `frontend/index.html` | edited |
| `frontend/public/manifest.json` | removed |
