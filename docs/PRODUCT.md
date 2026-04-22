# Product Documentation

KoloWise is a personal finance and savings intelligence application that helps users track money movement and decide how much they can safely save.

## Product Vision

Enable users to move from reactive money tracking to proactive savings decisions using:

1. Reliable transaction capture
2. Clear savings-goal progress
3. Explainable recommendation engine

## Primary User

A digitally active individual who:

- Has one or more bank/wallet accounts
- Wants to track inflows/outflows
- Wants to save consistently without risking short-term liquidity

## Core Value Proposition

- Single place for balances, transactions, and savings goals
- ML-assisted transaction categorization
- ML-assisted savings recommendation with transparent rule-based safeguards

## Product Modules

### 1) Authentication

- Register (API)
- Login (web)
- Current-user identity lookup
- JWT-based session for protected pages

### 2) Accounts

- Create account (name, institution, type, opening balance, currency)
- View accounts and balances

### 3) Transactions

- Manual transaction entry
- CSV statement import
- Category auto-fill when missing (ML + fallback)
- Transaction filtering by search text, direction, account, and category

### 4) Goals

- Create savings goals with target amount/date and priority
- Contribute from selected account
- Auto-update goal completion status
- View contribution history per goal

### 5) Insights

- Compute safe-to-save recommendation from user financial signals
- Show engine used (`ml` or `rule_based`)
- Show recommendation reason/explanation

### 6) Dashboard

- Total balance summary
- Accounts count
- Active/completed goals metrics
- Spending by category chart
- Recent cashflow chart
- Recent transactions feed

## End-to-End User Journey

### First-time user

1. Register account via API endpoint (`/api/v1/auth/register`)
2. Login in UI (`/`)
3. Create at least one account
4. Add manual transactions or upload CSV
5. Create goals
6. Contribute toward goals
7. Check dashboard and insights recommendation

### Returning user

1. Login
2. Add new transactions
3. Review category trends and goal progress
4. Refresh insights and decide savings amount

## Screens

Screenshots are available in `docs/screenshots`:

- `login.png`
- `dashboard.png`
- `accounts.png`
- `transactions.png`
- `goals.png`
- `insights.png`

## Recommendation UX Principle

The recommendation feature is intentionally conservative:

1. Rule engine computes safe baseline first.
2. ML proposes a value.
3. Final recommendation is capped by rule guardrails.

This prevents aggressive savings suggestions when cash safety is uncertain.

## Data and Currency Model

- Monetary storage: integer kobo in backend tables
- API also returns string-decimal values for display
- Default and primary currency in implementation: `NGN`

## Current Implementation Constraints

1. Web app has login page but no registration page (registration is API-only).
2. No password reset/profile management flows yet.
3. No account/transaction/goal edit-delete UI workflows.
4. No multi-user household or shared-goal collaboration features.
5. No notifications/reminders for contributions or deadlines.
6. No explicit budget planning module yet.

## Success Signals (Recommended)

Track these product metrics as next instrumentation steps:

1. Weekly active users
2. Transaction ingestion rate (manual + CSV)
3. Goal creation rate
4. Goal completion rate
5. Recommendation acceptance rate (whether users contribute after insight refresh)

## Product Roadmap Suggestions

1. Add in-app registration and onboarding checklist.
2. Add recurring contribution automation.
3. Add transaction edit/correct category workflow with feedback loop into ML.
4. Add budget envelopes and monthly spend targets.
5. Add push/email reminders for goal deadlines and anomalies.
6. Add model quality dashboard and per-user recommendation confidence indicators.
