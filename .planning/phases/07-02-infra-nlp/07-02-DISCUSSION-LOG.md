# Phase 07-02: Modelo — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Phase:** 07-02 — Modelo (BERTimbau fine-tuning, emotion classification pipeline, model validation)
**Areas discussed:** Fine-tuning approach, Training data & augmentation, Inference/classification strategy, Model validation, Serving integration, Model versioning & iteration

---

## Fine-tuning Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Full fine-tune (Recommended) | Fine-tune all BERTimbau layers. Higher quality for domain adaptation. | ✓ |
| LoRA adapters | Train only small adapter layers. Cheaper but lower quality for subtle emotion distinctions. | |
| Classifier head only | Freeze BERTimbau, train only the classification head. Fastest but worst quality. | |

**User's choice:** Full fine-tune (Recommended)
**Notes:** Quality for nuanced Portuguese emotions (saudade vs tristeza) requires full fine-tuning.

| Option | Description | Selected |
|--------|-------------|----------|
| Local / dev machine (Recommended) | Run training on dev machine with GPU. Model artifact copied into Docker build. | ✓ |
| Google Colab | Free/paid GPU runtime. Export trained weights. | |
| CI/CD pipeline | Dedicated GPU runner in CI. Most automated but complex. | |

**User's choice:** Local / dev machine (Recommended)
**Notes:** Training happens locally; only the trained model artifact goes into Docker.

| Option | Description | Selected |
|--------|-------------|----------|
| Training script in nlp-service/ (Recommended) | Single train_model.py alongside inference code | ✓ |
| Separate training repo | Dedicated directory for training pipeline | |
| Notebook-based | Jupyter notebook for training | |

**User's choice:** Training script in nlp-service/ (Recommended)
**Notes:** Single script, self-contained, lives alongside the serving code.

| Option | Description | Selected |
|--------|-------------|----------|
| Download at training time (Recommended) | Script downloads GoEmotions-PT from HuggingFace during training | ✓ |
| Checkpoint into repo | Download once, save snapshot in repo | |

**User's choice:** Download at training time (Recommended)
**Notes:** Always gets latest dataset version, no data in repo.

---

## Training Data & Augmentation

| Option | Description | Selected |
|--------|-------------|----------|
| Manual Portuguese curation (Recommended) | ~200-500 curated Portuguese emotion phrases covering cultural nuances | ✓ |
| Pure GoEmotions-PT | Use as-is, accept lower quality for Portuguese-specific emotions | |
| Synthetic augmentation | Use GPT/LLM to generate Portuguese emotion phrases | |

**User's choice:** Manual Portuguese curation (Recommended)
**Notes:** Covers saudade, vergonha, culpa — culturally relevant emotions that GoEmotions-PT may miss.

| Option | Description | Selected |
|--------|-------------|----------|
| Weighted loss + oversampling (Recommended) | Class weights + oversample minority classes | ✓ |
| Weighted loss only | Simpler but model sees fewer minority examples | |
| Undersample majority | Drop majority to match minority count — wastes data | |

**User's choice:** Weighted loss + oversampling (Recommended)
**Notes:** Standard approach for emotion datasets.

| Option | Description | Selected |
|--------|-------------|----------|
| Back-translation PT↔ES (Recommended) | Translate Portuguese→Spanish→Portuguese using MarianMT | ✓ |
| Synonym replacement + random swap | Replace with Portuguese synonyms | |
| Minimal augmentation | Basic dropout at embedding level | |

**User's choice:** Back-translation PT↔ES (Recommended)
**Notes:** Most natural augmentation for Portuguese; preserves emotion while generating diverse phrasing.

---

## Inference / Classification Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Multi-label with threshold (Recommended) | Return ALL emotions above confidence threshold | ✓ |
| Single-label (top-1) | Return only highest-scoring emotion | |
| Multi-label with top-K | Return top K regardless of score | |

**User's choice:** Multi-label with threshold (Recommended)
**Notes:** Portuguese speakers commonly express mixed emotions.

| Option | Description | Selected |
|--------|-------------|----------|
| 0.3 (Recommended) | Balanced — captures secondary emotions without too much noise | ✓ |
| 0.5 | Higher confidence, cleaner results | |
| Adaptive / no fixed threshold | Return all scores, let consumers decide | |

**User's choice:** 0.3 (Recommended)
**Notes:** Start here, tune based on validation results.

| Option | Description | Selected |
|--------|-------------|----------|
| Max emotion score (Recommended) | Intensidade = highest emotion confidence score | ✓ |
| Mean of top-K scores | Average of top scores, may dilute strong emotions | |
| Weighted combination | Custom formula — most complex | |

