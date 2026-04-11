-- eportalgo/db/schema/01_core.sql

CREATE TABLE schools (
  school_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_name     TEXT UNIQUE NOT NULL,
  subdomain       TEXT UNIQUE,
  status          TEXT NOT NULL DEFAULT 'pending',
  school_initial  TEXT UNIQUE,
  address         TEXT,
  phone_number    TEXT,
  email           TEXT,
  logo_url        TEXT,
  primary_color   TEXT,
  secondary_color TEXT,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE roles (
  role_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_name      TEXT UNIQUE NOT NULL,
  description    TEXT,
  is_school_role BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE users (
  user_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID REFERENCES schools(school_id) ON DELETE CASCADE,
  role_id             UUID NOT NULL REFERENCES roles(role_id) ON DELETE RESTRICT,
  first_name          TEXT NOT NULL,
  last_name           TEXT NOT NULL,
  email               TEXT UNIQUE NOT NULL,
  contact_email       TEXT,
  firebase_uid        TEXT UNIQUE,
  password_hash       TEXT,
  phone_number        TEXT,
  date_of_birth       DATE,
  gender              TEXT,
  profile_picture_url TEXT,
  is_active           BOOLEAN NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE applications (
  application_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id                  UUID REFERENCES schools(school_id) ON DELETE SET NULL,
  applicant_user_id          UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  applicant_first_name       TEXT NOT NULL,
  applicant_last_name        TEXT NOT NULL,
  applicant_email            TEXT NOT NULL,
  school_name_at_application TEXT NOT NULL,
  desired_role               TEXT NOT NULL,
  status                     TEXT NOT NULL DEFAULT 'Pending',
  notes                      TEXT,
  created_at                 TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                 TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
