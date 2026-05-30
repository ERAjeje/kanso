# Plan: fix-deleted-client-oauth

## Description
Corrigir erro `401: deleted_client` no login Google — o `frontend/.env` continha um Client ID deletado. Implementar fonte única no `.env` raiz via `envDir` do Vite.

## Tasks
1. `frontend/vite.config.ts` — Adicionar `envDir: path.resolve(__dirname, '..')`
2. Remover `frontend/.env` (não será mais necessário)
3. Verificar que `.env` raiz tem `VITE_GOOGLE_CLIENT_ID` correto

## Files Affected
- `frontend/vite.config.ts`
- `frontend/.env` (delete)