**User's choice:** Max emotion score (Recommended)
**Notes:** Clean, interpretable intensity metric.

| Option | Description | Selected |
|--------|-------------|----------|
| Return neutro with high confidence (Recommended) | Short/uninformative → classify as neutro | ✓ |
| Return low-confidence scores | Let model return whatever it produces | |
| Explicit length check | Skip inference if < 10 chars | |

**User's choice:** Return neutro with high confidence (Recommended)
**Notes:** Model learns this from training data naturally.

---

## Model Validation

| Option | Description | Selected |
|--------|-------------|----------|
| Weighted F1 >= 0.70 (Recommended) | On curated Portuguese test set. Accounts for class imbalance. | ✓ |
| Accuracy >= 0.75 | Simple accuracy — may be misleading with imbalance | |
| Human evaluation on 50 samples | Manual review by domain expert | |

**User's choice:** Weighted F1 >= 0.70 (Recommended)
**Notes:** Realistic target for 13-class Portuguese emotion classification with BERTimbau fine-tune.

| Option | Description | Selected |
|--------|-------------|----------|
| Iterate with more data + tuning (Recommended) | Add weak emotion examples, adjust LR, try unfreezing layers | ✓ |
| Accept and ship with monitoring | Ship at current quality, improve iteratively | |
| Switch to LoRA or different base model | Try different approach entirely | |

**User's choice:** Iterate with more data + hyperparameter tuning (Recommended)
**Notes:** Budget 2-3 iteration rounds before declaring done.

| Option | Description | Selected |
|--------|-------------|----------|
| Inline Python file in nlp-service/ (Recommended) | test_phrases.py with ~50 phrases per emotion | ✓ |
| JSON data file | Separate JSON file with phrases and labels | |
| Include in GoEmotions-PT download script | Unified data pipeline | |

**User's choice:** Inline Python file in nlp-service/ (Recommended)
**Notes:** Versioned, lives with training code, easy to extend.

---

## Serving Integration

| Option | Description | Selected |
|--------|-------------|----------|
| Server startup (Recommended) | Load model in server.py on startup, module-level variable | ✓ |
| Lazy load on first request | Load on first Analyze() call | |
| Separate warm-up endpoint | Dedicated /warmup endpoint | |

**User's choice:** Server startup (Recommended)
**Notes:** First request may wait ~5-10s for model load, subsequent requests fast.

| Option | Description | Selected |
|--------|-------------|----------|
| No batching (Recommended) | Each Analyze call runs inference immediately | ✓ |
| Dynamic batching | Collect requests for ~100ms window | |

**User's choice:** No batching (Recommended)
**Notes:** Single-user app processing one registration at a time. Can add batching later.

| Option | Description | Selected |
|--------|-------------|----------|
| CPU inference (Recommended) | BERTimbau base runs fine on CPU (~200-500ms) | ✓ |
| GPU inference | Faster (~20-50ms) but adds Docker complexity | |

**User's choice:** CPU inference (Recommended)
**Notes:** Simpler Docker setup without GPU passthrough.

---

## Model Versioning & Iteration

| Option | Description | Selected |
|--------|-------------|----------|
| Git-tracked version (Recommended) | Git commit hash or semver in modelo_versao field | ✓ |
| HuggingFace model registry | Push to HF Hub, pull on Docker build | |
| Docker image tag | Each model version = new Docker build | |

**User's choice:** Git-tracked version (Recommended)
**Notes:** Simple, auditable, no extra infra.

| Option | Description | Selected |
|--------|-------------|----------|
| Manual periodic retraining (Recommended) | Monthly or as data accumulates; manually label, retrain | ✓ |
| Active learning loop | Auto-flag low-confidence predictions for review | |
| Ship and forget — no iteration | Only retrain if issues reported | |

**User's choice:** Manual periodic retraining (Recommended)
**Notes:** Quality-controlled iteration cadence owned by developer.

---

## the agent's Discretion

- Exact hyperparameters (learning rate, epochs, batch size) for fine-tuning
- Specific MarianMT model for back-translation
- Python module structure within nlp-service/ for the classification pipeline
- Test file organization within test_phrases.py

## Deferred Ideas

- GPU inference in production — Defer until CPU latency becomes a bottleneck
- Active learning loop — Too complex for initial version
- HuggingFace model registry — Adds external dependency; defer until model proves value
- Personalized model per user — Future enhancement
