import logging
import torch
from transformers import AutoTokenizer, AutoModelForSequenceClassification
from src.model_config import LABELS, NUM_LABELS, THRESHOLD, MODEL_PATH, MAX_LENGTH, MODEL_VERSION

logger = logging.getLogger(__name__)


class EmotionClassifier:

    def __init__(self, model_path: str = MODEL_PATH):
        logger.info("Loading emotion classifier from %s", model_path)
        self.device = torch.device("cpu")
        self.tokenizer = AutoTokenizer.from_pretrained(model_path)
        self.model = AutoModelForSequenceClassification.from_pretrained(model_path)
        self.model.to(self.device)
        self.model.eval()
        logger.info("EmotionClassifier loaded (%d labels)", NUM_LABELS)

    @torch.no_grad()
    def predict(self, text: str) -> dict:
        inputs = self.tokenizer(
            text,
            return_tensors="pt",
            truncation=True,
            padding=True,
            max_length=MAX_LENGTH,
        )
        inputs = {k: v.to(self.device) for k, v in inputs.items()}
        outputs = self.model(**inputs)
        probabilities = torch.sigmoid(outputs.logits).squeeze()

        if probabilities.dim() == 0:
            probabilities = probabilities.unsqueeze(0)

        scores_list = probabilities.tolist()

        scores_dict = {
            LABELS[i]: round(float(scores_list[i]), 4)
            for i in range(NUM_LABELS)
        }

        above_threshold = [
            (LABELS[i], float(scores_list[i]))
            for i in range(NUM_LABELS)
            if scores_list[i] >= THRESHOLD
        ]

        if not above_threshold:
            neutro_idx = LABELS.index("neutro")
            above_threshold = [("neutro", float(scores_list[neutro_idx]))]

        above_threshold.sort(key=lambda x: x[1], reverse=True)
        emotion_principal = above_threshold[0][0]
        intensidade = max(scores_list)

        return {
            "emotion_principal": emotion_principal,
            "emotions": [
                {"emotion": e, "score": round(s, 4)}
                for e, s in above_threshold
            ],
            "scores": scores_dict,
            "intensidade": round(float(intensidade), 4),
        }

    @property
    def labels(self) -> list:
        return LABELS


_classifier_instance = None


def get_classifier(model_path: str = MODEL_PATH) -> EmotionClassifier:
    global _classifier_instance
    if _classifier_instance is None:
        _classifier_instance = EmotionClassifier(model_path)
    return _classifier_instance
