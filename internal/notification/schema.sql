-- notification service schema. Bootstrapped on startup (idempotent).
CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_email TEXT NOT NULL,
    kind            TEXT NOT NULL,   -- violation_issued | payment_completed
    message         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notifications_recipient_idx ON notifications (recipient_email, created_at DESC);
