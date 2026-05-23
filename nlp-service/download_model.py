import logging
import os
import sys
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

SAVE_PATH = os.environ.get("MODEL_SAVE_PATH", "/model")
CHECKPOINT_PATH = os.environ.get("CHECKPOINT_PATH", "")

if __name__ == "__main__":
    if CHECKPOINT_PATH and os.path.exists(CHECKPOINT_PATH):
        logger.info("Loading fine-tuned model from checkpoint: %s", CHECKPOINT_PATH)
        tokenizer = AutoTokenizer.from_pretrained(CHECKPOINT_PATH)
        model = AutoModelForSequenceClassification.from_pretrained(CHECKPOINT_PATH)
    else:
        logger.warning(
            "No CHECKPOINT_PATH provided or path does not exist: %s\n"
            "Running inference requires a trained model.\n"
            "Train first: python nlp-service/train_model.py\n"
            "Then set CHECKPOINT_PATH=<trained-model-dir>",
            CHECKPOINT_PATH
        )
        logger.info("Downloading base BERTimbau as fallback for testing")
        MODEL_NAME = "neuralmind/bert-base-portuguese-cased"
        tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
        model = AutoModelForSequenceClassification.from_pretrained(
            MODEL_NAME, num_labels=13
        )

    os.makedirs(SAVE_PATH, exist_ok=True)
    tokenizer.save_pretrained(SAVE_PATH)
    model.save_pretrained(SAVE_PATH)
    logger.info("Model saved to %s", SAVE_PATH)
