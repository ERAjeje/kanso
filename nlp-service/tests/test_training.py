import pytest
import torch
import numpy as np
from train_model import (
    MODEL_NAME, NUM_LABELS, BATCH_SIZE, LEARNING_RATE,
    WeightedBCEWithLogitsLoss, compute_metrics, get_model_version,
    oversample_minority, load_curated_phrases, tokenize_function
)
from train_augment import load_back_translation_model, augment_dataset, deduplicate
from data.mappings import map_goemotions_row, compute_pos_weights
from src.model_config import MAX_LENGTH


def test_weighted_loss_creation():
    pos_weights = torch.ones(NUM_LABELS)
    loss_fn = WeightedBCEWithLogitsLoss(pos_weights)
    assert isinstance(loss_fn, torch.nn.Module)


def test_weighted_loss_forward():
    loss_fn = WeightedBCEWithLogitsLoss(torch.ones(NUM_LABELS))
    logits = torch.randn(4, NUM_LABELS)
    labels = torch.randint(0, 2, (4, NUM_LABELS)).float()
    loss = loss_fn(logits, labels)
    assert loss.item() > 0
    assert not torch.isnan(loss)


def test_weighted_loss_higher_for_wrong_predictions():
    weights = torch.ones(NUM_LABELS)
    weights[0] = 10.0
    loss_fn = WeightedBCEWithLogitsLoss(weights)

    logits_correct = torch.full((1, NUM_LABELS), -10.0)
    labels_all_zero = torch.zeros((1, NUM_LABELS))
    loss_correct = loss_fn(logits_correct, labels_all_zero)

    logits_wrong = torch.full((1, NUM_LABELS), -10.0)
    logits_wrong[0, 0] = 10.0
    loss_wrong = loss_fn(logits_wrong, labels_all_zero)

    assert loss_wrong > loss_correct


def test_compute_metrics_shape():
    logits = np.random.randn(10, NUM_LABELS)
    labels = np.random.randint(0, 2, (10, NUM_LABELS))
    result = compute_metrics((logits, labels))
    assert "weighted_f1" in result
    assert "micro_f1" in result
    assert isinstance(result["weighted_f1"], float)
    assert 0 <= result["weighted_f1"] <= 1


def test_oversample_minority_balance():
    text = ["texto exemplo"] * 100
    labels = []
    for i in range(100):
        vec = [0] * NUM_LABELS
        if i < 98:
            vec[0] = 1
        else:
            vec[12] = 1
        labels.append(vec)

    from datasets import Dataset
    ds = Dataset.from_dict({"text": text, "labels": labels})
    oversampled = oversample_minority(ds, minority_threshold=0.05)
    assert len(oversampled) > len(ds)


def test_get_model_version_returns_string():
    version = get_model_version()
    assert isinstance(version, str)
    assert len(version) > 0


def test_deduplicate_removes_duplicates():
    duplicates = [("a", [1, 0]), ("a", [1, 0]), ("b", [0, 1])]
    result = deduplicate(duplicates)
    assert len(result) == 2


@pytest.mark.slow
def test_training_smoke():
    from datasets import Dataset
    from transformers import (
        AutoTokenizer, AutoModelForSequenceClassification,
        TrainingArguments
    )
    from train_model import CustomTrainer

    texts = ["Estou muito feliz hoje", "Que tristeza essa notícia",
             "Estou com medo", "Sinto raiva dessa situação"] * 4
    labels_list = [[1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
                   [0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
                   [0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0],
                   [0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]] * 4
    ds = Dataset.from_dict({"text": texts, "labels": labels_list})

    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    def tok(ex):
        return tokenizer(ex["text"], padding="max_length",
                       truncation=True, max_length=32)
    ds = ds.map(tok, batched=True)
    ds.set_format(type="torch", columns=["input_ids", "attention_mask", "labels"])

    model = AutoModelForSequenceClassification.from_pretrained(
        MODEL_NAME, num_labels=NUM_LABELS
    )
    pos_weights = torch.ones(NUM_LABELS)

    args = TrainingArguments(
        output_dir="/tmp/test-training",
        num_train_epochs=1,
        per_device_train_batch_size=4,
        logging_steps=1,
        max_steps=5,
        report_to="none",
        save_strategy="no",
        remove_unused_columns=False,
    )
    trainer = CustomTrainer(
        pos_weights=pos_weights,
        model=model,
        args=args,
        train_dataset=ds,
    )
    trainer.train()
    assert True
