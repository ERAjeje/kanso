---
phase: 10-sentiment-training
type: execute
depends_on: [07-03-integracao]
requirements: [NLP-01, NLP-02, NLP-03, REG-01]
files_modified:
  - frontend/src/types/index.ts
  - frontend/src/services/registros.ts
  - frontend/src/services/pouchdb.ts
  - frontend/src/components/RegistroCard.tsx
  - frontend/src/pages/History.tsx
  - backend/internal/repository/couchdb.go
  - backend/internal/service/watcher.go
  - backend/cmd/kanso-api/main.go
  - nlp-service/src/model_config.py
  - nlp-service/train_model.py
  - nlp-service/src/server.py
autonomous: false
must_haves:
  truths:
    - Sentiment edit is only available when sentimentoId is null/empty
    - Once sentiment is set (sentimentoId exists), the field is locked
    - Editing uses the 13 fixed emotions as the label vocabulary
    - Training data is stored in a CouchDB database as text + label pairs
    - Training only triggers when the training data has changed (content hash check)
    - After retraining, model version increments and re-analysis happens lazily
  artifacts:
    - path: frontend/src/components/SentimentoEditor.tsx
      provides: "Combobox for selecting from 13 emotions, disabled when sentiment is already set"
    - path: frontend/src/services/training.ts
      provides: "Service to save training example to treinamento DB"
    - path: backend/internal/service/treinamento.go
      provides: "Training data management, change detection, re-analysis trigger"
    - path: nlp-service/train_model.py
      provides: "Updated training script that reads from CouchDB treinamento DB"
---

<objective>
Permitir que o usuário edite o sentimento de um registro diretamente na listagem (History), usando as 13 emoções do modelo NLP como vocabulário. A edição só fica disponível quando o sentimento não foi preenchido (`sentimentoId === null`). Ao salvar, o par `(texto, label)` é armazenado como dado de treinamento. O retreinamento do modelo é feito em batch (manual ou scheduler), apenas quando há dados novos. Após o retreino, a re-análise dos registros é lazy.
</objective>

<execution_context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/07-03-integracao/07-03-CONTEXT.md
@.planning/phases/v2-nlp-analysis/v2-nlp-CONTEXT.md
</execution_context>

<context>
## Decisões da Discussão

### D-01: Vocabulário de 13 emoções
A edição usa exclusivamente as 13 emoções do modelo NLP como labels (alegria, tristeza, raiva, medo, nojo, surpresa, ansiedade, vergonha, culpa, saudade, amor, gratidão, neutro). O `SentimentoCombobox` genérico não é usado — cria-se um seletor específico com essas opções.

### D-02: Batch training com detecção de mudança
O retreinamento é acionado manualmente (ou por scheduler). Um hash do conteúdo da base de treinamento é comparado com o último hash conhecido. Só retreina se houver diferença.

### D-03: Re-análise lazy
Após retreinar, a versão do modelo (`modeloVersao`) é incrementada. Registros com versão desatualizada são re-analisados sob demanda (lazy), não em massa imediata.

### D-04: Database `treinamento` no CouchDB
Novo database CouchDB `treinamento` contém a base de treinamento completa:
- Frases curadas (migradas do `curated_phrases.py`)
- Exemplos do usuário (pares `texto + label`)
- GoEmotions PT-BR continua sendo carregado do HuggingFace (versionado por referência)

### D-05: Edição no History
No `RegistroCard`, quando `sentimentoId === null`, exibe `SentimentoEditor` (combobox). Quando preenchido, exibe texto simples/disabled. Edição no modo expandido do card.

### D-06: Sincronização offline
Usuário pode editar sentimento offline via PouchDB. O training example também é salvo no PouchDB `treinamento` e sincronizado quando online.
</context>

<waves>

