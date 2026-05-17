# Phase 6: Push Notifications — Context & Decisions

## Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | Push provider | FCM (Firebase) | Já temos conta Google; SDK Go maduro; documentação robusta |
| 2 | Scheduler architecture | Microsserviço Go separado (`kanso-scheduler`) | Desacopla responsabilidade; Go já é stack do backend |
| 3 | Token/preference storage | CouchDB | Mesmo banco do resto do app; consistência |
| 4 | Scheduler → Backend | Scheduler chama API do Go backend (`POST /api/push/send`) | Go backend já tem pushSvc com FCM SDK; evita duplicação |
| 5 | Tap notification UX | Abre tela de registro com data/hora atuais | Melhor experiência — captura o momento |
| 6 | Fuso horário | Detectado do browser (`Intl.DateTimeFormat`), salvo no CouchDB | Precisão sem config manual |
| 7 | Permissão notificação | Pedir no primeiro login | Timing natural; usuário já engajado |

## Architecture Overview

```
┌─────────────┐    subscribe token    ┌──────────────┐
│  PWA/React  │ ────────────────────▶ │  Go Backend  │
│  + SW       │   POST /api/push/     │  (kanso-api) │
│             │   subscribe           │              │
│             │ ◀──────────────────── │  pushSvc     │
│  Notificação │   FCM push message    │  (FCM SDK)   │
└─────────────┘                       └──────┬───────┘
                                             │ POST /api/push/send
                                             │
                                    ┌────────▼───────┐
                                    │  Scheduler      │
                                    │  (kanso-sched)  │
                                    │  Go ticker loop │
                                    │                 │
                                    │  Lê: CouchDB    │
                                    │  (prefs + tz)   │
                                    └─────────────────┘
```

## Data Model (CouchDB)

### User Push Preferences Document
```json
{
  "_id": "push_prefs:{userId}",
  "type": "push_prefs",
  "userId": "...",
  "enabled": true,
  "times": ["12:00", "18:00", "23:00"],
  "timezone": "America/Sao_Paulo",
  "fcmToken": "...",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

## Routes (Go Backend)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/push/subscribe | Recebe FCM token + timezone do browser |
| PUT | /api/push/preferences | Atualiza preferências (horários, enabled) |
| GET | /api/push/preferences | Retorna preferências do usuário |
| POST | /api/push/send | (Internal) Scheduler chama para disparar push |

## Stack

- **Frontend**: Service Worker (`push` + `notificationclick` events), Firebase JS SDK (ou Web Push API)
- **Backend**: `firebase.google.com/go/v4` — Admin SDK for FCM
- **Scheduler**: Go puro, chi ou apenas stdlib, ticker loop
- **Infra**: Novo Docker service no docker-compose (`kanso-scheduler`)
