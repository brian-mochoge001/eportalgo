-- eportalgo/db/schema/05_saas_billing.sql

CREATE TABLE school_settings (
  setting_id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id                 UUID UNIQUE NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  branding_logo_url         TEXT,
  branding_colors           JSONB NOT NULL DEFAULT '{ "primary": "#1a2b3c", "secondary": "#d1e2f3" }',
  timezone                  TEXT NOT NULL DEFAULT 'UTC',
  preferences               JSONB NOT NULL DEFAULT '{}',
  email_template_configs    JSONB NOT NULL DEFAULT '{}',
  payment_providers         JSONB,
  enable_strict_progression BOOLEAN NOT NULL DEFAULT FALSE,
  created_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subscription_tiers (
  tier_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tier_name         TEXT UNIQUE NOT NULL,
  description       TEXT,
  price_per_quarter DECIMAL(10, 2) NOT NULL,
  currency          TEXT NOT NULL DEFAULT 'USD',
  features          JSONB,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE school_subscriptions (
  subscription_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  tier_id           UUID NOT NULL REFERENCES subscription_tiers(tier_id) ON DELETE RESTRICT,
  start_date        DATE NOT NULL,
  end_date          DATE NOT NULL,
  status            TEXT NOT NULL,
  billing_cycle     TEXT NOT NULL, -- 'quarterly', 'annually'
  last_payment_date TIMESTAMPTZ(6),
  next_renewal_date DATE,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoices (
  invoice_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id       UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  subscription_id UUID REFERENCES school_subscriptions(subscription_id) ON DELETE SET NULL,
  invoice_number  TEXT UNIQUE NOT NULL,
  amount_due      DECIMAL(10, 2) NOT NULL,
  amount_paid     DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
  currency        TEXT NOT NULL DEFAULT 'USD',
  issue_date      DATE NOT NULL,
  due_date        DATE NOT NULL,
  status          TEXT NOT NULL,
  notes           TEXT,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
  transaction_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id             UUID NOT NULL REFERENCES invoices(invoice_id) ON DELETE CASCADE,
  school_id              UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  amount                 DECIMAL(10, 2) NOT NULL,
  transaction_date       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  payment_method         TEXT,
  gateway_transaction_id TEXT UNIQUE,
  status                 TEXT NOT NULL,
  notes                  TEXT,
  created_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE billing_contacts (
  billing_contact_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id          UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  name               TEXT NOT NULL,
  email              TEXT NOT NULL,
  phone_number       TEXT,
  role               TEXT,
  is_primary         BOOLEAN NOT NULL DEFAULT FALSE,
  created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, email)
);
