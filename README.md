# Kanso — Diário Emocional Offline-First

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go" alt="Go 1.26"/>
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react" alt="React 19"/>
  <img src="https://img.shields.io/badge/TypeScript-5.8-3178C6?logo=typescript" alt="TypeScript 5.8"/>
  <img src="https://img.shields.io/badge/Tailwind-4.3-06B6D4?logo=tailwindcss" alt="Tailwind 4.3"/>
  <img src="https://img.shields.io/badge/PWA-Offline--first-5A0FC8?logo=pwa" alt="PWA"/>
  <img src="https://img.shields.io/badge/CouchDB-3.5-E42528?logo=apachecouchdb" alt="CouchDB 3.5"/>
  <img src="https://img.shields.io/badge/Docker-2496ED?logo=docker" alt="Docker"/>
  <img src="https://img.shields.io/badge/Traefik-3-24A1C1?logo=traefikproxy" alt="Traefik 3"/>
  <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT License"/>
</p>

**Kanso** (japonês 簡素 — simplicidade, clareza, essencialidade) é um PWA de diário emocional
offline-first que ajuda o usuário a nomear e registrar emoções no momento em que são percebidas,
com sincronização automática para a nuvem e geração de relatórios PDF para compartilhar com
o terapeuta.

## Funcionalidades

### Implementadas

- **Autenticação Google OAuth** — login com conta Google, JWT (HS256) com refresh automático
- **Registro emocional estruturado** — 4 campos: sensações, sentimento (combobox customizável),
  contexto e pensamentos, com data/hora (suporte a registro retroativo)
- **Offline-first** — dados salvos localmente no PouchDB (IndexedDB), sincronização automática
  com CouchDB quando online via JWT auth
- **Combobox de sentimentos** — autocomplete com Headless UI v2, cria novos sentimentos
  automaticamente ao digitar
- **Histórico cronológico** — lista de registros em ordem reversa com cards expansíveis
- **Análise NLP de emoções** — watcher em goroutine observa `_changes` feed do CouchDB,
  envia texto para serviço Python (BERTimbau fine-tuned via gRPC com TLS), resultados
  salvos como `analise_nlp` e exibidos como chips coloridos
- **Relatórios PDF** — geração assíncrona via chromedp em container dedicado, com polling
  de status e download autenticado (inclui resumo de emoções + chips por registro)
- **Notificações push (FCM v1)** — scheduler Go dispara lembretes nos horários configurados,
  autenticação via OAuth2 com service account
- **Navegação por abas** — Registrar, Histórico e Perfil (configurações)
-   **Segurança** — auditoria completa com 19/24 itens corrigidos:
  - `validate_doc_update` para isolamento de dados por usuário no CouchDB (CR-03)
  - JWT secret forte (256-bit) + validação de algoritmo HMAC (CR-02, ME-03)
  - Push endpoint autenticado via API key + JWT admin (CR-04)
  - Security headers via Traefik (CSP, HSTS, XFO, XCTO) (HI-07)
  - Traefik File Provider — sem Docker socket (HI-01)
  - CouchDB sem porta exposta ao host (ME-02)
  - gRPC com TLS auto-assinado entre API e NLP (HI-02)
  - Containers não-root (backend, scheduler, NLP) (HI-03/04/05)
  - Logs estruturados com slog, sem PII (LO-04)
  - Database names como constantes (IN-03)
  - Chromedp sem flags inseguras (CR-05)
  - Rate limiting via Traefik (10 req/min auth, 30 req/min db)

### Planejado (backlog)

- Emotion chips — melhorar visualização no frontend e relatório PDF
- WhatsApp automático — enviar relatório via Twilio
- JWT em HttpOnly cookie (ME-05)
- Dark mode
- Bloqueio biométrico/PIN
- Deploy VPS Hostinger

## Stack

| Camada | Tecnologia | Versão |
|--------|-----------|--------|
| Frontend | React + Vite + TypeScript + Tailwind CSS | 19.2 / 8 / 5.8 / 4.3 |
| Estado local | PouchDB + PouchDB Find | 9.0 |
| Backend | Go + Chi + chromedp | 1.26 / 5.2 |
| Autenticação | Google OAuth + JWT (HS256) | — |
| Banco de dados | CouchDB | 3.5 |
| Proxy reverso | Traefik v3 (File Provider) | 3 |
| Notificações | FCM HTTP v1 (OAuth2) | — |
| NLP | Python + gRPC + BERTimbau | 3.12 / 1.69 |
| Infraestrutura | Docker Compose | — |

## Arquitetura

