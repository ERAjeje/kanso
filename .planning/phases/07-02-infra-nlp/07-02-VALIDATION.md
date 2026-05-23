---
phase: 07-02
slug: modelo
status: finalized
nyquist_compliant: true
wave_0_complete: true
created: 2026-05-23
updated: 2026-05-23
---

# Phase 07-02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | pytest 8.x |
| **Config file** | `nlp-service/tests/conftest.py` (pytest-asyncio plugin) |
| **Quick run command** | `python -m pytest nlp-service/tests/ -x -q` |
| **Full suite command** | `python -m pytest nlp-service/tests/ -v --tb=short` |
| **Estimated runtime** | ~120 seconds (includes model inference for integration tests) |

---

## Sampling Rate

- **After every task commit:** Run `python -m pytest nlp-service/tests/ -x --timeout=60 -q`
- **After every plan wave:** Run `python -m pytest nlp-service/tests/ -v --tb=short`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 07-02-01-01 | 01 | 1 | NLP-02 | T-07-02-01 | Label mapping validation | unit | `python -c "from data.mappings import GOEMOTIONS_LABELS, LABEL_MAP; assert len(LABEL_MAP) == 28; print('OK')"` | ❌ W0 | ⬜ pending |
| 07-02-01-02 | 01 | 1 | NLP-02 | T-07-02-02 | Phrase count ≥200 per label | unit | `python -c "from data.curated_phrases import CURATED_PHRASES; assert len(CURATED_PHRASES) >= 2600; print('OK')"` | ❌ W0 | ⬜ pending |
| 07-02-01-03 | 01 | 1 | NLP-02 | — | Data pipeline unit tests | unit | `python -m pytest nlp-service/tests/test_data.py -x -q` | ❌ W0 | ⬜ pending |
| 07-02-02-01 | 02 | 2 | NLP-02 | T-07-02-04 | Training structure check | unit | `python -c "from train_model import CustomTrainer, WeightedBCEWithLogitsLoss; print('OK')"` | ❌ W0 | ⬜ pending |
| 07-02-02-02 | 02 | 2 | NLP-02 | T-07-02-05 | Training smoke tests | unit | `python -m pytest nlp-service/tests/test_training.py -x -q -k \"not slow\"` | ❌ W0 | ⬜ pending |
| 07-02-03-01 | 03 | 2 | NLP-01, NLP-02 | T-07-02-08 | EmotionClassifier structure | unit | `python -c "from src.classifier import EmotionClassifier; print('OK')"` | ❌ W0 | ⬜ pending |
| 07-02-03-02 | 03 | 2 | NLP-01, D-28, D-30 | T-07-02-11 | Phrase count + label validity | unit | `python -m pytest nlp-service/tests/test_phrases.py -x -q` | ❌ W0 | ⬜ pending |
| 07-02-03-02 (eval) | 03 | 2 | D-28 | T-07-02-11 | Weighted F1 ≥ 0.70 evaluation | eval | `python -m pytest nlp-service/tests/test_phrases.py::test_weighted_f1_threshold -x` | ❌ W0 | ⬜ pending |
| 07-02-03-03 | 03 | 2 | NLP-01, NLP-02, NLP-03 | T-07-02-08, T-07-02-12 | Server integration tests | integration | `python -m pytest nlp-service/tests/test_server.py -x -q -k "test_analyze_empty or test_grpc_server_starts"` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `nlp-service/tests/test_data.py` — data pipeline tests (Plan 01 Task 3)
- [x] `nlp-service/tests/test_training.py` — training smoke test (Plan 02 Task 3)
- [x] `nlp-service/tests/test_classifier.py` — inference pipeline unit tests (Plan 03 Task 2)
- [x] `nlp-service/tests/test_phrases.py` — curated test phrases + Weighted F1 evaluation (Plan 03 Task 2)
- [x] `nlp-service/tests/conftest.py` — pytest-asyncio plugin + shared fixtures (Plan 01 Task 3, updated in Plan 03 Task 3)

**Notes:**
- `test_integration.py` (end-to-end gRPC with toy model) intentionally omitted — the server tests in `test_server.py` serve this purpose via direct `AnalysisServicer.Analyze()` calls, which exercise the same code path without requiring a running gRPC server.
- No separate `pytest.ini` needed — `conftest.py` provides pytest-asyncio plugin registration.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have automated verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter
- [x] D-28 evaluation harness: `test_weighted_f1_threshold()` in test_phrases.py loads model, computes Weighted F1 via evaluate, asserts >= 0.70 (conditional skip when no model artifact)
- [x] server.py module-level import does not crash when model artifact missing (try/except + lazy init fallback)

**Approval:** ✅ Self-validated (revision iteration 1)
