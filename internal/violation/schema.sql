-- violation service schema. Bootstrapped on startup (idempotent).

-- Plate registry: which member owns which plate (see CLAUDE.md decision).
CREATE TABLE IF NOT EXISTS plates (
    plate       TEXT PRIMARY KEY,
    owner_email TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS violations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plate              TEXT NOT NULL,
    violation_type     TEXT NOT NULL,
    location           TEXT NOT NULL,
    occurred_at        TIMESTAMPTZ NOT NULL,
    photo_object       TEXT,
    owner_email        TEXT,            -- resolved from the plate registry (nullable)
    issued_by_email    TEXT NOT NULL,

    -- Fine calculation snapshot — captured at issue time, NEVER recomputed.
    rule_version_id    UUID NOT NULL,
    rule_version       INTEGER NOT NULL,
    base_amount        BIGINT NOT NULL,
    time_multiplier    DOUBLE PRECISION NOT NULL,
    repeat_multiplier  DOUBLE PRECISION NOT NULL,
    prior_unpaid_count INTEGER NOT NULL,
    final_amount       BIGINT NOT NULL,

    payment_status     TEXT NOT NULL DEFAULT 'unpaid',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS violations_plate_idx ON violations (plate);
CREATE INDEX IF NOT EXISTS violations_owner_idx ON violations (owner_email);