```
Browser (PWA)
  ├── React (UI)
  ├── PouchDB (IndexedDB) — offline-first
  └── Live Sync ─── {live:true, retry:true}
        ├── Dev:  /db/*  → Vite proxy → localhost:80 → Traefik → CouchDB (HTTP)
        └── Prod: https://kanso.app/db/*  → Traefik → CouchDB (HTTPS + JWT)
                                                             ↕
                                                      Go API (chi)
                                                        ├── Google OAuth / JWT
                                                        ├── PDF (chromedp container)
                                                        ├── NLP Watcher (goroutine)
                                                        │   └── gRPC + TLS ──→ Python NLP (BERTimbau)
                                                        ├── FCM Scheduler
                                                        └── Twilio (futuro)

  Traefik v3 (File Provider — sem Docker socket)
  ├── api → Go API :8080
  ├── db  → CouchDB :5984 (JWT auth + validate_doc_update)
  └── Security: CSP, HSTS, XFO, XCTO, rate-limit, CORS
```

O frontend sincroniza **diretamente** com o CouchDB via PouchDB — o Go backend não participa
de operações CRUD de registros. O backend lida com autenticação, efeitos colaterais
(geração de PDF, disparo de notificações, etc.) e o **NLP Watcher** — uma goroutine que
observa o feed `_changes` do CouchDB via gRPC com TLS e envia novos registros para análise.

## Pré-requisitos

- **Go** 1.26+
- **Node.js** 22+ e **pnpm** 9+
- **Docker** + **Docker Compose** v2
- **Google Cloud Platform** — projeto com **Google Identity Services** habilitado
  e um **Client ID do tipo Web** configurado (OAuth 2.0)

## Configuração do Ambiente

### 1. Clone o repositório

```bash
git clone https://github.com/<seu-user>/kanso.git
cd kanso
```

### 2. Configure o Google OAuth

