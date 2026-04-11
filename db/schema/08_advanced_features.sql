-- eportalgo/db/schema/08_advanced_features.sql

-- Badges & Gamification
CREATE TABLE badges (
  badge_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id      UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  badge_name     TEXT NOT NULL,
  description    TEXT,
  icon_url       TEXT,
  criteria       TEXT NOT NULL,
  created_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, badge_name)
);

CREATE TABLE student_badges (
  student_badge_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id          UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id         UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  badge_id           UUID NOT NULL REFERENCES badges(badge_id) ON DELETE CASCADE,
  awarded_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  awarded_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  notes              TEXT,
  created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(student_id, badge_id)
);

CREATE TABLE badge_courses (
  badge_course_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  badge_id        UUID UNIQUE NOT NULL REFERENCES badges(badge_id) ON DELETE CASCADE,
  school_id       UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  teacher_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  title           TEXT NOT NULL,
  description     TEXT,
  is_free         BOOLEAN NOT NULL DEFAULT TRUE,
  completion_type TEXT NOT NULL,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Short Courses
CREATE TABLE short_course_enrollments (
  enrollment_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id              UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id             UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  course_id              UUID NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
  enrollment_date        DATE NOT NULL DEFAULT CURRENT_DATE,
  status                 TEXT NOT NULL DEFAULT 'Enrolled',
  attempt_type           TEXT NOT NULL DEFAULT 'first_attempt',
  previous_enrollment_id UUID UNIQUE REFERENCES short_course_enrollments(enrollment_id),
  created_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(student_id, course_id)
);

-- Transfers
CREATE TABLE transfer_requests (
  transfer_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type           TEXT NOT NULL, -- 'Student', 'Teacher', 'Staff'
  entity_id             UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  source_school_id      UUID NOT NULL REFERENCES schools(school_id),
  destination_school_id UUID NOT NULL REFERENCES schools(school_id),
  initiated_by_user_id  UUID NOT NULL REFERENCES users(user_id),
  status                TEXT NOT NULL DEFAULT 'pending',
  request_date          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completion_date       TIMESTAMPTZ(6),
  notes                 TEXT,
  created_at            TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at            TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Health/Clinic
CREATE TABLE clinic_visits (
  visit_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id  UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  nurse_id   UUID REFERENCES users(user_id) ON DELETE SET NULL,
  visit_date TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  symptoms   TEXT,
  diagnosis  TEXT,
  treatment  TEXT,
  notes      TEXT,
  created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Audit Logging
CREATE TABLE audit_logs (
  log_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id   UUID REFERENCES schools(school_id) ON DELETE CASCADE,
  user_id     UUID REFERENCES users(user_id) ON DELETE SET NULL,
  action      TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id   UUID,
  old_value   JSONB,
  new_value   JSONB,
  ip_address  TEXT,
  user_agent  TEXT,
  logged_at   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
