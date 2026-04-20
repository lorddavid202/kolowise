from fastapi import FastAPI
import pandas as pd

from app.schemas import (
    CategoryPredictionRequest,
    CategoryPredictionResponse,
    SafeToSaveRequest,
    SafeToSaveResponse,
)
from app.model_loader import load_category_model, load_safe_to_save_model

app = FastAPI(title="KoloWise ML Service")

category_model = None
safe_to_save_model = None


@app.on_event("startup")
def startup_event():
    global category_model, safe_to_save_model
    category_model = load_category_model()
    safe_to_save_model = load_safe_to_save_model()


@app.get("/healthz")
def healthz():
    return {
        "status": "ok",
        "models": {
            "category_model_loaded": category_model is not None,
            "safe_to_save_model_loaded": safe_to_save_model is not None,
        },
    }


@app.post("/predict/category", response_model=CategoryPredictionResponse)
def predict_category(payload: CategoryPredictionRequest):
    text = f"{payload.narration} {payload.merchant_name} {payload.direction} {bucket_amount(payload.amount)}"
    prediction = category_model.predict([text])[0]

    confidence = 0.0
    if hasattr(category_model, "predict_proba"):
        probabilities = category_model.predict_proba([text])[0]
        confidence = float(max(probabilities))

    return CategoryPredictionResponse(
        category=prediction,
        confidence=round(confidence, 4),
        model_name="transaction_category_logreg_v1",
    )


@app.post("/predict/safe-to-save", response_model=SafeToSaveResponse)
def predict_safe_to_save(payload: SafeToSaveRequest):
    features = pd.DataFrame(
        [
            {
                "available_balance": payload.available_balance,
                "avg_monthly_income": payload.avg_monthly_income,
                "avg_monthly_expense": payload.avg_monthly_expense,
                "emergency_buffer": payload.emergency_buffer,
                "monthly_surplus": payload.monthly_surplus,
                "active_goal_need": payload.active_goal_need,
            }
        ]
    )

    predicted = float(safe_to_save_model.predict(features)[0])
    if predicted < 0:
        predicted = 0.0

    return SafeToSaveResponse(
        recommended_amount=round(predicted, 2),
        model_name="safe_to_save_rf_v1",
    )


def bucket_amount(amount: float) -> str:
    if amount < 5000:
        return "small"
    if amount < 50000:
        return "medium"
    return "large"