1. Acesse [Google Cloud Console](https://console.cloud.google.com)
2. Crie um projeto ou selecione um existente
3. Habilite a API **Google Identity Services**
4. Crie uma credencial **OAuth 2.0 Client ID** do tipo **Web application**
5. Adicione como **Authorized JavaScript origins**:
   - `https://kanso.local`
   - `http://localhost:5173`
6. Adicione como **Authorized redirect URIs**:
   - `https://kanso.local/login`
   - `http://localhost:5173/login`

### 3. Configure as variáveis de ambiente

```bash
cp .env.example .env
# Edite o .env com seus valores reais
```

Gere os secrets obrigatórios:
```bash
openssl rand -base64 32   # JWT_SECRET
openssl rand -base64 12   # COUCHDB_PASSWORD
openssl rand -base64 32   # SCHEDULER_API_KEY
```

### 4. Gere os certificados gRPC (TLS)

```bash
bash infra/certs/gen-grpc-certs.sh
```

### 5. Suba a infraestrutura

```bash
make up
```

Isso inicia via Docker Compose:

| Serviço | Função | Acesso |
|---------|--------|--------|
| **Traefik** | Proxy reverso TLS (File Provider) | `kanso.local:443` |
| **CouchDB** | Banco de dados (sem porta exposta ao host) | Rede Docker interna |
| **API** | Go backend | `api:8080` |
| **Chromedp** | Headless Chrome para PDFs | `chromedp:9222` |
| **NLP** | Análise de emoções (BERTimbau) | `nlp:50051` (gRPC + TLS) |
| **Scheduler** | Disparo de push notifications | Rede Docker interna |

### 6. Inicie o frontend

```bash
cd frontend
pnpm install
pnpm dev
# ou: make dev
```

O frontend estará disponível em `http://localhost:5173`.

O Vite proxy roteia:
- `/api` → `http://localhost:8080` (Go backend)
- `/db` → `http://localhost:80` → Traefik → CouchDB (dev, via HTTP)

Em produção, o PouchDB sync usa `VITE_COUCHDB_URL` absoluto (ex: `https://kanso.app/db`) diretamente — sem Vite proxy.

**Importante (produção):** Adicione `127.0.0.1 kanso.local` ao `/etc/hosts` para testes locais com HTTPS.

## Testes

```bash
# Todos os testes (via Make)
make test

# Frontend (Vitest + Testing Library)
cd frontend && pnpm test

# Backend — testes unitários
cd backend && go test ./...

# Backend — verificação de tipos
cd backend && go vet ./...

# NLP service (Python)
cd nlp-service && python -m pytest
```

## Estrutura do Projeto

```
kanso/
├── backend/                    # Go API (chi)
│   ├── cmd/kanso-api/main.go   # Entry point, rotas, DI
│   └── internal/
│       ├── config/             # Variáveis de ambiente
│       ├── handler/            # Handlers HTTP (auth, push, report)
│       ├── service/            # Lógica de negócio (auth, push, report, watcher)
│       ├── repository/         # Acesso a dados CouchDB (couchdb.go)
│       ├── middleware/         # JWT middleware com validação HMAC
│       ├── nlp/                # Cliente gRPC + TLS para NLP service
│       ├── pdf/                # Gerador PDF (chromedp remote)
│       └── templates/          # Templates HTML para PDF
├── frontend/                   # React PWA
│   ├── src/
│   │   ├── components/         # UI (RegistroCard, SyncStatus, etc.)
│   │   ├── hooks/              # useAuth, usePouchSync, usePushNotifications
│   │   ├── pages/              # Login, Register, History, Profile
│   │   ├── services/           # auth, pouchdb, registros, reports, push
│   │   └── types/              # TypeScript types
│   └── vite.config.ts
├── scheduler/                  # Go scheduler para push notifications (FCM v1)
├── nlp-service/                # Python gRPC + BERTimbau fine-tuned
│   ├── src/                    # Servidor gRPC com TLS, classificador
│   └── proto/                  # Definições protobuf
├── infra/
│   ├── couchdb/                # Config JWT auth (local.ini gitignored)
│   ├── traefik/                # Traefik v3 (File Provider — sem Docker socket)
│   ├── chromedp/               # Container dedicado headless-shell
│   ├── certs/                  # Certificados gRPC TLS (gitignored)
│   └── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

## API Endpoints

### Públicos (sem autenticação)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/health` | Health check |
| POST | `/api/auth/google` | Login com Google (recebe `idToken`) |
| POST | `/api/auth/refresh` | Renova o JWT via cookie |
| POST | `/api/auth/logout` | Limpa a sessão |

### Protegidos (`Authorization: Bearer <jwt>`)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/auth/me` | Dados do usuário autenticado |
| POST | `/api/reports` | Solicita geração de relatório (202 + `jobId`) |
| GET | `/api/reports` | Lista relatórios do usuário |
| GET | `/api/reports/{id}` | Status de um relatório |
| GET | `/api/reports/{id}/download` | Download do PDF gerado |
| POST | `/api/push/subscribe` | Registra subscription FCM + preferências |

### Internas (rede Docker — API key)

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/api/push/send` | Dispara notificações push (scheduler) |

## Variáveis de Ambiente

### Backend (`backend/`)

| Variável | Descrição |
|----------|-----------|
| `PORT` | Porta do servidor (padrão: `8080`) |
| `COUCHDB_URL` | URL do CouchDB (padrão: `http://couchdb:5984`) |
| `COUCHDB_USER` | Usuário admin CouchDB (padrão: `admin`) |
| `COUCHDB_PASSWORD` | **Obrigatória** — senha do CouchDB |
| `JWT_SECRET` | **Obrigatória** — chave JWT 256-bit |
| `GOOGLE_CLIENT_ID` | **Obrigatória** — Client ID Google OAuth |
| `SCHEDULER_API_KEY` | **Obrigatória** — chave do scheduler p/ push |
| `FCM_PROJECT_ID` | Project ID Firebase (FCM v1) |
| `FCM_SERVICE_ACCOUNT_B64` | Service account JSON em base64 (FCM v1) |
| `GRPC_CA_CERT` | Caminho CA cert gRPC TLS (padrão: `/certs/ca.crt`) |

### Frontend (`frontend/`)

| Variável | Descrição |
|----------|-----------|
| `VITE_API_URL` | URL base da API (padrão: `/api`) |
| `VITE_COUCHDB_URL` | URL do CouchDB (padrão: `/db` — Vite proxy em dev; absoluto em produção como `https://kanso.app/db`) |
| `VITE_GOOGLE_CLIENT_ID` | Client ID Google OAuth |
| `VITE_VAPID_PUBLIC_KEY` | VAPID key para Web Push |

### Docker Compose (`infra/docker-compose.yml`)

| Variável | Obrigatória | Descrição |
|----------|-------------|-----------|
| `COUCHDB_PASSWORD` | ✅ | Senha do CouchDB |
| `JWT_SECRET` | ✅ | Chave JWT 256-bit |
| `GOOGLE_CLIENT_ID` | ✅ | Client ID Google OAuth |
| `SCHEDULER_API_KEY` | ✅ | Chave do scheduler |
| `FCM_PROJECT_ID` | — | Project ID Firebase |
| `FCM_SERVICE_ACCOUNT_B64` | — | Service account JSON base64 |

---

## Licença

MIT — Código aberto para portfólio.

---

*Kanso — Simplicidade. Clareza. Essencialidade.*