<wave number="1" name="Training Database & Pipeline">
  <task>
    <name>Task 1.1: Adicionar database treinamento no backend + frontend</name>
    <files>
      - backend/internal/repository/couchdb.go (const DBTreinamento)
      - frontend/src/services/pouchdb.ts (createSyncedDB('treinamento'))
    </files>
    <action>
      **Backend:**
      - Adicionar constant `DBTreinamento = "treinamento"` em `couchdb.go`
      - Garantir que `validate_doc_update` no CouchDB permita escrita para este DB

      **Frontend:**
      - Adicionar `export const { local: treinamentoDB } = createSyncedDB('treinamento')` em `pouchdb.ts`
    </action>
    <verify>
      - `grep -c "DBTreinamento" backend/internal/repository/couchdb.go` retorna 1
      - `grep -c "treinamentoDB" frontend/src/services/pouchdb.ts` retorna 1
      - `docker compose build` passa
    </verify>
  </task>

  <task>
    <name>Task 1.2: Criar tipo TreinamentoDoc e service training.ts</name>
    <files>
      - frontend/src/types/index.ts
      - frontend/src/services/training.ts (novo)
    </files>
    <action>
      **Type (types/index.ts):**
      ```typescript
      export interface TreinamentoDoc {
        _id: string
        _rev?: string
        type: 'treinamento'
        texto: string
        label: string  // uma das 13 emoções
        userId: string
        origem: 'usuario' | 'curada'
        criadoEm: string
      }
      ```

      **Service (training.ts):**
      ```typescript
      import { treinamentoDB, getUserId } from './pouchdb'
      import type { TreinamentoDoc } from '../types'

      export async function saveTrainingExample(texto: string, label: string): Promise<TreinamentoDoc> {
        const doc: TreinamentoDoc = {
          _id: crypto.randomUUID(),
          type: 'treinamento',
          texto,
          label,
          userId: getUserId(),
          origem: 'usuario',
          criadoEm: new Date().toISOString(),
        }
        await treinamentoDB.put(doc)
        return doc
      }

      export async function getTotalTrainingCount(): Promise<number> {
        const result = await treinamentoDB.allDocs({ limit: 0 })
        return result.total_rows
      }
      ```
    </action>
    <verify>
      - TypeScript compila sem erros
      - Testes básicos da função passam
    </verify>
  </task>

  <task>
    <name>Task 1.3: Migrar curated_phrases.py para CouchDB treinamento</name>
    <files>
      - nlp-service/data/migrate_curated.py (novo script migratório)
    </files>
    <action>
      Script one-shot que:
      1. Conecta ao CouchDB `treinamento`
      2. Para cada frase em `curated_phrases.py`, converte o multi-hot vector para label única (a posição com valor 1)
      3. Cria documento `TreinamentoDoc` com `origem: 'curada'`
      4. Ignora se já existir (idempotente — baseado no hash da frase)

      ```python
      # migrate_curated.py
      # Lê CURATED_PHRASES de curated_phrases.py
      # Converte multi-hot → label string via LABELS index
      # PUT em CouchDB treinamento DB
      ```
    </action>
    <verify>
      - Script executa sem erros
      - Documentos criados no CouchDB `treinamento` com `type: 'treinamento'` e `origem: 'curada'`
    </verify>
  </task>

  <task>
    <name>Task 1.4: Criar backend service treinamento.go (change detection)</name>
    <files>
      - backend/internal/service/treinamento.go (novo)
      - backend/internal/repository/couchdb.go (métodos de treinamento)
    </files>
    <action>
      **Repository methods (couchdb.go):**
      ```go
      type TreinamentoDoc struct {
          ID       string `json:"_id,omitempty"`
          Rev      string `json:"_rev,omitempty"`
          Type     string `json:"type"`
          Texto    string `json:"texto"`
          Label    string `json:"label"`
          UserSub  string `json:"userSub"`
          Origem   string `json:"origem"` // "usuario" | "curada"
          CriadoEm string `json:"criadoEm"`
      }

      type TrainingCheckpointDoc struct {
          ID           string `json:"_id,omitempty"`
          Rev          string `json:"_rev,omitempty"`
          Type         string `json:"type"`
          ContentHash  string `json:"contentHash"`
          ModelVersion string `json:"modelVersion"`
          UpdatedAt    string `json:"updatedAt"`
      }

      func (c *CouchDB) GetTrainingCheckpoint() (*TrainingCheckpointDoc, error)
      func (c *CouchDB) SaveTrainingCheckpoint(hash, version string) error
      func (c *CouchDB) GetTrainingData() ([]TreinamentoDoc, error)
      func (c *CouchDB) ComputeTrainingHash() (string, error)
      func (c *CouchDB) HasTrainingChanged() (bool, string, error)
      ```

      **Service (treinamento.go):**
      ```go
      type TreinamentoService struct {
          repo     *repository.CouchDB
          nlpClient nlp.Analyzer
          cfg      *config.Config
      }

      // CheckAndTrain verifica se a base mudou e dispara treinamento
      func (s *TreinamentoService) CheckAndTrain(ctx context.Context) error
      // GetCurrentModelVersion retorna a versão atual do modelo
      func (s *TreinamentoService) GetCurrentModelVersion() string
      // ReanalyzeRegistros re-analisa registros com versão desatualizada
      func (s *TreinamentoService) ReanalyzeRegistros(ctx context.Context) error
      ```

      **Detecção de mudança:** `ComputeTrainingHash` calcula SHA256 do JSON concatenado de todos os docs `treinamento` com `origem: 'usuario'`. Compara com `TrainingCheckpointDoc.ContentHash`. Só retreina se diferente.
    </action>
    <verify>
      - `go build ./internal/service/` compila sem erros
      - `go vet ./internal/service/` passa
    </verify>
  </task>

  <task>
    <name>Task 1.5: Atualizar train_model.py para ler do CouchDB</name>
    <files>
      - nlp-service/train_model.py
      - nlp-service/src/model_config.py
    </files>
    <action>
      **train_model.py:**
      Adicionar função `load_training_from_couchdb()` que:
      1. Conecta ao CouchDB (URL via env `COUCHDB_URL`)
      2. Busca todos os docs `type: 'treinamento'`
      3. Converte `label` (string) para multi-hot vector (array de 13 posições)
      4. Retorna Dataset no mesmo formato que `load_goemotions_dataset()` e `load_curated_phrases()`

      Modificar pipeline de treinamento:
      ```python
      logger.info("Step 2b: Loading training data from CouchDB")
      couchdb_ds = load_training_from_couchdb()
      if couchdb_ds is not None and len(couchdb_ds) > 0:
          logger.info("CouchDB training data loaded: %d rows", len(couchdb_ds))
          combined = concatenate_datasets([combined, couchdb_ds])
      ```

      **model_config.py:**
      Adicionar:
      ```python
      COUCHDB_URL = os.environ.get("COUCHDB_URL", "http://admin:password@couchdb:5984")
      COUCHDB_TREINAMENTO_DB = "treinamento"
      ```
    </action>
    <verify>
      - `cd nlp-service && python -c "from train_model import load_training_from_couchdb; print('OK')"` não lança erro
      - Testes existentes do treinamento continuam passando
    </verify>
  </task>
