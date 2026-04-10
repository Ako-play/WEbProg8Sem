CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY,
    base_code VARCHAR(3) NOT NULL,
    target_code VARCHAR(3) NOT NULL,
    rate NUMERIC(18, 8) NOT NULL,
    rate_date DATE NOT NULL,
    source TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT exchange_rates_unique UNIQUE (base_code, target_code, rate_date)
);

CREATE INDEX IF NOT EXISTS exchange_rates_lookup_idx
    ON exchange_rates (base_code, target_code, rate_date DESC);
