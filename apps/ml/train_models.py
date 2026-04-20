from pathlib import Path
import random
import joblib
import numpy as np
import pandas as pd
from sklearn.pipeline import Pipeline
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.ensemble import RandomForestRegressor

BASE_DIR = Path(__file__).resolve().parent
ARTIFACTS_DIR = BASE_DIR / "artifacts"
ARTIFACTS_DIR.mkdir(exist_ok=True)


def build_category_training_data() -> pd.DataFrame:
    records = [
        {"narration": "Salary payment", "merchant_name": "Employer", "direction": "credit", "amount": 50000, "category": "income"},
        {"narration": "Monthly salary", "merchant_name": "Company", "direction": "credit", "amount": 120000, "category": "income"},
        {"narration": "Uber trip", "merchant_name": "Uber", "direction": "debit", "amount": 3500, "category": "transport"},
        {"narration": "Bolt ride", "merchant_name": "Bolt", "direction": "debit", "amount": 2700, "category": "transport"},
        {"narration": "Fuel purchase", "merchant_name": "Mobil", "direction": "debit", "amount": 18000, "category": "transport"},
        {"narration": "Lunch", "merchant_name": "Chicken Republic", "direction": "debit", "amount": 4500, "category": "food"},
        {"narration": "Dinner", "merchant_name": "KFC", "direction": "debit", "amount": 6200, "category": "food"},
        {"narration": "Groceries", "merchant_name": "Shoprite", "direction": "debit", "amount": 22000, "category": "groceries"},
        {"narration": "Market shopping", "merchant_name": "Market", "direction": "debit", "amount": 9000, "category": "groceries"},
        {"narration": "Electricity token", "merchant_name": "IKEDC", "direction": "debit", "amount": 15000, "category": "utilities"},
        {"narration": "Water bill", "merchant_name": "Water Corp", "direction": "debit", "amount": 3500, "category": "utilities"},
        {"narration": "Internet subscription", "merchant_name": "MTN", "direction": "debit", "amount": 6000, "category": "airtime_data"},
        {"narration": "Data bundle", "merchant_name": "Airtel", "direction": "debit", "amount": 4000, "category": "airtime_data"},
        {"narration": "Airtime recharge", "merchant_name": "Glo", "direction": "debit", "amount": 1000, "category": "airtime_data"},
        {"narration": "Rent payment", "merchant_name": "Landlord", "direction": "debit", "amount": 250000, "category": "rent"},
        {"narration": "House rent", "merchant_name": "Property Agent", "direction": "debit", "amount": 300000, "category": "rent"},
        {"narration": "Pharmacy purchase", "merchant_name": "MedsPlus", "direction": "debit", "amount": 5000, "category": "health"},
        {"narration": "Hospital bill", "merchant_name": "Hospital", "direction": "debit", "amount": 45000, "category": "health"},
        {"narration": "Transfer from friend", "merchant_name": "Bank Transfer", "direction": "credit", "amount": 15000, "category": "transfer_in"},
        {"narration": "Bank transfer", "merchant_name": "Mobile App", "direction": "debit", "amount": 10000, "category": "transfer_out"},
        {"narration": "POS withdrawal", "merchant_name": "POS", "direction": "debit", "amount": 20000, "category": "cash_withdrawal"},
        {"narration": "ATM withdrawal", "merchant_name": "ATM", "direction": "debit", "amount": 10000, "category": "cash_withdrawal"},
        {"narration": "Netflix subscription", "merchant_name": "Netflix", "direction": "debit", "amount": 5500, "category": "entertainment"},
        {"narration": "Cinema ticket", "merchant_name": "Filmhouse", "direction": "debit", "amount": 8000, "category": "entertainment"},
        {"narration": "Book purchase", "merchant_name": "RovingHeights", "direction": "debit", "amount": 12000, "category": "education"},
        {"narration": "School fees", "merchant_name": "University", "direction": "debit", "amount": 180000, "category": "education"},
        {"narration": "Goal contribution", "merchant_name": "", "direction": "debit", "amount": 25000, "category": "goal_savings"},
        {"narration": "Savings contribution", "merchant_name": "", "direction": "debit", "amount": 15000, "category": "goal_savings"},
    ]
    return pd.DataFrame(records)


def combine_text(df: pd.DataFrame) -> pd.Series:
    amount_bucket = df["amount"].apply(
        lambda x: "small" if x < 5000 else "medium" if x < 50000 else "large"
    )
    return (
        df["narration"].fillna("").astype(str)
        + " "
        + df["merchant_name"].fillna("").astype(str)
        + " "
        + df["direction"].fillna("").astype(str)
        + " "
        + amount_bucket.astype(str)
    )


def train_category_model():
    df = build_category_training_data()
    X = combine_text(df)
    y = df["category"]

    pipeline = Pipeline(
        steps=[
            ("tfidf", TfidfVectorizer(ngram_range=(1, 2))),
            ("clf", LogisticRegression(max_iter=2000)),
        ]
    )
    pipeline.fit(X, y)

    joblib.dump(pipeline, ARTIFACTS_DIR / "category_pipeline.joblib")
    print("Saved category model")


def generate_safe_to_save_data(n: int = 3000) -> pd.DataFrame:
    rows = []
    random.seed(42)
    np.random.seed(42)

    for _ in range(n):
        available_balance = random.uniform(5000, 2000000)
        avg_monthly_income = random.uniform(30000, 1200000)
        avg_monthly_expense = random.uniform(10000, avg_monthly_income * 0.95)
        emergency_buffer = max(avg_monthly_expense * 0.2, 10000)
        monthly_surplus = max(avg_monthly_income - avg_monthly_expense, 0)
        active_goal_need = random.uniform(0, 1000000)

        base = min(max((available_balance - emergency_buffer) * 0.5, 0), monthly_surplus * 0.4)
        target = min(base, active_goal_need) if active_goal_need > 0 else base
        noise = random.uniform(-2000, 2000)
        recommended_amount = max(target + noise, 0)

        rows.append(
            {
                "available_balance": available_balance,
                "avg_monthly_income": avg_monthly_income,
                "avg_monthly_expense": avg_monthly_expense,
                "emergency_buffer": emergency_buffer,
                "monthly_surplus": monthly_surplus,
                "active_goal_need": active_goal_need,
                "recommended_amount": recommended_amount,
            }
        )

    return pd.DataFrame(rows)


def train_safe_to_save_model():
    df = generate_safe_to_save_data()
    X = df[
        [
            "available_balance",
            "avg_monthly_income",
            "avg_monthly_expense",
            "emergency_buffer",
            "monthly_surplus",
            "active_goal_need",
        ]
    ]
    y = df["recommended_amount"]

    model = RandomForestRegressor(
        n_estimators=200,
        random_state=42,
        n_jobs=-1,
    )
    model.fit(X, y)

    joblib.dump(model, ARTIFACTS_DIR / "safe_to_save_model.joblib")
    print("Saved safe-to-save model")


if __name__ == "__main__":
    train_category_model()
    train_safe_to_save_model()
    print("All models trained successfully")