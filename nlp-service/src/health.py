import threading
import logging
from fastapi import FastAPI

from src.model_config import MODEL_VERSION

app = FastAPI(title="Kanso NLP Service")
logger = logging.getLogger(__name__)

_train_lock = threading.Lock()
_current_model_version = MODEL_VERSION


@app.get("/health")
async def health():
    return {"status": "ok", "service": "nlp", "grpc_port": 50051}


@app.get("/model/version")
async def model_version():
    return {"version": _current_model_version}


@app.post("/train")
async def handle_train():
    global _current_model_version

    if not _train_lock.acquire(blocking=False):
        return {"status": "already_running", "model_version": _current_model_version}

    try:
        from train_model import train

        logger.info("Training triggered via HTTP endpoint")
        train()

        # After training, increment model version
        from src.model_config import MODEL_VERSION as BASE_VERSION

        version_parts = BASE_VERSION.lstrip("v").split(".")
        major = int(version_parts[0])
        minor = int(version_parts[1]) if len(version_parts) > 1 else 0
        minor += 1
        _current_model_version = f"v{major}.{minor}"

        logger.info("Training complete. New model version: %s", _current_model_version)

        return {
            "status": "ok",
            "model_version": _current_model_version,
            "trained_count": -1,
        }
    except Exception as e:
        logger.error("Training failed: %s", e)
        return {
            "status": "error",
            "message": str(e),
            "model_version": _current_model_version,
        }
    finally:
        _train_lock.release()
