# Phase 8: V3 — Integração & Qualidade — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Phase:** 8-V3-Integracao-Qualidade
**Areas discussed:** Isolamento CouchDB, Gerenciamento de Secrets, CSP/Headers, Chromedp flags, Push endpoint

---

## Isolamento de Dados CouchDB (CR-03)

| Option | Description | Selected |
|--------|-------------|----------|
| Bancos por usuario | Cada usuario tem seus proprios databases. PouchDB sync mantido. Isolamento real. | |
| Proxy via API Go | Remove PouchDB sync direto. Todo acesso via API Go. Perde offline-first. | |
| validate_doc_update | Mantem DBs compartilhados com funcoes CouchDB. Complexo. | |

**User's choice:** Nenhuma das acima. O usuario sugeriu abordagem propria: documento `config:active_users` no CouchDB com allowlist de emails + bloqueio de novas contas + aprovacao manual.
**Notes:** App e uso pessoal, nao vai liberar para outros usuarios agora. Comecar do zero — ninguem tem acesso ate aprovacao manual.

Armazenamento da allowlist:
| Option | Description | Selected |
|--------|-------------|----------|
| Variavel de ambiente | Lista fixa via env var. Exige restart. | |
| Documento no CouchDB | Documento separado config:users. Pode adicionar sem restart. | ✓ |
| Campo active no doc usuario | Cada usuario tem active: true/false. | |

**User's choice:** Documento separado no CouchDB

Usuarios existentes:
| Option | Description | Selected |
|--------|-------------|----------|
| Manter usuarios atuais | Todos que ja tem conta continuam ativos. | |
| Comecar do zero | Ninguem tem acesso ate ativar manualmente. | ✓ |
| So voce como admin | Apenas seu email tem acesso garantido. | |

**User's choice:** Comecar do zero

---

## Gerenciamento de Secrets (CR-01, CR-02)

OAuth revogado?
| Option | Description | Selected |
|--------|-------------|----------|
| Sim, ja revoguei | Secret ja invalidado. | |
| Nao, ainda nao | Precisa revogar antes de implementar. | ✓ |
| Nao sei como | Mostrar passo a passo. | |

**User's choice:** Nao revogou ainda. Deixou para fazer depois.

Onde armazenar secrets:
| Option | Description | Selected |
|--------|-------------|----------|
| .env (gitignored) | Simples, ja no .gitignore. | ✓ |
| Docker secrets | Mais seguro mas exige mudancas. | |

**User's choice:** .env (gitignored)

local.ini no git:
| Option | Description | Selected |
|--------|-------------|----------|
| Remover do git + gerar no deploy | Adicionar ao .gitignore. Gerar dinamicamente. | |
| Manter no git com placeholder | Manter placeholder, substituir via env var. | ✓ |

**User's choice:** Manter no git com placeholder

---

## CSP e Headers (HI-07)

| Option | Description | Selected |
|--------|-------------|----------|
| Restritiva | So 'self', accounts.google.com, fcm.googleapis.com, fonts.googleapis.com, fonts.gstatic.com | ✓ |
| Minima | So HSTS + X-Content-Type-Options + X-Frame-Options. CSP depois. | |

**User's choice:** Restritiva (recomendado)

---

## Chromedp Flags (CR-05)

| Option | Description | Selected |
|--------|-------------|----------|
| Remover ambas | Nao sao necessarias para gerar PDF. | ✓ |
| Deixar condicional | Manter em dev, remover em prod via env var. | |

**User's choice:** Remover ambas

---

## Protecao Push Endpoint (CR-04)

| Option | Description | Selected |
|--------|-------------|----------|
| API key compartilhada | Scheduler envia X-API-Key. Chave via env var. | |
| So rede interna | Restringir via middleware para subnet Docker. | ✓ |

**User's choice:** So rede interna Docker

---

## Deferred Ideas

- Refatorar emotion chips (outra sub-atividade da Fase 8)
- WhatsApp automatico (outra sub-atividade da Fase 8)
- Migrar JWT para HttpOnly cookie (deferido)
- gRPC com TLS (manter insecure por ora)
- Docker secrets (preferiu .env)
- Atualizar dependencias (nao discutido)
- Containers como root (nao discutido)
- CouchDB porta 5984 exposta (nao discutido)
- Proxy /db do Vite (nao discutido)
- Validacao algoritmo JWT (nao discutido)

---

*Phase: 8-V3-Integracao-Qualidade*
*Discussion log: 2026-05-23*