</wave>

<wave number="2" name="Frontend — Sentiment Editor no History">
  <task>
    <name>Task 2.1: Criar componente SentimentoEditor</name>
    <files>
      - frontend/src/components/SentimentoEditor.tsx (novo)
    </files>
    <action>
      Componente que renderiza um combobox com as 13 emoções do modelo.

      ```typescript
      interface SentimentoEditorProps {
        currentValue: string
        disabled: boolean
        onSave: (label: string) => Promise<void>
      }
      ```

      - Usa `@headlessui/react` Combobox (mesmo padrão do `SentimentoCombobox`)
      - Lista fixa: as 13 emoções de `model_config.py`
      - Quando `disabled === true`, exibe texto simples com a cor da emoção (reusa `EMOTION_CHIP_COLORS`)
      - Quando `disabled === false`, exibe o combobox com placeholder "Selecionar sentimento"
      - Emoções com acentos: alegria, tristeza, raiva, medo, nojo, surpresa, ansiedade, vergonha, culpa, saudade, amor, gratidão, neutro
      - Exibe as emoções ordenadas alfabeticamente
    </action>
    <verify>
      - Teste unitário: renderiza 13 opções
      - Teste unitário: disabled mode mostra texto, não combobox
    </verify>
  </task>

  <task>
    <name>Task 2.2: Atualizar RegistroCard com edição de sentimento</name>
    <files>
      - frontend/src/components/RegistroCard.tsx
      - frontend/src/services/registros.ts
    </files>
    <action>
      **registros.ts — novo método:**
      ```typescript
      export async function updateRegistroSentimento(
        registro: RegistroDoc,
        sentimentoNome: string,
        sentimentoId: string
      ): Promise<RegistroDoc> {
        const doc = await registrosDB.get(registro._id)
        const updated: RegistroDoc = {
          ...doc,
          sentimentoNome,
          sentimentoId,
          updatedAt: new Date().toISOString(),
        }
        await registrosDB.put(updated)
        return updated
      }
      ```

      **RegistroCard.tsx — alterações:**
      - Adicionar estado `isEditing` (boolean)
      - Quando `expanded === true` e `sentimentoId === null`: exibir `SentimentoEditor` no lugar do campo "Sentimento"
      - Quando `expanded === false` ou `sentimentoId` preenchido: comportamento atual (texto estático)
      - No `onSave` do editor:
        1. Chama `updateRegistroSentimento()` com o registro atual
        2. Chama `saveTrainingExample()` com o texto combinado (sensacoes + contexto + pensamentos) e a label
        3. Atualiza o estado local para refletir o novo sentimento
        4. Exibe toast de confirmação (via `onShowToast` prop)

      **Importante:** O card precisa receber uma callback `onSentimentoUpdated?: () => void` para que o History.tsx possa recarregar a lista.
    </action>
    <verify>
      - Card sem `sentimentoId`: mostra editor quando expandido
      - Card com `sentimentoId`: mostra texto estático sempre
      - Ao salvar: registro é atualizado no PouchDB e training example é criado
    </verify>
  </task>

  <task>
    <name>Task 2.3: Atualizar History.tsx com refresh pós-edição</name>
    <files>
      - frontend/src/pages/History.tsx
    </files>
    <action>
      - Adicionar callback `onSentimentoUpdated` que recarrega `getRegistros()`
      - Passar a callback para `RegistroCard` via prop
    </action>
    <verify>
      - Após editar sentimento, o card atualiza imediatamente na listagem
    </verify>
  </task>
