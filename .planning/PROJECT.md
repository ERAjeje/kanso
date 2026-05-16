# Kanso

## What This Is

Kanso é um aplicativo mobile (PWA) de diário emocional offline-first para auxiliar o processo terapêutico. O usuário registra sensações, sentimentos, contexto e pensamentos no momento em que são percebidos — com suporte a registro retroativo via notificações push — e pode gerar relatórios periódicos que são enviados via WhatsApp para a psicóloga. O app usa análise de NLP para extrair padrões emocionais dos registros ao longo do tempo.

## Core Value

O usuário consegue nomear e registrar suas emoções no momento em que as sente, criando um histórico que torna o processo terapêutico mais concreto e orientado a dados.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] **Registro de sensação**: Usuário pode registrar data/hora (com suporte retroativo), sensações (texto livre), sentimento (combobox customizável), contexto (texto livre) e pensamentos (texto livre)
- [ ] **Sincronização offline-first**: Dados locais (PouchDB) sincronizam com servidor (CouchDB) automaticamente quando online
- [ ] **Autenticação via Google**: Login com Google OAuth, JWT próprio para autorização
- [ ] **Navegação por abas**: Três abas — Registrar (padrão), Histórico, Perfil/Config
- [ ] **Notificações push**: Lembretes em horários customizáveis (padrão 12, 18, 23h) para registrar
- [ ] **Geração de relatório PDF**: Usuário gera relatório manualmente cobrindo período desde o último relatório até a data atual
- [ ] **Envio via WhatsApp**: Relatório PDF enviado automaticamente para a psicóloga via Twilio
- [ ] **Análise NLP**: Backend analisa registros com modelo de emoções em português e enriquece o registro com emoções detectadas
- [ ] **Histórico de registros**: Visualização cronológica dos registros salvos

### Out of Scope

- App nativo (iOS/Android) — PWA suficiente para MVP e portfólio
- Múltiplos terapeutas — apenas uma psicóloga por usuário
- Chat em tempo real com terapeuta — apenas relatório periódico via WhatsApp
- Registro de usuários sem Google Sign-In — OAuth já prepara para escalabilidade

## Context

O nome Kanso vem do japonês (簡素), significando simplicidade, clareza e essencialidade — com trocadilho linguístico com "cansar/cansado". O app nasce de uma necessidade pessoal do desenvolvedor de complementar a terapia com registros emocionais estruturados.

A abordagem de "nomear o sentimento" é uma técnica terapêutica reconhecida — o ato de identificar e rotular emoções ajuda a processá-las. O campo de sentimento é intencionalmente customizável porque o nome do sentimento pode ser descoberto durante a terapia com auxílio da psicóloga.

MVP incremental planejado: registro manual + sincronização → relatórios → NLP → push notifications → WhatsApp.

## Constraints

- **Tech stack**: React (Vite, TypeScript, Tailwind, PouchDB) / Go (chi, JWT, FCM, Twilio) / Python (FastAPI, transformers) / CouchDB / Traefik / Docker
- **Plataforma**: PWA mobile-first (inicialmente apenas o desenvolvedor como usuário)
- **Arquitetura**: Offline-first com PouchDB ↔ CouchDB, docker-compose para orquestração
- **Deploy**: VPS com Traefik como gateway
- **Código aberto**: Fonte disponível no portfólio — boas práticas e escalabilidade desde o início

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Google Sign-In desde v1 | Preparar para escalabilidade e portfólio público | — Pending |
| PWA em vez de app nativo | Suficiente para uso inicial, sem barreira de loja | — Pending |
| Offline-first com PouchDB/CouchDB | App móvel pode ficar sem conexão; sincronização nativa incremental | — Pending |
| Sentimento como combobox customizável | Permite descobrir e nomear sentimentos durante a terapia | — Pending |
| Relatório manual (não automático) | Usuário controla quando compartilhar com a psicóloga | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-16 after initialization*
