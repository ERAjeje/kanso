# Phase 07-02: Modelo — Context

**Gathered:** 2026-05-23
**Status:** Ready for planning

<domain>
## Phase Boundary

BERTimbau fine-tuning and emotion classification pipeline for 13 Portuguese emotion labels. Delivers the trained model, inference code, and model validation that replaces the current placeholder response in nlp-service. Sub-phase 2 of the NLP Analysis milestone — builds on the infra scaffolded in 07-01.

**Requirements addressed:** NLP-01, NLP-02, NLP-03 (model layer)

</domain>

<decisions>
## Implementation Decisions

### Fine-tuning Approach
- **D-18:** **Full fine-tune** BERTimbau (all layers) — not LoRA or adapter. Quality for nuanced Portuguese emotions (saudade vs tristeza) requires full fine-tuning.
- **D-19:** Training runs on **local dev machine** (GPU). Model artifact copied into Docker build context for deployment.
- **D-20:** Training script lives inside `nlp-service/` (`train_model.py`) — alongside serving code, not a separate repo.
- **D-21:** **GoEmotions-PT** downloaded at training time from HuggingFace (`datasets` library) — not checked into repo.

### Training Data & Augmentation
- **D-22:** **Manual Portuguese curation** — create ~200-500 culturally relevant Portuguese emotion phrases per label (including saudade, vergonha, culpa). Covers gaps from GoEmotions-PT's translated origin.
- **D-23:** **Weighted loss + oversampling** for class imbalance. Minority emotions (saudade, culpa, nojo) get higher loss weights AND more training examples via oversampling.
- **D-24:** **Back-translation Portuguese↔Spanish** (MarianMT) for data augmentation. Preserves emotion while generating diverse phrasing.

### Inference / Classification Strategy
- **D-25:** **Multi-label with threshold** — return all emotions with confidence >= 0.3 (default). Matches Portuguese emotional expression (mixed emotions are common).
- **D-26:** **Intensidade = max emotion score** (simplest, most interpretable).
- **D-27:** Short/uninformative input → classify as **neutro** with high confidence. No explicit length guard — model learns this naturally.

### Model Validation
- **D-28:** Primary metric: **Weighted F1 >= 0.70** on curated Portuguese test set.
- **D-29:** If target not met: **iterate with more data + hyperparameter tuning** (2-3 rounds), not fallback to simpler approach.
- **D-30:** Curated test set as **inline Python file** (`test_phrases.py`) in `nlp-service/tests/` — ~50 phrases per emotion, split into validation/test.

### Serving Integration
- **D-31:** Model loaded on **server startup** (module-level variable in `server.py`). First request may be slow (~5-10s model load), subsequent requests fast.
- **D-32:** **No batching** — each Analyze request processed individually. Single-user app doesn't need dynamic batching; can be added later if throughput requires.
- **D-33:** **CPU inference** — BERTimbau base (110M params) runs fine on CPU for single-request inference (~200-500ms). Simpler Docker setup without GPU passthrough.

### Model Versioning & Iteration
- **D-34:** **Git-tracked version** in `modelo_versao` field of AnalyzeResponse (git commit hash or semver). Updated on manual retraining.
- **D-35:** **Manual periodic retraining** — when enough real user data accumulates, export anonymized registrations, manually label a sample, add to training set, retrain.

### the agent's Discretion
- Exact hyperparameters (learning rate, epochs, batch size) for fine-tuning
- Specific MarianMT model for back-translation
- How to structure the Python classification module (class vs function, file layout within nlp-service/)
- Test file organization within test_phrases.py

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone & Requirements
- `.planning/REQUIREMENTS.md` §NLP — NLP-01, NLP-02, NLP-03 requirements
- `.planning/PROJECT.md` §Active — NLP feature description
- `.planning/phases/v2-nlp-analysis/v2-nlp-CONTEXT.md` — Full milestone decisions (D-01 through D-17)
- `.planning/ROADMAP.md` §Phase 7 — Phase structure and sub-phase boundaries

### Phase 07-01 (Infra — already built)
- `.planning/phases/07-01-infra-nlp/07-01-CONTEXT.md` — Existing infra decisions
- `.planning/phases/07-01-infra-nlp/07-01-PLAN.md` — What was built (gRPC server, Dockerfile, proto)
- `nlp-service/proto/analysis.proto` — gRPC interface (AnalyzeRequest / AnalyzeResponse schema)
- `nlp-service/src/server.py` — Existing gRPC server with placeholder response
- `nlp-service/download_model.py` — Model download at Docker build time (13 labels, BERTimbau)
- `nlp-service/Dockerfile` — Multi-stage build with model in image
- `nlp-service/requirements.txt` — Python dependencies

### Existing Patterns
- `infra/docker-compose.yml` — Service pattern for Docker services
- `backend/internal/config/config.go` — Env-based config pattern
- `nlp-service/src/__main__.py` — Entrypoint launching both gRPC + health HTTP

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **nlp-service/src/server.py** — Placeholder Analyze() method; replace with real classification
- **nlp-service/download_model.py** — Already downloads BERTimbau with 13 labels at Docker build time
- **nlp-service/proto/analysis.proto** — gRPC schema matches the model output structure
- **nlp-service/requirements.txt** — Has transformers, torch, grpcio — training dependencies are superset

### Established Patterns
- **Service isolation** — nlp-service runs as separate Docker container, internal network only
- **Model in image** — Model weights baked into Docker image at build time (not downloaded at runtime)
- **gRPC communication** — Go backend calls Python via gRPC on internal Docker network

### Integration Points
- **server.py Analyze()** — Replace placeholder with real model inference pipeline
- **download_model.py** — May need to update if training changes model format or adds preprocessing artifacts
- **Dockerfile** — Training happens outside Docker; only the trained model artifact is copied into the image

</code_context>

<specifics>
## Specific Ideas

- 13 emotion labels: alegria, tristeza, raiva, medo, nojo, surpresa, ansiedade, vergonha, culpa, saudade, amor, gratidão, neutro
- Curated Portuguese phrases should include culturally specific scenarios (e.g., "Sinto falta da minha avó" → saudade)
- BERTimbau base (neuralmind/bert-base-portuguese-cased) — the 110M parameter version, not the large variant
- Model version follows `v{major}.{minor}` manually bumped (starts at v1.0)

</specifics>

<deferred>
## Deferred Ideas

- **GPU inference in production** — Not needed for MVP; can add CUDA passthrough later if CPU latency becomes an issue
- **Active learning loop** — Automated flagging of low-confidence predictions for review; too complex for initial version
- **HuggingFace model registry** — Publishing fine-tuned model to HF Hub; adds external dependency, defer until model proves valuable
- **Personalized model per user** — Fine-tuning on individual user's writing patterns; future enhancement

</deferred>

---

*Phase: 07-02 — Modelo*
*Context gathered: 2026-05-23 via discuss-phase*