</wave>

<wave number="3" name="Backend — Treinamento e Re-análise Lazy">
  <task>
    <name>Task 3.1: Endpoint para trigger de treinamento</name>
    <files>
      - backend/cmd/kanso-api/main.go
      - backend/internal/handler/treinamento.go (novo)
    </files>
    <action>
      **Novo handler:**
      ```go
      // POST /api/train — verifica mudanças e dispara treinamento se necessário
      func (h *TreinamentoHandler) HandleTrain(w http.ResponseWriter, r *http.Request)
      // GET /api/train/status — status do último treinamento
      func (h *TreinamentoHandler) HandleTrainStatus(w http.ResponseWriter, r *http.Request)
      // POST /api/reanalyze — re-analisa registros com versão desatualizada
      func (h *TreinamentoHandler) HandleReanalyze(w http.ResponseWriter, r *http.Request)
      ```

      **TreinamentoHandler.HandleTrain:**
      1. Verifica `HasTrainingChanged()` no repositório
      2. Se mudou: dispara goroutine que:
         a. Chama NLP service endpoint `/train` (novo endpoint HTTP no nlp-service)
         b. Aguarda conclusão
         c. Atualiza `TrainingCheckpointDoc` com novo hash e versão
         d. Incrementa `modeloVersao`
      3. Se não mudou: retorna 200 com `{ trained: false, reason: "no_changes" }`

      **Roteamento em main.go:**
      ```go
      r.Post("/api/train", treinamentoHandler.HandleTrain)
      r.Get("/api/train/status", treinamentoHandler.HandleTrainStatus)
      r.Post("/api/reanalyze", treinamentoHandler.HandleReanalyze)
      ```
    </action>
    <verify>
      - `go build ./cmd/kanso-api` compila sem erros
      - `go vet ./cmd/kanso-api` passa
      - Teste: POST /api/train retorna 200
    </verify>
  </task>

  <task>
    <name>Task 3.2: Endpoint de treinamento no NLP service</name>
    <files>
      - nlp-service/src/server.py
      - nlp-service/src/health.py
    </files>
    <action>
      Adicionar endpoint HTTP no FastAPI (health.py) ou novo arquivo:

      ```python
      # POST /train — executa treinamento
      @app.post("/train")
      async def handle_train():
          # 1. Verifica se treinamento já está em execução (mutex)
          # 2. Chama train() do train_model.py em subprocess ou thread
          # 3. Retorna { status, model_version, trained_count }
          pass

      # GET /model/version — versão atual do modelo
      @app.get("/model/version")
      async def model_version():
          return {"version": MODEL_VERSION}
      ```

      **Proteção:** Usar `threading.Lock` para evitar treinamento concorrente.
    </action>
    <verify>
      - `curl http://nlp:8000/model/version` retorna versão
      - POST /train executa sem erro
    </verify>
  </task>

  <task>
    <name>Task 3.3: Re-análise lazy no watcher + service</name>
    <files>
      - backend/internal/service/watcher.go
      - backend/internal/service/treinamento.go
    </files>
    <action>
      **WatcherService modificação:**
      - Ao detectar análise desatualizada (versão do modelo não corresponde), marca para re-análise
      - Re-análise é feita sob demanda via `ReanalyzeRegistros()`

      **TreinamentoService.ReanalyzeRegistros:**
      1. Busca todos `RegistroDoc` do usuário
      2. Para cada um com `analise_nlp.modeloVersao !== currentVersion`:
         a. Envia para gRPC Analyze
         b. Atualiza `AnaliseNlpDoc` com nova versão
      3. Rate limit de 50ms entre chamadas

      **Integração:** Quando frontend chama `POST /api/reanalyze`, o handler dispara re-análise em background.
    </action>
    <verify>
      - Re-análise executa sem bloquear a API
      - `slog.Info` registra progresso
    </verify>
  </task>
