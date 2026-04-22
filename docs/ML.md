# KoloWise ML Documentation

This document explains the ML subsystem used by KoloWise for transaction categorization and safe-to-save recommendations.

## Scope

The ML layer provides two online predictions:

1. Transaction category classification
2. Safe-to-save amount regression

The Go API consumes these predictions and applies business guardrails.

## Service Location

- Service code: `apps/ml/app`
- Training code: `apps/ml/train_models.py`
- Artifacts: `apps/ml/artifacts/*.joblib`
- Docker image build retrains models during build (`python train_models.py`)

## Runtime API (FastAPI)

### Health

- `GET /healthz`
- Returns service status and whether model artifacts are loaded.

### Category Prediction

- `POST /predict/category`

Request:

```json
{
  "narration": "Uber trip",
  "merchant_name": "Uber",
  "amount": 3500,
  "direction": "debit"
}
```

Response:

```json
{
  "category": "transport",
  "confidence": 0.88,
  "model_name": "transaction_category_logreg_v1"
}
```

### Safe-to-Save Prediction

- `POST /predict/safe-to-save`

Request:

```json
{
  "available_balance": 100000,
  "avg_monthly_income": 250000,
  "avg_monthly_expense": 160000,
  "emergency_buffer": 32000,
  "monthly_surplus": 90000,
  "active_goal_need": 100000
}
```

Response:

```json
{
  "recommended_amount": 4200,
  "model_name": "safe_to_save_rf_v1"
}
```

## Model Artifacts

Artifacts are loaded at service startup from:

- `apps/ml/artifacts/category_pipeline.joblib`
- `apps/ml/artifacts/safe_to_save_model.joblib`

If files are missing, startup fails with `FileNotFoundError`.

## Model 1: Transaction Category Classifier

### Problem Type

- Multi-class text classification

### Training Data

- Small curated dataset in `build_category_training_data()`
- Includes local finance categories such as:
  - `income`
  - `transport`
  - `food`
  - `groceries`
  - `utilities`
  - `airtime_data`
  - `rent`
  - `health`
  - `transfer_in` / `transfer_out`
  - `cash_withdrawal`
  - `entertainment`
  - `education`
  - `goal_savings`

### Feature Engineering

Input fields are merged into a text feature:

1. narration
2. merchant name
3. direction
4. amount bucket (`small`, `medium`, `large`)

### Algorithm

Scikit-learn pipeline:

1. `TfidfVectorizer(ngram_range=(1,2))`
2. `LogisticRegression(max_iter=2000)`

### Inference Notes

- Confidence uses max class probability when `predict_proba` is available.
- Confidence is rounded to 4 decimals.

## Model 2: Safe-to-Save Regressor

### Problem Type

- Regression (continuous amount output)

### Training Data

- Synthetic data generated in `generate_safe_to_save_data(n=3000)`
- Random seed fixed (`42`) for reproducibility

Input features:

1. `available_balance`
2. `avg_monthly_income`
3. `avg_monthly_expense`
4. `emergency_buffer`
5. `monthly_surplus`
6. `active_goal_need`

Target generation uses rule-like logic plus controlled random noise.

### Algorithm

- `RandomForestRegressor(n_estimators=200, random_state=42, n_jobs=-1)`

### Inference Notes

- Negative predictions are clamped to zero.
- Output is rounded to 2 decimal places.

## Training Workflow

Run locally:

```bash
cd apps/ml
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python train_models.py
```

Expected output:

- `Saved category model`
- `Saved safe-to-save model`
- `All models trained successfully`

## Integration in Go API

The Go API calls ML through `internal/mlclient`:

- `PredictCategory(...)` used during transaction create/import when category is missing
- `PredictSafeToSave(...)` used by `/insights/safe-to-save`

## Guardrails and Fallbacks

### Category fallback

- If ML call fails or returns empty category, API uses `uncategorized`.

### Safe-to-save fallback

- API always computes rule-based baseline first.
- If ML prediction succeeds:
  1. Convert ML recommendation to kobo.
  2. Clamp to zero.
  3. Cap by rule-based recommendation (`safeCap`).
- If ML fails, API returns pure rule-based result.

This ensures ML cannot over-recommend savings beyond deterministic safety bounds.

## Operational Notes

- API ML client timeout: 10 seconds
- No async queue; calls are synchronous in request path
- Model names are returned in responses for transparency

## Known Limitations

- Category training dataset is small and handcrafted
- Safe-to-save model is trained on synthetic, not historical user-labeled data
- No model version registry or drift monitoring pipeline yet
- No feature store; features are assembled at request time

## Recommended Next Improvements

1. Add evaluation metrics output during training (accuracy, F1, MAE/RMSE).
2. Introduce dataset versioning and experiment tracking.
3. Expand category corpus using anonymized real transaction patterns.
4. Add model validation checks before artifact promotion.
5. Add request/response telemetry for inference quality monitoring.
