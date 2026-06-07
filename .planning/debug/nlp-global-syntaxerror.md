---
status: resolved
trigger: "nlp-service crash on startup — SyntaxError: name '_current_model_version' used prior to global declaration"
slug: nlp-global-syntaxerror
created: 2026-06-07
updated: 2026-06-07
---

## Symptoms

1. **Expected behavior**: NLP service starts successfully, loads model, serves health/version/train endpoints
2. **Actual behavior**: Container starts, logs show model loaded, then crashes with SyntaxError
3. **Error messages**:
   ```
   File "/app/src/health.py", line 37
       global _current_model_version
       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   SyntaxError: name '_current_model_version' is used prior to global declaration
   ```
4. **Timeline**: Started after last deployment — likely Python version difference (Python 3.12+ made this a hard error)
5. **Reproduction**: `docker compose up` — NLP container fails to start every time

## Root Cause

In `nlp-service/src/health.py`, the `handle_train` function:
- References `_current_model_version` at line 27 (`return {"status": "already_running", "model_version": _current_model_version}`)
- But declares `global _current_model_version` only at line 37
- Python 3.12+ enforces that `global` declaration must appear before any reference to the name

## Fix

Move `global _current_model_version` to the top of `handle_train()`, before line 27.

## Resolution

- **hypothesis**: Confirmed — Python 3.12+ requires `global` before variable reference
- **fix**: Moved `global _current_model_version` from line 37 to line 26
- **verification**: 4/4 health tests passing ✅ (test_health_endpoint, test_health_method, test_model_version_endpoint, test_train_endpoint_returns_json)
- **files_changed**: `nlp-service/src/health.py`
- **resolved**: 2026-06-07
