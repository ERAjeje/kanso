import logging
import random
from typing import List, Tuple
from transformers import M2M100ForConditionalGeneration, M2M100Tokenizer
from src.model_config import MAX_LENGTH

logger = logging.getLogger(__name__)

MODEL_NAME = "facebook/m2m100_418M"
SOURCE_LANG = "pt"
PIVOT_LANG = "es"
NUM_BEAMS = 4

_model = None
_tokenizer = None


def load_back_translation_model():
    global _model, _tokenizer
    if _model is None:
        logger.info("Loading back-translation model: %s", MODEL_NAME)
        _model = M2M100ForConditionalGeneration.from_pretrained(MODEL_NAME)
        _tokenizer = M2M100Tokenizer.from_pretrained(MODEL_NAME)
        logger.info("Back-translation model loaded")
    return _model, _tokenizer


def back_translate(text: str, model, tokenizer) -> str:
    try:
        tokenizer.src_lang = SOURCE_LANG
        encoded = tokenizer(text, return_tensors="pt", padding=True, truncation=True, max_length=MAX_LENGTH)
        generated_tokens = model.generate(
            **encoded,
            forced_bos_token_id=tokenizer.get_lang_id(PIVOT_LANG),
            num_beams=NUM_BEAMS,
        )
        pivot_text = tokenizer.batch_decode(generated_tokens, skip_special_tokens=True)[0]

        tokenizer.src_lang = PIVOT_LANG
        encoded_pivot = tokenizer(pivot_text, return_tensors="pt", padding=True, truncation=True, max_length=MAX_LENGTH)
        generated_back = model.generate(
            **encoded_pivot,
            forced_bos_token_id=tokenizer.get_lang_id(SOURCE_LANG),
            num_beams=NUM_BEAMS,
        )
        back_translated = tokenizer.batch_decode(generated_back, skip_special_tokens=True)[0]
        return back_translated
    except Exception as e:
        logger.warning("Back-translation failed for '%s': %s", text[:30], e)
        return text


def augment_dataset(phrases: List[Tuple[str, List[int]]], model, tokenizer, augment_ratio: float = 0.5) -> List[Tuple[str, List[int]]]:
    num_to_augment = int(len(phrases) * augment_ratio)
    sampled = random.sample(phrases, min(num_to_augment, len(phrases)))
    augmented = list(phrases)
    success = 0
    failed = 0
    for text, label_vector in sampled:
        aug_text = back_translate(text, model, tokenizer)
        if aug_text != text:
            augmented.append((aug_text, label_vector))
            success += 1
        else:
            failed += 1
    logger.info("Augmentation: %d succeeded, %d failed, total %d phrases", success, failed, len(augmented))
    return augmented


def deduplicate(phrases: List[Tuple[str, List[int]]]) -> List[Tuple[str, List[int]]]:
    seen = set()
    result = []
    for text, vec in phrases:
        if text.lower() not in seen:
            seen.add(text.lower())
            result.append((text, vec))
    return result


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    model, tokenizer = load_back_translation_model()
    from data.curated_phrases import CURATED_PHRASES
    augmented = augment_dataset(CURATED_PHRASES, model, tokenizer, augment_ratio=0.5)
    print(f"Original: {len(CURATED_PHRASES)}, After aug: {len(augmented)}")
