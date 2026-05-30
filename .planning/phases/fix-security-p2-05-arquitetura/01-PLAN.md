---
phase: fix-security-p2-05-arquitetura
plan: 01
type: decision
wave: 3
depends_on:
  - fix-security-p2-01-credenciais
  - fix-security-p2-02-correcoes-rapidas
  - fix-security-p0
  - fix-security-p1-01
files_modified: []
autonomous: false
requirements:
  - HI-01: Decidir se Docker socket no Traefik deve ser mitigado
  - HI-02: Decidir se gRPC sem TLS deve ser mitigado
  - ME-01: Decidir se Vite proxy /db deve ser removido
  - ME-05: Decidir se JWT deve migrar de localStorage para HttpOnly cookie
must_haves:
  truths:
    - Decisão documentada para cada item com rationale
    - Se fix, tasks de implementação definidas
    - Se accept, risco documentado em threat model
  artifacts:
    - path: this file
      provides: "Decisões arquiteturais documentadas"
      contains: "decisão + rationale"
---

<objective>
Discutir e decidir o tratamento de 4 itens arquiteturais de segurança: Docker socket (HI-01), gRPC TLS (HI-02), Vite proxy /db (ME-01), JWT localStorage (ME-05).
</objective>

<execution_context>
@infra/docker-compose.yml
@infra/traefik/traefik.yml
@backend/internal/nlp/client.go
@backend/cmd/kanso-api/main.go
@frontend/vite.config.ts
@frontend/src/services/auth.ts
</execution_context>

---

## Nota

Este é um plano de **decisão** (type: decision), não de execução. O usuário deve revisar as opções e escolher o tratamento. Após a decisão, as tasks de implementação (se houver) serão executadas em planos separados.

---

## Item 1: HI-01 — Docker Socket no Traefik

**Problema:** `/var/run/docker.sock:/var/run/docker.sock:ro` montado read-only no Traefik. Se Traefik for comprometido, atacante consegue inspecionar containers e ler env vars com secrets.

**Opções:**

| Opção | Descrição | Esforço | Impacto |
|-------|-----------|---------|---------|
| A — Aceitar risco | Documentar que Docker socket é necessário para Docker provider do Traefik. Risco mitigado por ser read-only. | 0 | Nenhum |
| B — Migrar para file provider | Substituir Docker provider por configuração 100% file-based no dynamic.yml. Remove necessidade de docker.sock. | 1-2h | Remove auto-descoberta de containers — toda rota deve ser declarada manualmente |
| C — Restringir acesso via auth plugin | Usar plugin de autorização Docker. Complexo. | 4h+ | Alto esforço para ganho marginal |

**Recomendação:** Opção A (aceitar) — o mount é read-only e o Traefik já está em produção. A migração para file provider pode ser feita como melhoria futura.

---

## Item 2: HI-02 — gRPC sem TLS

**Problema:** gRPC entre Go backend e NLP service usa `insecure.NewCredentials()` — dados emocionais trafegam sem criptografia na rede interna Docker.

**Opções:**

| Opção | Descrição | Esforço | Impacto |
|-------|-----------|---------|---------|
| A — Aceitar risco | Rede interna Docker é isolada. Apenas containers autorizados têm acesso. | 0 | Risco baixo — atacante precisaria de acesso à rede Docker |
| B — Adicionar TLS auto-assinado | Gerar certificado, configurar server e client gRPC com TLS. | 2-3h | Criptografia em trânsito + autenticação mútua |
| C — mTLS | Mutual TLS — backend e NLP se autenticam mutuamente. | 3-4h | Máxima segurança interna |

**Recomendação:** Opção A (aceitar) — rede Docker interna é considerada segura. Pode ser revisitado se NLP for exposto externamente.

---

## Item 3: ME-01 — Vite Proxy /db

**Problema:** Vite dev proxy em `vite.config.ts` expõe CouchDB diretamente ao frontend via `/db`. Usuários podem fazer chamadas CouchDB diretas bypassando API Go.

**Contexto atual:** Com `validate_doc_update` implantado (P0), cada usuário só consegue escrever documentos onde `doc.userSub === userCtx.name`. O risco de leitura cruzada permanece via Mango queries sem filtro de userSub — mas nenhuma query no frontend faz isso.

**Opções:**

| Opção | Descrição | Esforço | Impacto |
|-------|-----------|---------|---------|
| A — Aceitar com mitigação | Manter /db proxy. validate_doc_update impede escrita cruzada. Risco de leitura cruzada é baixo porque PouchDB sincroniza apenas docs do próprio usuário. | 0 | Perde defesa em profundidade |
| B — Remover proxy, rotear via API | Todo acesso a dados passa pela API Go. PouchDB sync precisa ser refeito. Perde offline-first. | 3-4h | Quebra offline-first — P2 precisa do PouchDB |
| C — Remover apenas em produção | Em dev, manter para desenvolvimento. Em produção, Traefik pode bloquear /db. | 30min | Melhor compromisso |

**Recomendação:** Opção C — adicionar regra no Traefik para bloquear `/db` externamente em produção. Em dev, manter para facilitar desenvolvimento.

---

## Item 4: ME-05 — JWT em localStorage

**Problema:** JWT armazenado em `localStorage` — qualquer XSS pode roubar o token. HttpOnly cookie seria mais seguro.

**Opções:**

| Opção | Descrição | Esforço | Impacto |
|-------|-----------|---------|---------|
| A — Aceitar risco | Sem XSS conhecido. localStorage é padrão na indústria para SPAs com JWT. | 0 | Risco teórico |
| B — Migrar para HttpOnly cookie | Backend seta cookie com Set-Cookie (HttpOnly, Secure, SameSite). Frontend remove localStorage. | 3-4h | Elimina risco de roubo via XSS |
| C — Curto prazo: aceitar. Longo prazo: migrar. | Documentar como tech debt para próxima versão. | 0 | Adia a decisão |

**Recomendação:** Opção C — adiar para v4. A migração para HttpOnly cookie exige mudanças no fluxo de auth (backend precisa setar cookie, frontend precisa lidar com cookie em vez de header Authorization) e não há XSS conhecido no código.

---

## Decisões (a preencher pelo usuário)

| Item | Decisão | Responsável |
|------|---------|-------------|
| HI-01 — Docker socket | [ ] Aceitar / [ ] Fix: file provider | Usuário |
| HI-02 — gRPC sem TLS | [ ] Aceitar / [ ] Fix: TLS | Usuário |
| ME-01 — Vite proxy /db | [ ] Aceitar / [ ] Fix: Traefik block / [ ] Fix: remover | Usuário |
| ME-05 — JWT localStorage | [ ] Aceitar / [ ] Fix: HttpOnly cookie / [ ] Adiar | Usuário |

---

## Threat Model

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-HI-01-01 | Privilege Escalation | Traefik | accept (pending decision) | Docker socket read-only |
| T-HI-02-01 | Eavesdropping | gRPC | accept (pending decision) | Rede Docker interna isolada |
| T-ME-01-01 | Unauthorized Access | CouchDB proxy | mitigate | validate_doc_update + Traefik block |
| T-ME-05-01 | XSS Token Theft | localStorage | accept (pending decision) | Sem XSS conhecido no código |

## Verification

- Decisões registradas neste documento
- Se fix, plano de execução separado criado
- Se accept, risco documentado para auditoria futura
