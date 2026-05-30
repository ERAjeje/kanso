import pytest
import torch
from src.classifier import EmotionClassifier
from src.model_config import LABELS, THRESHOLD, NUM_LABELS, MODEL_VERSION


class TestEmotionClassifierContract:

    def test_predict_structure(self, classifier):
        result = classifier.predict("Estou muito feliz hoje")
        assert isinstance(result, dict)
        assert "emotion_principal" in result
        assert "emotions" in result
        assert "scores" in result
        assert "intensidade" in result

    def test_predict_emotion_principal_is_string(self, classifier):
        result = classifier.predict("Estou com medo")
        assert isinstance(result["emotion_principal"], str)
        assert result["emotion_principal"] in LABELS

    def test_predict_emotions_is_list(self, classifier):
        result = classifier.predict("Que dia lindo")
        assert isinstance(result["emotions"], list)
        assert len(result["emotions"]) >= 1
        for e in result["emotions"]:
            assert "emotion" in e
            assert "score" in e
            assert e["emotion"] in LABELS
            assert 0 <= e["score"] <= 1

    def test_predict_scores_all_labels(self, classifier):
        result = classifier.predict("Estou cansado")
        assert isinstance(result["scores"], dict)
        assert len(result["scores"]) == NUM_LABELS
        for label in LABELS:
            assert label in result["scores"]
            assert 0 <= result["scores"][label] <= 1

    def test_predict_intensidade_is_max_score(self, classifier):
        result = classifier.predict("Estou muito bem")
        max_score = max(result["scores"].values())
        assert abs(result["intensidade"] - max_score) < 0.001

    def test_predict_empty_text_returns_neutro(self, classifier):
        result = classifier.predict("")
        assert result["emotion_principal"] == "neutro"

    def test_predict_short_text_returns_neutro(self, classifier):
        result = classifier.predict("ok")
        assert result["emotion_principal"] == "neutro"

    def test_modelo_versao_string(self, classifier):
        version = MODEL_VERSION
        assert isinstance(version, str)
        assert len(version) > 0
