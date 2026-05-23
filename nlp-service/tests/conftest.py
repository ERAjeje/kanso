import sys
import os
import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session")
def labels():
    from src.model_config import LABELS
    return LABELS
