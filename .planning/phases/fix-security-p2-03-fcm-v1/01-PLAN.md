---
phase: fix-security-p2-03-fcm-v1
plan: 01
type: execute
wave: 2
depends_on: []
files_modified:
  - backend/internal/service/push.go
  - backend/internal/config/config.go
  - .env
  - backend/cmd/kanso-api/main.go
autonomous: false
requirements:
  - LO-02: FCM legacy HTTP API migrado para HTTP v1 com OAuth2
must_haves:
  truths:
    - Push service usa FCM HTTP v1 API (https://fcm.googleapis.com/v1/projects/{project_id}/messages:send)
    - Autenticação via OAuth2 access token (não server key)
    - Service account JSON configurado via env var (GOOGLE_APPLICATION_CREDENTIALS ou FCM_SERVICE_ACCOUNT)
    - Token OAuth2 é renovado automaticamente quando expira (ou em cada requisição)
  artifacts:
    - path: backend/internal/service/push.go
      provides: "FCM HTTP v1 integration"
      contains: "google.golang.org/api/fcm"
      contains: "oauth2.TokenSource"
---

<objective>
Migrar o push notification service da FCM Legacy HTTP API (deprecada) para a FCM HTTP v1 API com autenticação OAuth2.
</objective>

<execution_context>
@backend/internal/service/push.go
@backend/internal/config/config.go
@backend/cmd/kanso-api/main.go
@.env
</execution_context>

---

## Tasks

### Task 1: Adicionar dependências Go

**Ação:**

```bash
cd backend && go get golang.org/x/oauth2 google.golang.org/api/fcm google.golang.org/api/option
```

**Verificação:**
- `go build ./...` exit 0
- `go vet ./...` exit 0

### Task 2: Criar service account e configurar env var

**Ação manual (usuário):**

1. No Google Cloud Console > IAM > Service Accounts, criar uma service account com permissão `Firebase Cloud Messaging API Admin`
2. Gerar chave JSON e salvar como `.secrets/fcm-service-account.json` (já gitignorado)
3. Adicionar ao `.env`:
   ```
   GOOGLE_APPLICATION_CREDENTIALS=.secrets/fcm-service-account.json
   ```
   Ou, alternativa mais segura: codificar o JSON em base64 e passar como env var:
   ```
   FCM_SERVICE_ACCOUNT_JSON=<base64 do JSON>
   ```

**Nota:** A abordagem via env var é mais segura em containers que arquivo montado.

### Task 3: Refatorar PushService para usar FCM HTTP v1

**Arquivo:** `backend/internal/service/push.go`

**Ação:**

1. Adicionar campo `ProjectID string` ao `PushService`
2. Remover `FCMServerKey string` e `FCMURL string`
3. Adicionar campo `TokenSource oauth2.TokenSource`
4. Modificar `NewPushService` para aceitar `projectID` e `tokenSource` em vez de `fcmServerKey`
5. Reimplementar `Send()`:

```go
import (
    "context"
    fcm "google.golang.org/api/fcm/v1"
    "google.golang.org/api/option"
)

func (s *PushService) Send(sub string) error {
    prefs, err := s.CouchRepo.GetPushPrefs(sub)
    if err != nil {
        return fmt.Errorf("get prefs: %w", err)
    }
    if prefs == nil || prefs.FCMToken == "" {
        return fmt.Errorf("no FCM token for user %s", sub)
    }

    ctx := context.Background()
    fcmSvc, err := fcm.NewService(ctx, option.WithTokenSource(s.TokenSource))
    if err != nil {
        return fmt.Errorf("fcm new service: %w", err)
    }

    msg := &fcm.Message{
        Token: prefs.FCMToken,
        Notification: &fcm.Notification{
            Title: "Kanso",
            Body:  "Como você está se sentindo agora?",
        },
    }

    resp, err := fcmSvc.Projects.Messages.Send("projects/"+s.ProjectID, msg).Do()
    if err != nil {
        return fmt.Errorf("fcm send: %w", err)
    }
    _ = resp
    return nil
}
```

**Nota:** A Google API Go client library gerencia renovação automática do token OAuth2.

### Task 4: Atualizar config e wiring em main.go

**Arquivo:** `backend/internal/config/config.go`, `backend/cmd/kanso-api/main.go`

**Ação:**

1. Em config, adicionar campos:
   ```go
   FCMProjectID      string
   FCMServiceAccount string // base64 JSON ou caminho do arquivo
   ```

2. Em main.go, inicializar token source:
   ```go
   import "golang.org/x/oauth2/google"

   var ts oauth2.TokenSource
   if cfg.FCMServiceAccount != "" {
       creds, err := google.CredentialsFromJSON(ctx, []byte(cfg.FCMServiceAccount), fcm.CloudPlatformScope)
       if err != nil {
           log.Fatalf("failed to parse FCM service account: %v", err)
       }
       ts = creds.TokenSource
   }
   ```

3. Passar `ts` e `projectID` para `NewPushService`.

**Verificação:**
- `go build ./...` exit 0
- `go vet ./...` exit 0
- Push notification ainda funciona (teste manual ou integração)

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-LO-02-01 | Deprecated API | Push service | mitigate | Migrado para FCM HTTP v1 — API ativa e suportada |
| T-LO-02-02 | Static Credential | Push service | mitigate | Substituído server key estático por OAuth2 com renovação automática |

## Verification

- `go build ./...` compila sem erros
- `go vet ./...` passa
- Push notification enviado com sucesso (teste manual via scheduler ou curl)

## Success Criteria

1. FCM HTTP v1 implementado com `google.golang.org/api/fcm`
2. Autenticação via OAuth2 com service account (não server key)
3. Server key legacy removido do código
4. Build e vet passam
