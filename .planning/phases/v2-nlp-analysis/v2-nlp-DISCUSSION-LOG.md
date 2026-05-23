# Milestone v2: NLP Analysis — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Milestone:** v2 — NLP Analysis
**Areas discussed:** Trigger Model, Modelo de Emoções, Arquitetura Go→Python, Gestão do Modelo, Schema de Emoções, Enriquecimento

---

## Phase Classification

| Option | Description | Selected |
|--------|-------------|----------|
| Fase 7 — NLP Analysis | Próxima fase sequencial após Phase 6 | |
| Fase 7 — NLP + WhatsApp | Agrupar as duas features v2 | |
| Parte de v2 Milestone | Criar milestone v2 com sub-fases | ✓ |

**User's choice:** Parte de v2 Milestone
**Notes:** NLP será dividido em 3 sub-fases: Infra → Modelo → Integração

---

## Sub-fases Structure

| Option | Description | Selected |
|--------|-------------|----------|
| Infra → Modelo → Integração | 1: Python/FastAPI/Docker, 2: Modelo+análise, 3: Go+CouchDB | ✓ |
| Backend → Frontend | 1: Serviço NLP completo, 2: UI display | |
| Async Job → Streaming | 1: Batch pós-registro, 2: Streaming em tempo real | |

**User's choice:** Infra → Modelo → Integração

---

## Trigger Model

| Option | Description | Selected |
|--------|-------------|----------|
| A cada novo registro | Análise disparada quando registro sync ao CouchDB | ✓ |
| Durante geração de relatório | Batch apenas na geração do PDF | |
| Job schedulado (1x/dia) | Processa registros não analisados periodicamente | |

**User's choice:** A cada novo registro
**Notes:** Go backend escuta CouchDB _changes feed. Backfill de registros existentes no primeiro deploy.

---

## Go Backend → Python Communication

| Option | Description | Selected |
|--------|-------------|----------|
| HTTP (FastAPI) | POST /analyze, simples, já stack planejado | |
| Fila (Redis/RabbitMQ) | Desacoplado, mas adiciona infra | |
| gRPC | Performance, schema typed | ✓ |

**User's choice:** gRPC
**Notes:** Request schema: registro completo (id, sensacoes, contexto, pensamentos, dataHora). Response: emocaoPrincipal, emocoesSecundarias[], scores, intensidade, analiseAdicional, metadadosModelo.

---

## Modelo de Emoções

| Option | Description | Selected |
|--------|-------------|----------|
| BERTimbau fine-tuned | BERT pt-BR, melhor acurácia | ✓ |
| Multilingual BERT | mBERT, performance inferior | |
| HuggingFace pipeline | Mais rápido, menor | |
| HuggingFace Inference API | Sem modelo local | |

**User's choice:** BERTimbau fine-tuned
**Notes:** Fine-tuning para 10-15 emoções do português. Dataset GoEmotions-PT + expansão com dados reais.

---

## Gestão do Modelo

| Option | Description | Selected |
|--------|-------------|----------|
| Download no Docker build | Imagem maior, zero runtime | ✓ |
| Download na primeira execução | Imagem leve, primeira init lenta | |
| Volume montado | Cache entre deploys, script setup | |

**User's choice:** Download no Docker build

---

## Schema de Armazenamento

| Option | Description | Selected |
|--------|-------------|----------|
| Campo no registro existente | Simples, doc cresce | |
| Documento separado | Desacoplado, fácil re-analisar | ✓ |
| Ambos | Redundância para exibição rápida | |

**User's choice:** Documento separado (`analise:{registroId}`, type: `analise_nlp`)

---

## Enriquecimento

| Option | Description | Selected |
|--------|-------------|----------|
| Relatório PDF + Histórico | Emoções aparecem nos dois lugares | ✓ |
| Só no relatório | Sem mudanças de UI no frontend | |

**User's choice:** Relatório PDF + Histórico
**Notes:** Sub-fase 3 inclui mudanças no frontend para exibir emoções no RegistroCard

---

## the agent's Discretion

- Tamanho exato do modelo e estratégia de cache de inferência no Python
- Estratégia de polling vs long-poll no _changes feed
- Detalhes de UI para exibição das emoções no RegistroCard
- Dependências Python exatas (grpcio, transformers, torch, etc.)

## Deferred Ideas

- **WhatsApp integration** — Próxima feature v2 após NLP
- **Análise em tempo real (streaming)** — Evolução futura do gRPC
- **Modelo fine-tuned por perfil de usuário** — Personalização futura
