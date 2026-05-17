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

### Implementadas (v1)

- **Autenticação Google OAuth** — login com conta Google, JWT com refresh automático
- **Registro emocional estruturado** — 4 campos: sensações, sentimento (combobox customizável),
  contexto e pensamentos, com data/hora (suporte a registro retroativo)
- **Offline-first** — dados salvos localmente no PouchDB (IndexedDB), sincronização automática
  com CouchDB quando online (`live: true, retry: true`)
- **Combobox de sentimentos** — autocomplete com Headless UI v2, cria novos sentimentos
  automaticamente ao digitar
- **Navegação por abas** — Registrar, Histórico (placeholder) e Perfil
- **Relatórios PDF** — geração assíncrona via chromedp + headless-shell, com polling de status
  e download autenticado

### Futuro (v2)

- Histórico com busca/filtros
- Análise NLP de emoções (Python + transformers)
- Notificações push (FCM)
- Envio de relatórios via WhatsApp (Twilio)
- Dark mode
- Bloqueio biométrico/PIN

## Stack

| Camada | Tecnologia | Versão |
|--------|-----------|--------|
| Frontend | React + Vite + TypeScript + Tailwind CSS | 19.2 / 8 / 5.8 / 4.3 |
| Estado local | PouchDB + PouchDB Find | 9.0 |
| Backend | Go + Chi + chromedp | 1.26 / 5.2 |
| Autenticação | Google OAuth + JWT (HS256) | — |
| Banco de dados | CouchDB | 3.5 |
| Infraestrutura | Docker Compose + Traefik v3 | — |
| NLP (futuro) | Python + FastAPI + transformers | — |

## Arquitetura

```
Browser (PWA)
  ├── React (UI)
  ├── PouchDB (IndexedDB) — offline-first
  └── Live Sync ─── {live:true, retry:true} ──→ CouchDB
                              ↕
                        Go API (chi)
                          ├── Google OAuth / JWT
                          ├── PDF (chromedp)
                          ├── FCM (futuro)
                          └── Twilio (futuro)
                              ↕
                    Python NLP (futuro — interno)
```

O frontend sincroniza **diretamente** com o CouchDB via PouchDB — o Go backend não participa
de operações CRUD de registros. O backend lida apenas com autenticação e efeitos colaterais
(geração de PDF, disparo de notificações, etc.).

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
   - `https://kanso.local` (com Traefik/TLS)
   - `http://localhost:5173` (dev sem Traefik)
6. Adicione como **Authorized redirect URIs**:
   - `https://kanso.local/login` (com Traefik/TLS)
   - `http://localhost:5173/login` (dev sem Traefik)

### 3. Configure as variáveis de ambiente

Copie o arquivo de exemplo e preencha:

```bash
cp .env.example .env
# Edite o .env com seus valores
```

### 4. Suba a infraestrutura

```bash
make up
```

Isso inicia via Docker Compose:
- **Traefik** em `localhost:443` (TLS)
- **CouchDB** em `localhost:5984`
- **API Go** em `localhost:8080`

### 5. Inicie o frontend

Em outro terminal:

```bash
cd frontend
pnpm install
pnpm dev
# ou: make dev
```

O frontend estará disponível em `http://localhost:5173`.

O Vite proxy roteia automaticamente:
- `/api` → `http://localhost:8080`
- `/db` → `http://localhost:5984`

### Alternativa: rodar o backend local (sem Docker)

```bash
cd backend
export COUCHDB_URL=http://localhost:5984
export COUCHDB_PASSWORD=admin123
export JWT_SECRET=dev-secret-change-in-production
export GOOGLE_CLIENT_ID=seu-client-id
go run ./cmd/kanso-api
```

> O backend pode rodar localmente ou em Docker. O CouchDB precisa estar acessível.

## Testes

```bash
# Frontend (Vitest + Testing Library)
cd frontend && pnpm test

# Backend — testes unitários
cd backend && go test ./...

# Backend — testes de integração (requer Chrome/headless-shell)
cd backend && go test ./... -tags=integration

# Backend — verificação de tipos
cd backend && go vet ./...
```

## Estrutura do Projeto

