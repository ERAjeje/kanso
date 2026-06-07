import os
import sys
import logging
import subprocess
import numpy as np
import torch
import torch.nn as nn
from typing import List, Tuple, Optional
from datasets import Dataset, load_dataset, concatenate_datasets
from transformers import (
    AutoTokenizer, AutoModelForSequenceClassification,
    Trainer, TrainingArguments, EarlyStoppingCallback
)
import evaluate

from src.model_config import LABELS, NUM_LABELS, MAX_LENGTH, MODEL_VERSION, COUCHDB_URL, COUCHDB_TREINAMENTO_DB
from data.mappings import map_goemotions_row, compute_pos_weights, NUM_GOEMOTIONS
from data.curated_phrases import CURATED_PHRASES
from train_augment import load_back_translation_model, augment_dataset, deduplicate

logger = logging.getLogger(__name__)

MODEL_NAME = "neuralmind/bert-base-portuguese-cased"
SAVE_PATH = os.environ.get("MODEL_SAVE_PATH", "./model")

LEARNING_RATE = 3e-5
BATCH_SIZE = 8
GRADIENT_ACCUMULATION_STEPS = 2
NUM_EPOCHS = 5
WARMUP_RATIO = 0.1
EVAL_STRATEGY = "epoch"
SAVE_STRATEGY = "epoch"
MAX_GRAD_NORM = 1.0
WEIGHT_DECAY = 0.01


def get_model_version() -> str:
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            capture_output=True, text=True,
            timeout=5,
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return MODEL_VERSION


def load_goemotions_dataset() -> Dataset:
    try:
        ds = load_dataset("antoniomenezes/go_emotions_ptbr", split="train")
    except Exception as e:
        logger.warning("Failed to load GoEmotions-PT from HF: %s", e)
        raise

    def map_row(row):
        binary_28 = [row.get(ge_label, 0) for ge_label in [
            "admiration", "amusement", "anger", "annoyance", "approval", "caring",
            "confusion", "curiosity", "desire", "disappointment", "disapproval",
            "disgust", "embarrassment", "excitement", "fear", "gratitude", "grief",
            "joy", "love", "nervousness", "optimism", "pride", "realization",
            "relief", "remorse", "sadness", "surprise", "neutral"
        ]]
        our_labels = map_goemotions_row(binary_28)
        return {"text": row["texto"], "labels": our_labels}

    ds = ds.map(map_row, remove_columns=[col for col in ds.column_names if col != "texto"])
    return ds


def load_curated_phrases() -> Dataset:
    texts, labels = zip(*[(t, v) for t, v in CURATED_PHRASES])
    ds = Dataset.from_dict({"text": list(texts), "labels": list(labels)})
    return deduplicate_dataset(ds)


def label_to_multihot(label: str) -> list:
    """Convert a single label string to a multi-hot vector of length NUM_LABELS."""
    vec = [0] * NUM_LABELS
    try:
        idx = LABELS.index(label)
        vec[idx] = 1
    except ValueError:
        pass
    return vec


def load_training_from_couchdb():
    """Load training data from the CouchDB treinamento database.

    Connects to CouchDB, queries all docs with type 'treinamento',
    converts each label string to a multi-hot vector, and returns
    a Dataset in the same format as load_goemotions_dataset().

    Returns None if the connection fails or no data is found.
    """
    import requests

    try:
        # Determine the _find endpoint URL
        couchdb_find_url = COUCHDB_URL.rstrip("/") + "/" + COUCHDB_TREINAMENTO_DB + "/_find"

        query = {
            "selector": {"type": "treinamento"},
            "limit": 10000,
        }

        resp = requests.post(couchdb_find_url, json=query, timeout=10)
        resp.raise_for_status()
        data = resp.json()

        docs = data.get("docs", [])
        if not docs:
            logger.info("No training examples found in CouchDB treinamento DB")
            return None

        texts = []
        labels_list = []
        for doc in docs:
            texto = doc.get("texto", "").strip()
            label_str = doc.get("label", "").strip().lower()
            if not texto or not label_str:
                continue
            vector = label_to_multihot(label_str)
            texts.append(texto)
            labels_list.append(vector)

        if not texts:
            return None

        ds = Dataset.from_dict({"text": texts, "labels": labels_list})
        logger.info("Loaded %d training examples from CouchDB", len(ds))
        return ds

    except Exception as e:
        logger.warning("Failed to load training data from CouchDB: %s", e)
        return None


