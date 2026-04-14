-- eportalgo/db/schema/14_banners.sql

CREATE TABLE banners (
  banner_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id   UUID REFERENCES schools(school_id) ON DELETE CASCADE, -- NULL means global banner
  title       TEXT,
  image_url   TEXT NOT NULL,
  target_url  TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  "order"     INT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_banners_school_id ON banners(school_id);
