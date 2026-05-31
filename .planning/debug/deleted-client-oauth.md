---
status: resolved
trigger: "Login Google retorna 401 deleted_client"
created: 2026-05-30
updated: 2026-05-30
resolved: 2026-05-30
---

# Debug Session: deleted-client-oauth

## Symptoms

- **Expected behavior**: User clicks "Entrar com Google" → redirected to Google OAuth → redirected back to app authenticated
- **Actual behavior**: Browser shows error page "Acesso bloqueado: erro de autorização. The OAuth client was deleted. Erro 401: deleted_client"
- **Error messages**: `401: deleted_client` — "The OAuth client was deleted"
- **Timeline**: Já funcionou antes. Usuário recriou as credenciais no Google Cloud Console.
- **Reproduction**: Acessar frontend → clicar "Entrar com Google" → erro 401

## Current Focus

- **hypothesis**: O Client ID do Google OAuth foi deletado do Google Cloud Console quando as credenciais foram recriadas, mas os arquivos `.env` ainda contêm o Client ID antigo.
- **test**: Verificar Client IDs atuais nos arquivos `.env` e comparar com o que existe no Google Cloud Console.
- **expecting**: Os Client IDs em `.env` e `frontend/.env` não existem mais no GCP.
- **next_action**: Obter novo Client ID do GCP e atualizar os arquivos `.env`

## Evidence

- `.env` root: `GOOGLE_CLIENT_ID=721313208894-8tk11j8838outa181pbsj9d5fppqku0m.apps.googleusercontent.com`
- `frontend/.env`: `VITE_GOOGLE_CLIENT_ID=721313208894-66etlcvv0gp2kin2a33vuqkona4hp675.apps.googleusercontent.com` — **DELETADO**
- Os dois valores são DIFERENTES — inconsistência entre backend e frontend
- `.env` root também tem `VITE_GOOGLE_CLIENT_ID=721313208894-8tk11j8838outa181pbsj9d5fppqku0m.apps.googleusercontent.com`
- Usuário recriou credenciais no GCP Console; o Client ID `...66etlcvv` foi deletado
- O novo Client ID do GCP é `...8tk11j8` (confirmado pelo usuário)

## Eliminated

- *(nenhuma ainda)*

## Resolution

- **root_cause**: O arquivo `frontend/.env` continha um Client ID (`...66etlcvv`) diferente do arquivo `.env` raiz (`...8tk11j8`). Esse Client ID foi deletado do Google Cloud Console quando as credenciais foram recriadas, causando o erro `401: deleted_client` no login via Google.
- **fix**: Remover `frontend/.env` e configurar Vite para ler `.env` da raiz via `envDir` — fonte única de verdade.
- **verification**: 69/69 testes passam. Vite carrega vars do `.env` raiz.
- **files_changed**: `frontend/vite.config.ts` (+1 linha), `frontend/.env` (removido)
