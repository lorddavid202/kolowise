from pydantic import BaseModel


class CategoryPredictionRequest(BaseModel):
    narration: str = ""
    merchant_name: str = ""
    amount: float
    direction: str


class CategoryPredictionResponse(BaseModel):
    category: str
    confidence: float
    model_name: str


class SafeToSaveRequest(BaseModel):
    available_balance: float
    avg_monthly_income: float
    avg_monthly_expense: float
    emergency_buffer: float
    monthly_surplus: float
    active_goal_need: float


class SafeToSaveResponse(BaseModel):
    recommended_amount: float
    model_name: str