```
kanso/
├── backend/                    # Go API (chi)
│   ├── cmd/kanso-api/main.go   # Entry point, rotas, DI
│   └── internal/
│       ├── config/config.go    # Variáveis de ambiente
│       ├── handler/            # Handlers HTTP
│       ├── service/            # Lógica de negócio
│       ├── repository/         # Acesso a dados (CouchDB)
│       ├── middleware/         # JWT middleware
│       ├── pdf/                # Gerador PDF (chromedp)
│       └── templates/          # Templates HTML para PDF
├── frontend/                   # React PWA
│   ├── src/
│   │   ├── components/        # Componentes UI
│   │   ├── hooks/             # Hooks React (useAuth, usePouchSync)
│   │   ├── pages/             # Páginas (Login, Register, Profile, History)
│   │   ├── services/          # Serviços (auth, pouchdb, registros, reports)
│   │   └── types/             # TypeScript types
│   └── vite.config.ts
├── infra/
│   ├── docker-compose.yml     # CouchDB + API + Traefik
│   └── traefik/               # Configuração do Traefik v3
├── Makefile                   # Comandos unificados (up/down/dev/test)
├── nlp-service/               # (v2 — análise NLP de emoções)
├── .env.example               # Exemplo de variáveis de ambiente
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

### Protegidos (requerem `Authorization: Bearer <jwt>`)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/auth/me` | Dados do usuário autenticado |
| POST | `/api/reports` | Solicita geração de relatório (retorna 202 + `jobId`) |
| GET | `/api/reports` | Lista relatórios do usuário |
| GET | `/api/reports/{id}` | Status de um relatório |
| GET | `/api/reports/{id}/download` | Download do PDF gerado |

## Variáveis de Ambiente

### Backend (`backend/`)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `8080` | Porta do servidor |
| `COUCHDB_URL` | `http://couchdb:5984` | URL de conexão com CouchDB |
| `COUCHDB_USER` | `admin` | Usuário admin do CouchDB |
| `COUCHDB_PASSWORD` | — | Senha do CouchDB |
| `JWT_SECRET` | — | Chave para assinar JWTs (trocar em produção) |
| `GOOGLE_CLIENT_ID` | — | Client ID do Google OAuth |
| `PDF_TMP_DIR` | `/tmp/kanso-pdf` | Diretório para PDFs temporários |
| `CHROMEDP_PATH` | — | Caminho do headless Chrome (auto no container) |

### Frontend (`frontend/`)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `VITE_API_URL` | `/api` | URL base da API (proxy Vite em dev) |
| `VITE_COUCHDB_URL` | `/db` | URL base do CouchDB (proxy Vite em dev) |
| `VITE_GOOGLE_CLIENT_ID` | — | Client ID do Google OAuth (mesmo do backend) |

### Docker Compose (`infra/docker-compose.yml`)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `COUCHDB_PASSWORD` | `admin123` | Senha do CouchDB |
| `JWT_SECRET` | `dev-secret-change-in-production` | Chave JWT |
| `GOOGLE_CLIENT_ID` | `dev-client-id` | Client ID Google OAuth |

---

## Dívida Técnica

Todos os itens identificados durante o MVP foram resolvidos na Phase 4:

| # | Item | Status | Resolvido Em |
|---|------|--------|-------------|
| 1 | **URLs da API hardcoded** → env vars `VITE_API_URL` / `VITE_COUCHDB_URL` | ✅ | 2026-05-17 |
| 2 | **CORS middleware** (`go-chi/cors` no backend) | ✅ | 2026-05-17 |
| 3 | **Vite proxy** (`/api` → `:8080`, `/db` → `:5984`) | ✅ | 2026-05-17 |
| 4 | **Traefik** no docker-compose com TLS | ✅ | 2026-05-17 |
| 5 | **Makefile** com comandos unificados | ✅ | 2026-05-17 |
| 6 | **nlp-service/README.md** com documentação v2 | ✅ | 2026-05-17 |

**Ambiente de desenvolvimento local** configurado: `make up` + `make dev` sem dependência de DNS ou TLS.

---

## Licença

MIT — Código aberto para portfólio. Consulte o arquivo `LICENSE` (não incluído — adicionar
conforme necessário).

---

*Kanso — Simplicidade. Clareza. Essencialidade.*