def deduplicate_dataset(ds: Dataset) -> Dataset:
    seen = set()
    text_col = ds["text"]
    label_col = ds["labels"]
    dedup_texts, dedup_labels = [], []
    for t, l in zip(text_col, label_col):
        if t.lower() not in seen:
            seen.add(t.lower())
            dedup_texts.append(t)
            dedup_labels.append(l)
    return Dataset.from_dict({"text": dedup_texts, "labels": dedup_labels})


def oversample_minority(dataset: Dataset, minority_threshold: float = 0.02) -> Dataset:
    labels_arr = np.array(dataset["labels"])
    freq = labels_arr.mean(axis=0)
    minority_indices = np.where(freq < minority_threshold)[0]

    if len(minority_indices) == 0:
        return dataset

    texts, labels_list = dataset["text"], dataset["labels"]
    extra_texts, extra_labels = [], []
    for t, l in zip(texts, labels_list):
        if any(l[idx] for idx in minority_indices):
            extra_texts.append(t)
            extra_labels.append(l)
            extra_texts.append(t)
            extra_labels.append(l)

    combined_texts = list(texts) + extra_texts
    combined_labels = list(labels_list) + extra_labels
    return Dataset.from_dict({"text": combined_texts, "labels": combined_labels})


def tokenize_function(examples: dict) -> dict:
    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    tokenized = tokenizer(
        examples["text"],
        padding="max_length",
        truncation=True,
        max_length=MAX_LENGTH,
    )
    tokenized["labels"] = examples["labels"]
    return tokenized


def compute_metrics(eval_pred):
    logits, labels = eval_pred
    probabilities = 1.0 / (1.0 + np.exp(-logits))
    threshold = 0.3
    predictions = (probabilities >= threshold).astype(int)

    f1_metric = evaluate.load("f1", config_name="multilabel")
    weighted_f1 = f1_metric.compute(predictions=predictions, references=labels.astype(int), average="weighted")["f1"]
    micro_f1 = f1_metric.compute(predictions=predictions, references=labels.astype(int), average="micro")["f1"]

    per_label_f1 = {}
    for i, label in enumerate(LABELS):
        if predictions[:, i].sum() > 0 or labels[:, i].sum() > 0:
            f1 = f1_metric.compute(predictions=predictions[:, i:i+1], references=labels[:, i:i+1].astype(int), average="binary")["f1"]
            per_label_f1[f"f1_{label}"] = round(float(f1), 4)

    return {"weighted_f1": round(float(weighted_f1), 4), "micro_f1": round(float(micro_f1), 4), **per_label_f1}


class WeightedBCEWithLogitsLoss(nn.Module):
    def __init__(self, pos_weights: torch.Tensor):
        super().__init__()
        self.loss_fn = nn.BCEWithLogitsLoss(pos_weight=pos_weights)

    def forward(self, logits, labels):
        return self.loss_fn(logits.view(-1, NUM_LABELS), labels.float().view(-1, NUM_LABELS))


