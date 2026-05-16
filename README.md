# Kanso — Diário Emocional Offline-First

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

### 4. Suba a infraestrutura (CouchDB + API)

```bash
cd infra
docker compose up
```

Isso inicia:
- **CouchDB** em `localhost:5984`
- **API Go** em `localhost:8080`

Aguarde o health check do CouchDB antes de usar o sistema.

### 5. Inicie o frontend

Em outro terminal:

```bash
cd frontend
pnpm install
pnpm dev
```

O frontend estará disponível em `http://localhost:5173`.

> **⚠️ Limitação atual:** O frontend possui URLs hardcoded para `https://kanso.local/api`
> e `https://kanso.local/db` (veja a seção [Dívida Técnica](#-dívida-técnica) abaixo).
> Para desenvolvimento local sem Traefik, você precisará:
> - Alterar `API_BASE` em `frontend/src/services/auth.ts` para `http://localhost:8080/api`
> - Alterar `COUCHDB_URL` em `frontend/src/services/pouchdb.ts` para `http://localhost:5984/db`

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
│   └── docker-compose.yml     # CouchDB + API
├── nlp-service/               # (vazio — v2)
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

| Variável | Descrição |
|----------|-----------|
| `VITE_GOOGLE_CLIENT_ID` | Client ID do Google OAuth (mesmo do backend) |

### Docker Compose (`infra/docker-compose.yml`)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `COUCHDB_PASSWORD` | `admin123` | Senha do CouchDB |
| `JWT_SECRET` | `dev-secret-change-in-production` | Chave JWT |
| `GOOGLE_CLIENT_ID` | `dev-client-id` | Client ID Google OAuth |

---

## 🧱 Dívida Técnica

Os itens abaixo impedem o setup de desenvolvimento fora do ecossistema Docker completo
e devem ser resolvidos antes de novas funcionalidades:

| # | Item | Arquivos | Impacto | Prioridade |
|---|------|----------|---------|------------|
| 1 | **URLs da API hardcoded** | `auth.ts` (`https://kanso.local/api`), `pouchdb.ts` (`https://kanso.local/db`) | Bloqueia dev em localhost sem editar código | 🔴 P0 |
| 2 | **Sem middleware CORS no backend** | `main.go` (não usa `go-chi/cors`) | Requisições cross-origin falham no dev sem proxy | 🔴 P0 |
| 3 | **Frontend sem proxy Vite** | `vite.config.ts` (sem `server.proxy`) | Depende de Traefik para roteamento dev | 🔴 P0 |
| 4 | **Sem `.env.example`** | — | Novo dev não sabe quais variáveis configurar | 🟡 P1 |
| 5 | **Traefik ausente do docker-compose** | `infra/docker-compose.yml` | Sem HTTPS/TLS local; PWA service worker não funciona | 🟡 P1 |
| 6 | **Sem Makefile / justfile** | — | Cada serviço usa comando diferente; sem `make up` unificado | 🟡 P1 |
| 7 | **nlp-service vazio** | `nlp-service/` (diretório vazio) | Falso positivo — parece implementado mas não está | 🟢 P2 |

### Próximos passos (execução imediata)

1. Extrair `API_BASE` e `COUCHDB_URL` para variáveis de ambiente do Vite (`VITE_API_URL`,
   `VITE_COUCHDB_URL`) e atualizar os serviços
2. Adicionar middleware CORS no backend (`go-chi/cors`)
3. Configurar `server.proxy` no `vite.config.ts` para rotear `/api` e `/db` para os
   serviços corretos em dev

---

## Licença

MIT — Código aberto para portfólio. Consulte o arquivo `LICENSE` (não incluído — adicionar
conforme necessário).

---

*Kanso — Simplicidade. Clareza. Essencialidade.*
