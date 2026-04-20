from pathlib import Path
import joblib

BASE_DIR = Path(__file__).resolve().parent.parent
ARTIFACTS_DIR = BASE_DIR / "artifacts"

CATEGORY_MODEL_PATH = ARTIFACTS_DIR / "category_pipeline.joblib"
SAFE_TO_SAVE_MODEL_PATH = ARTIFACTS_DIR / "safe_to_save_model.joblib"


def load_category_model():
    if not CATEGORY_MODEL_PATH.exists():
        raise FileNotFoundError(f"Missing category model at {CATEGORY_MODEL_PATH}")
    return joblib.load(CATEGORY_MODEL_PATH)


def load_safe_to_save_model():
    if not SAFE_TO_SAVE_MODEL_PATH.exists():
        raise FileNotFoundError(f"Missing safe-to-save model at {SAFE_TO_SAVE_MODEL_PATH}")
    return joblib.load(SAFE_TO_SAVE_MODEL_PATH)