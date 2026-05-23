import sys
import os
import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session")
def labels():
    from src.model_config import LABELS
    return LABELS


@pytest.fixture(scope="session")
def classifier():
    from src.model_config import MODEL_PATH
    if not os.path.exists(MODEL_PATH):
        pytest.skip(f"Model not found at {MODEL_PATH} — run training first")
    from src.classifier import EmotionClassifier
    return EmotionClassifier(MODEL_PATH)