class CustomTrainer(Trainer):
    def __init__(self, pos_weights, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.loss_fn = WeightedBCEWithLogitsLoss(pos_weights)

    def compute_loss(self, model, inputs, return_outputs=False, num_items_in_batch=None):
        labels = inputs.pop("labels")
        outputs = model(**inputs)
        logits = outputs.logits
        loss = self.loss_fn(logits, labels)
        return (loss, outputs) if return_outputs else loss


def train():
    logger.info("Step 1: Loading GoEmotions PT-BR dataset")
    goemotions_ds = load_goemotions_dataset()
    logger.info("GoEmotions-PT loaded: %d rows", len(goemotions_ds))

    logger.info("Step 2: Loading curated Portuguese phrases")
    curated_ds = load_curated_phrases()
    logger.info("Curated phrases loaded: %d rows", len(curated_ds))

    logger.info("Step 3: Combining datasets")
    combined = concatenate_datasets([goemotions_ds, curated_ds])

    logger.info("Step 3b: Loading training data from CouchDB treinamento DB")
    couchdb_ds = load_training_from_couchdb()
    if couchdb_ds is not None and len(couchdb_ds) > 0:
        logger.info("CouchDB training data loaded: %d rows", len(couchdb_ds))
        combined = concatenate_datasets([combined, couchdb_ds])

    logger.info("Step 4: Splitting train/validation (80/20)")
    split = combined.train_test_split(test_size=0.2, seed=42)
    train_ds = split["train"]
    eval_ds = split["test"]

    logger.info("Step 5 (optional): Augment training data via back-translation")
    augment = os.environ.get("AUGMENT", "false").lower() == "true"
    if augment:
        model_bt, tokenizer_bt = load_back_translation_model()
        train_phrases = list(zip(train_ds["text"], train_ds["labels"]))
        augmented = augment_dataset(train_phrases, model_bt, tokenizer_bt, augment_ratio=0.5)
        aug_texts, aug_labels = zip(*augmented)
        train_ds = Dataset.from_dict({"text": list(aug_texts), "labels": list(aug_labels)})

    logger.info("Step 6: Oversampling minority classes")
    train_ds = oversample_minority(train_ds, minority_threshold=0.02)
    logger.info("Train size after oversampling: %d", len(train_ds))

    logger.info("Step 7: Tokenizing")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)

    def tok_fn(examples):
        tokenized = tokenizer(
            examples["text"],
            padding="max_length",
            truncation=True,
            max_length=MAX_LENGTH,
        )
        tokenized["labels"] = examples["labels"]
        return tokenized

    train_ds = train_ds.map(tok_fn, batched=True)
    eval_ds = eval_ds.map(tok_fn, batched=True)

    logger.info("Step 8: Computing pos_weights from training data")
    labels_array = np.array(train_ds["labels"])
    pos_weights = compute_pos_weights(labels_array)

    logger.info("Step 9: Setting tensor format")
    train_ds.set_format(type="torch", columns=["input_ids", "attention_mask", "labels"])
    eval_ds.set_format(type="torch", columns=["input_ids", "attention_mask", "labels"])

    logger.info("Step 10: Loading BERTimbau (full fine-tune, no LoRA)")
    model = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME, num_labels=NUM_LABELS)

    logger.info("Step 11: Configuring Trainer")
    training_args = TrainingArguments(
        output_dir="./training-output",
        learning_rate=LEARNING_RATE,
        per_device_train_batch_size=BATCH_SIZE,
        gradient_accumulation_steps=GRADIENT_ACCUMULATION_STEPS,
        num_train_epochs=NUM_EPOCHS,
        warmup_ratio=WARMUP_RATIO,
        evaluation_strategy=EVAL_STRATEGY,
        save_strategy=SAVE_STRATEGY,
        max_grad_norm=MAX_GRAD_NORM,
        weight_decay=WEIGHT_DECAY,
        logging_steps=10,
        logging_dir="./logs",
        remove_unused_columns=False,
        dataloader_pin_memory=False,
        report_to="none",
    )

    trainer = CustomTrainer(
        pos_weights=pos_weights,
        model=model,
        args=training_args,
        train_dataset=train_ds,
        eval_dataset=eval_ds,
        compute_metrics=compute_metrics,
        callbacks=[EarlyStoppingCallback(early_stopping_patience=2)],
    )

    logger.info("Step 12: Training")
    trainer.train()

    logger.info("Step 13: Saving model")
    os.makedirs(SAVE_PATH, exist_ok=True)
    model.save_pretrained(SAVE_PATH)
    tokenizer.save_pretrained(SAVE_PATH)

    version = get_model_version()
    logger.info("Model version: %s", version)
    logger.info("Model saved to: %s", SAVE_PATH)
    logger.info("To use in Docker: cp -r %s/* /model", SAVE_PATH)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    train()
