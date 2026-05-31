---
phase: fix-chromedp-healthcheck
plan: 01
type: execute
wave: 1
depends_on:
  - fix-chromedp-separation-01
files_modified:
  - infra/docker-compose.yml
  - infra/chromedp/Dockerfile
autonomous: false
requirements:
  - "BUG-006: Container kanso-chromedp unhealthy — healthcheck usa wget que não existe na imagem"
must_haves:
  truths:
    - healthcheck do chromedp usa comando disponível na imagem base
    - chromedp/headless-shell:latest gerencia headless-shell corretamente sem CMD自定义
    - healthcheck passa consistentemente (exit 0)
    - Container é marcado como healthy
  artifacts:
    - path: infra/docker-compose.yml
      provides: "Healthcheck compatível com imagem base"
      not_contains: "wget"
    - path: infra/chromedp/Dockerfile
      provides: "Dockerfile minimal — sem CMD redundante"
      not_contains: "--remote-debugging-port"
---

<objective>
Corrigir healthcheck do container kanso-chromedp que falha porque `wget` não está disponível na imagem base `chromedp/headless-shell:latest`.

**Causa raiz dupla:**

1. **Healthcheck inválido:** Usa `wget` — binário inexistente na imagem mínima `chromedp/headless-shell`
2. **CMD redundante:** O Dockerfile define `CMD` que conflita com o entrypoint `run.sh` da imagem base. A imagem já gerencia headless-shell + socat. O CMD extra causa `bind() failed: Address already in use (98)`
</objective>

<execution_context>
@infra/docker-compose.yml
@infra/chromedp/Dockerfile
</execution_context>

<context>
**Problema 1 — Healthcheck:**
```yaml
# Atual — FALHA: wget não existe
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:9222/json/version"]
```

**Problema 2 — CMD no Dockerfile:**
```dockerfile
FROM chromedp/headless-shell:latest
EXPOSE 9222
CMD ["/headless-shell/headless-shell", "--remote-debugging-port=9222", ...]
```

A imagem `chromedp/headless-shell:latest` tem entrypoint próprio (`run.sh`) que:
1. Executa headless-shell com `--remote-debugging-port=9223`
2. Inicia `socat` para forward 9222 → 9223

Nosso `CMD` é passado como argumento extra ao entrypoint, fazendo headless-shell receber 2x `--remote-debugging-port` (9223 + 9222). O último vence (9222), mas o socat tenta bind na mesma porta e falha com `Address already in use`.

**Solução:** Remover o `CMD` do Dockerfile (deixar a imagem funcionar como foi projetada) e trocar healthcheck para usar `socat` (disponível na imagem).

**Evidência empírica:**
```bash
# wget não existe
$ docker exec kanso-chromedp wget → exec: "wget": file not found

# socat existe e porta 9222 responde
$ docker exec kanso-chromedp socat /dev/null TCP4:localhost:9222,connect-timeout=3
→ exit 0
```
</context>

<tasks>

<task type="auto" tdd="false">
  <name>Task 1 — Corrigir Dockerfile: remover CMD redundante</name>
  <files>infra/chromedp/Dockerfile</files>
  <action>
    Substituir conteúdo de `infra/chromedp/Dockerfile` para:
    ```dockerfile
    FROM chromedp/headless-shell:latest
    EXPOSE 9222
    ```
    
    A imagem base já contém entrypoint + CMD padrão que gerencia headless-shell corretamente (porta 9223 + socat forward 9222).
  </action>
  <verify>
    <automated>grep -c "CMD" infra/chromedp/Dockerfile</automated>
    <expected>0</expected>
    <automated>grep -c "EXPOSE" infra/chromedp/Dockerfile</automated>
    <expected>1</expected>
  </verify>
</task>

<task type="auto" tdd="false">
  <name>Task 2 — Corrigir healthcheck no docker-compose.yml</name>
  <files>infra/docker-compose.yml</files>
  <action>
    Substituir o healthcheck do service `chromedp`:
    ```diff
    -      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9222/json/version"]
    +      test: ["CMD-SHELL", "socat /dev/null TCP4:localhost:9222,connect-timeout=3"]
    ```
    
    Usar `CMD-SHELL` porque `socat` aceita argumentos como string única. O `connect-timeout=3` garante que o healthcheck não trave se a porta não responder em 3s.
  </action>
  <verify>
    <automated>grep -c "wget" infra/docker-compose.yml</automated>
    <expected>0</expected>
    <automated>grep -c "socat" infra/docker-compose.yml</automated>
    <expected>1</expected>
  </verify>
</task>

</tasks>

<verification>

### Build
- `docker compose build chromedp` — build sem erros
- `docker compose config` — mostra healthcheck com socat

### Runtime
- `docker compose up -d chromedp` — container sobe
- `docker ps --filter name=kanso-chromedp` — mostra `healthy` (não `unhealthy`)
- `docker exec kanso-chromedp socat /dev/null TCP4:localhost:9222,connect-timeout=3` — exit 0

### Full stack
- `make up` — todos os containers saudáveis
- Healthcheck do chromedp passa em < 15s

</verification>

<success_criteria>

1. Container kanso-chromedp fica `healthy` após o startup
2. `make up` completa sem erros
3. Nenhum `wget` referenciado no docker-compose.yml
4. Dockerfile limpo — sem CMD redundante que cause conflito de porta
5. Porta 9222 acessível via socat dentro do container

</success_criteria>
