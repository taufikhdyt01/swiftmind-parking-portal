-- payment service schema. Bootstrapped on startup (idempotent).
CREATE TABLE IF NOT EXISTS invoices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    violation_id   UUID NOT NULL UNIQUE,
    plate          TEXT NOT NULL,
    violation_type TEXT NOT NULL,
    owner_email      TEXT,
    issued_by_email  TEXT,
    amount         BIGINT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'open',   -- open | paid
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS issued_by_email TEXT;

CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id     UUID NOT NULL REFERENCES invoices (id),
    amount         BIGINT NOT NULL,
    scenario       TEXT NOT NULL,                  -- success | failed (test input)
    status         TEXT NOT NULL,                  -- paid | failed
    transaction_id TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS invoices_owner_idx ON invoices (owner_email);
CREATE INDEX IF NOT EXISTS payments_invoice_idx ON payments (invoice_id);
