import pytest
import numpy as np
from src.model_config import LABELS, NUM_LABELS, THRESHOLD, MODEL_VERSION
from data.mappings import (
    GOEMOTIONS_LABELS, LABEL_MAP, NUM_GOEMOTIONS, NUM_OUR_LABELS,
    map_goemotions_row, compute_pos_weights
)
from data.curated_phrases import CURATED_PHRASES, get_label_distribution


def test_model_config_constants():
    assert len(LABELS) == 13
    assert THRESHOLD == 0.3
    assert MODEL_VERSION == "v1.0"


def test_goemotions_labels_count():
    assert len(GOEMOTIONS_LABELS) == 28
    assert NUM_GOEMOTIONS == 28
    assert NUM_OUR_LABELS == 13


def test_label_map_coverage():
    for ge_label in GOEMOTIONS_LABELS:
        assert ge_label in LABEL_MAP, f"Missing mapping for {ge_label}"
        mapped_indices = LABEL_MAP[ge_label]
        assert len(mapped_indices) >= 1
        assert len(mapped_indices) <= 2
        for idx in mapped_indices:
            assert 0 <= idx < 13, f"Index {idx} out of range for {ge_label}"


def test_map_goemotions_row_single_label():
    binary_28 = [0] * 28
    binary_28[GOEMOTIONS_LABELS.index("admiration")] = 1
    result = map_goemotions_row(binary_28)
    assert len(result) == 13
    assert result[10] == 1
    assert result[11] == 1


def test_map_goemotions_row_empty():
    result = map_goemotions_row([0] * 28)
    assert result == [0] * 13


def test_map_goemotions_row_neutral():
    binary_28 = [0] * 28
    binary_28[GOEMOTIONS_LABELS.index("neutral")] = 1
    result = map_goemotions_row(binary_28)
    assert result[12] == 1


def test_compute_pos_weights_shape_and_clip():
    label_matrix = np.zeros((100, 13))
    label_matrix[:80, 0] = 1
    label_matrix[:2, 5] = 1
    weights = compute_pos_weights(label_matrix)
    assert weights.shape == (13,)
    assert weights[0] < weights[5]
    assert weights.min() >= 0.5
    assert weights.max() <= 50.0


def test_curated_phrases_count():
    assert len(CURATED_PHRASES) >= 2600


def test_curated_phrases_format():
    for text, vec in CURATED_PHRASES:
        assert isinstance(text, str) and len(text) > 0
        assert isinstance(vec, list) and len(vec) == 13
        assert all(v in (0, 1) for v in vec)


def test_curated_phrases_label_coverage():
    counts = get_label_distribution()
    for label in LABELS:
        assert counts[label] >= 200, f"{label} only has {counts[label]} phrases"


def test_curated_phrases_portuguese():
    portuguese_markers = [
        "ão", "ções", "sinto", "estou", "meu", "minha", "com",
        "para", "por", "que", "não", "mais", "como", "dos", "das"
    ]
    non_portuguese = []
    for text, _ in CURATED_PHRASES:
        has_marker = any(m in text.lower() for m in portuguese_markers)
        if not has_marker:
            non_portuguese.append(text)
    assert len(non_portuguese) < len(CURATED_PHRASES) * 0.05
