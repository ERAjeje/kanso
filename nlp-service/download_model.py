import logging
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

MODEL_NAME = "neuralmind/bert-base-portuguese-cased"
SAVE_PATH = "/model"

if __name__ == "__main__":
    logger.info("Downloading model: %s", MODEL_NAME)
    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    model = AutoModelForSequenceClassification.from_pretrained(
        MODEL_NAME,
        num_labels=13
    )
    tokenizer.save_pretrained(SAVE_PATH)
    model.save_pretrained(SAVE_PATH)
    logger.info("Model saved to %s", SAVE_PATH)
