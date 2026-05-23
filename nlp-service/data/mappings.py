import numpy as np
import torch
from src.model_config import LABELS, NUM_LABELS

GOEMOTIONS_LABELS = [
    "admiration", "amusement", "anger", "annoyance", "approval", "caring",
    "confusion", "curiosity", "desire", "disappointment", "disapproval",
    "disgust", "embarrassment", "excitement", "fear", "gratitude", "grief",
    "joy", "love", "nervousness", "optimism", "pride", "realization",
    "relief", "remorse", "sadness", "surprise", "neutral"
]

LABEL_MAP = {
    "admiration": [10, 11],
    "amusement": [0],
    "anger": [2],
    "annoyance": [2],
    "approval": [0],
    "caring": [10],
    "confusion": [5],
    "curiosity": [5],
    "desire": [10],
    "disappointment": [1],
    "disapproval": [2],
    "disgust": [4],
    "embarrassment": [7],
    "excitement": [0],
    "fear": [3, 6],
    "gratitude": [11],
    "grief": [1, 9],
    "joy": [0],
    "love": [10],
    "nervousness": [6],
    "optimism": [0],
    "pride": [0],
    "realization": [5],
    "relief": [0, 11],
    "remorse": [8],
    "sadness": [1, 9],
    "surprise": [5],
    "neutral": [12],
}

NUM_GOEMOTIONS = 28
NUM_OUR_LABELS = 13


def goemotions_to_our_labels(goemotions_label: str) -> list[str]:
    indices = LABEL_MAP.get(goemotions_label, [])
    return [LABELS[i] for i in indices]


def map_goemotions_row(go_emotions_binary_vector: list[int]) -> list[int]:
    result = [0] * NUM_OUR_LABELS
    for ge_idx, ge_val in enumerate(go_emotions_binary_vector):
        if ge_val:
            ge_label = GOEMOTIONS_LABELS[ge_idx]
            our_indices = LABEL_MAP.get(ge_label, [])
            for idx in our_indices:
                result[idx] = 1
    return result


def compute_pos_weights(label_matrix: np.ndarray) -> torch.Tensor:
    num_examples = label_matrix.shape[0]
    pos_counts = label_matrix.sum(axis=0)
    neg_counts = num_examples - pos_counts
    pos_weights = neg_counts / np.maximum(pos_counts, 1)
    pos_weights = np.clip(pos_weights, 0.5, 50.0)
    return torch.FloatTensor(pos_weights)
