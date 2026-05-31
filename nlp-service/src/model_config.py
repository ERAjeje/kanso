import os

LABELS = [
    "alegria", "tristeza", "raiva", "medo", "nojo", "surpresa",
    "ansiedade", "vergonha", "culpa", "saudade", "amor", "gratidão", "neutro",
]

NUM_LABELS = len(LABELS)

THRESHOLD = 0.3

MODEL_VERSION = os.environ.get("MODEL_VERSION", "v1.0")

MODEL_PATH = os.environ.get("MODEL_PATH", "/model")

MAX_LENGTH = 128
