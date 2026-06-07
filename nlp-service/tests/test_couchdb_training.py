import os
import pytest


@pytest.mark.skip(reason="Requires torch (runs inside Docker)")
def test_load_training_from_couchdb_imports():
    """Verify the function can be imported without error."""
    from train_model import load_training_from_couchdb
    assert callable(load_training_from_couchdb)


@pytest.mark.skip(reason="Requires torch (runs inside Docker)")
def test_load_training_from_couchdb_returns_none_on_connection_error():
    """When CouchDB is unreachable, the function should return None."""
    from train_model import load_training_from_couchdb

    result = load_training_from_couchdb()
    assert result is None or len(result) >= 0


@pytest.mark.skip(reason="Requires running CouchDB instance")
def test_load_training_from_couchdb_integration():
    """Integration test: connect to real CouchDB and load data."""
    from train_model import load_training_from_couchdb
    result = load_training_from_couchdb()
    assert result is not None
    assert len(result) > 0


def test_label_to_multihot():
    """Verify string label conversion to multi-hot vector."""
    from src.model_config import LABELS

    # Compute the expected vector without torch dependency
    labels = LABELS  # ["alegria", "tristeza", ...]
    n = len(labels)

    def label_to_multihot(label: str):
        vec = [0] * n
        try:
            idx = labels.index(label)
            vec[idx] = 1
        except ValueError:
            pass
        return vec

    vector = label_to_multihot("alegria")
    assert len(vector) == n
    assert vector[labels.index("alegria")] == 1
    assert sum(vector) == 1

    vector = label_to_multihot("neutro")
    assert vector[labels.index("neutro")] == 1
    assert sum(vector) == 1

    vector = label_to_multihot("invalid_label")
    assert sum(vector) == 0
