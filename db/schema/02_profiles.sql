-- eportalgo/db/schema/02_profiles.sql

CREATE TABLE saas_company_profiles (
  saas_profile_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id              UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  job_title            TEXT NOT NULL,
  department_name_text TEXT,
  internal_employee_id TEXT UNIQUE,
  hire_date            DATE,
  created_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE school_staff_profiles (
  staff_profile_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id              UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  school_id            UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  employee_id          TEXT UNIQUE NOT NULL,
  department_name_text TEXT,
  designation          TEXT NOT NULL,
  hire_date            DATE,
  office_location      TEXT,
  department_id        UUID, -- Will be referenced in 03_academic_structure.sql
  created_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, employee_id)
);

CREATE TABLE teacher_profiles (
  teacher_profile_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                   UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  school_id                 UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  teacher_id_number         TEXT UNIQUE,
  specialization            TEXT,
  qualification             TEXT,
  employment_start_date     DATE,
  is_class_teacher          BOOLEAN NOT NULL DEFAULT FALSE,
  class_teacher_of_class_id UUID, -- Will be referenced in 03_academic_structure.sql
  department_name_text      TEXT,
  department_id             UUID, -- Will be referenced in 03_academic_structure.sql
  created_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE student_profiles (
  student_profile_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id             UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  enrollment_number   TEXT UNIQUE NOT NULL,
  current_grade_level TEXT,
  admission_date      DATE NOT NULL,
  current_class_id    UUID, -- Will be referenced in 03_academic_structure.sql
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, enrollment_number)
);

CREATE TABLE parent_profiles (
  parent_profile_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                 UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  school_id               UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  home_address            TEXT,
  occupation              TEXT,
  emergency_contact_name  TEXT,
  emergency_contact_phone TEXT,
  created_at              TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at              TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE health_records (
  record_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id         UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  blood_group        TEXT,
  allergies          TEXT,
  medical_conditions TEXT,
  emergency_notes    TEXT,
  created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
