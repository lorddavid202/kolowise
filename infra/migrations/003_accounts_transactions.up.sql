CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_name TEXT NOT NULL,
    institution_name TEXT,
    account_type TEXT NOT NULL DEFAULT 'checking',
    current_balance_kobo BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'NGN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS accounts_user_id_idx
ON accounts(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount_kobo BIGINT NOT NULL CHECK (amount_kobo > 0),
    direction TEXT NOT NULL CHECK (direction IN ('credit', 'debit')),
    narration TEXT,
    merchant_name TEXT,
    category TEXT,
    txn_date TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS transactions_user_date_idx
ON transactions(user_id, txn_date DESC);

CREATE INDEX IF NOT EXISTS transactions_account_date_idx
ON transactions(account_id, txn_date DESC);