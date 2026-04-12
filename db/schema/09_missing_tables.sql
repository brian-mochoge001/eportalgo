-- eportalgo/db/schema/09_missing_tables.sql

CREATE TABLE events (
  event_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id   UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  event_date  TIMESTAMPTZ(6) NOT NULL,
  end_date    TIMESTAMPTZ(6),
  location    TEXT,
  event_type  TEXT NOT NULL,
  organizer_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  is_public   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at  TIMESTAMPTZ(6)
);

CREATE TABLE meetings (
  meeting_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id        UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  title            TEXT NOT NULL,
  agenda           TEXT,
  meeting_date     TIMESTAMPTZ(6) NOT NULL,
  duration_minutes INT,
  location         TEXT,
  meeting_type     TEXT NOT NULL,
  organizer_id     UUID REFERENCES users(user_id) ON DELETE SET NULL,
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at       TIMESTAMPTZ(6)
);

CREATE TABLE meeting_attendees (
  meeting_id UUID NOT NULL REFERENCES meetings(meeting_id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  school_id  UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  PRIMARY KEY (meeting_id, user_id)
);

CREATE TABLE online_class_sessions (
  session_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id       UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  class_id        UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  teacher_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  session_title   TEXT NOT NULL,
  start_time      TIMESTAMPTZ(6) NOT NULL,
  end_time        TIMESTAMPTZ(6) NOT NULL,
  meeting_link    TEXT NOT NULL,
  description     TEXT,
  recording_link  TEXT,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE external_certifications (
  cert_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  name             TEXT NOT NULL,
  issuer           TEXT NOT NULL,
  credential_id    TEXT UNIQUE,
  verification_url TEXT,
  issue_date       DATE,
  expiry_date      DATE,
  is_verified      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at       TIMESTAMPTZ(6)
);

CREATE TABLE quiz_submissions (
  submission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quiz_id       UUID NOT NULL REFERENCES quizzes(quiz_id) ON DELETE CASCADE,
  student_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  score         DECIMAL(5, 2),
  submitted_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status        TEXT NOT NULL DEFAULT 'completed',
  UNIQUE(quiz_id, student_id),
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE quiz_answers (
  answer_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quiz_submission_id  UUID NOT NULL REFERENCES quiz_submissions(submission_id) ON DELETE CASCADE,
  question_id         UUID NOT NULL REFERENCES questions(question_id) ON DELETE CASCADE,
  student_answer_text TEXT,
  selected_option_id  UUID REFERENCES options(option_id) ON DELETE SET NULL,
  is_correct          BOOLEAN,
  UNIQUE(quiz_submission_id, question_id),
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE short_course_grades (
  grade_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  enrollment_id     UUID UNIQUE NOT NULL REFERENCES short_course_enrollments(enrollment_id) ON DELETE CASCADE,
  course_id         UUID NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
  student_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  score             DECIMAL(5, 2) NOT NULL,
  feedback          TEXT,
  graded_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  graded_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE student_course_progress (
  progress_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  enrollment_id       UUID UNIQUE NOT NULL REFERENCES enrollments(enrollment_id) ON DELETE CASCADE,
  progress_percentage DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
  last_activity_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE transcripts (
  transcript_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  academic_year     TEXT NOT NULL,
  cumulative_gpa    DECIMAL(4, 2),
  transcript_data   JSONB NOT NULL DEFAULT '{}',
  issued_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  issued_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE teacher_availability (
  availability_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  teacher_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  day_of_week     INT NOT NULL, -- 0 (Sunday) to 6 (Saturday)
  start_time      TIME NOT NULL,
  end_time        TIME NOT NULL,
  is_recurring    BOOLEAN NOT NULL DEFAULT TRUE,
  notes           TEXT,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE teacher_workloads (
  workload_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  teacher_id             UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  max_hours_per_week     DECIMAL(5, 2) NOT NULL,
  current_hours_per_week DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
  created_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE timetables (
  timetable_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  academic_year TEXT NOT NULL,
  semester      TEXT,
  title         TEXT NOT NULL,
  description   TEXT,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);