</wave>

<wave number="4" name="Infra & Scheduler">
  <task>
    <name>Task 4.1: Configurar CouchDB database treinamento</name>
    <files>
      - backend/cmd/kanso-api/main.go
    </files>
    <action>
      No startup da API, garantir que o database `treinamento` existe (mesmo padrão dos outros 5 DBs).
      Adicionar `ensureDatabase(couchClient, "treinamento")` no bloco de inicialização.
      Adicionar `validate_doc_update` para o database `treinamento` (mesmo padrão de isolamento por userId).
    </action>
    <verify>
      - Container sobe sem erro
      - `curl couchdb:5984/_all_dbs` inclui `treinamento`
    </verify>
  </task>

  <task>
    <name>Task 4.2: Scheduler semanal (opcional)</name>
    <files>
      - backend/internal/service/treinamento.go
    </files>
    <action>
      Opcional — adicionar ticker no `TreinamentoService` para verificar mudanças periodicamente:
      ```go
      func (s *TreinamentoService) StartScheduler(ctx context.Context, interval time.Duration) {
          ticker := time.NewTicker(interval)
          go func() {
              for {
                  select {
                  case <-ticker.C:
                      s.CheckAndTrain(ctx)
                  case <-ctx.Done():
                      ticker.Stop()
                      return
                  }
              }
          }()
      }
      ```
      Intervalo configurável via env `TRAIN_INTERVAL` (default: 7 dias = `168h`).
    </action>
    <verify>
      - Scheduler compila
      - Não dispara treinamento se base não mudou
    </verify>
  </task>
</wave>

</waves>

<data_flow>
## Fluxo Completo

### Edição de Sentimento (User Flow)
```
[History.tsx] → RegistroCard (sentimentoId === null)
  → Usuário clica para expandir o card
  → SentimentoEditor aparece no campo "Sentimento"
  → Usuário seleciona uma das 13 emoções
  → onSave dispara:
      1. updateRegistroSentimento(registro._id, "ansiedade", "label-ansiedade")
         → PouchDB.put() → sync CouchDB
      2. saveTrainingExample("texto combinado", "ansiedade")
         → treinamentoDB.put() → sync CouchDB
      3. Card atualiza estado local → exibe sentimento fixo
```

### Treinamento (System Flow)
```
[Usuário/Admin] → POST /api/train
  → TreinamentoService.CheckAndTrain()
    → HasTrainingChanged()?
      → NO → 200 { trained: false, reason: "no_changes" }
      → YES → POST http://nlp:8000/train
        → NLP Service executa train_model.py:
          1. load_goemotions_dataset() ← HuggingFace
          2. load_training_from_couchdb() ← CouchDB treinamento
          3. Combina datasets
          4. Fine-tune BERTimbau
          5. Salva modelo em /model
          6. Incrementa MODEL_VERSION
        → TreinamentoService atualiza checkpoint (hash + versão)
        → 200 { trained: true, model_version: "v1.1" }
```

### Re-análise Lazy (System Flow)
```
[Frontend] → POST /api/reanalyze
  → TreinamentoService.ReanalyzeRegistros()
    → Busca registros com modeloVersao != currentVersion
    → Para cada: gRPC Analyze() → atualiza AnaliseNlpDoc
    → 200 { reanalyzed: N, total: M }

[Ou sob demanda no watcher]
  WatcherService detecta registro com versão desatualizada
  → Marca para re-análise
  → Processa quando chamar ReanalyzeRegistros()
```
</data_flow>

<timeline>
## Esforço Estimado

| Wave | Tarefas | Estimativa |
|------|---------|------------|
| 1 — Training Database & Pipeline | 5 tasks | ~3h |
| 2 — Frontend Editor | 3 tasks | ~2h |
| 3 — Backend Treinamento & Re-análise | 3 tasks | ~2h |
| 4 — Infra & Scheduler | 2 tasks | ~1h |
| **Total** | **13 tasks** | **~8h** |
</timeline>

<out_of_scope>
- Interface de administração para treinamento (dashboard UI)
- Backfill de registros existentes (re-análise é lazy, não backfill)
- Suporte a múltiplos labels por texto (multilabel — o modelo suporta, mas a UI será single-label)
- Notificação push quando treinamento concluir
- Versionamento de múltiplos modelos
</out_of_scope>
