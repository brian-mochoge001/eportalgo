-- eportalgo/db/schema/13_early_warning_system.sql

CREATE TYPE risk_level AS ENUM ('Low', 'Medium', 'High');

CREATE TABLE student_risk_scores (
  risk_score_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  attendance_rate   DECIMAL(5, 2) NOT NULL DEFAULT 100.00,
  average_grade     DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
  risk_score        INT NOT NULL DEFAULT 0, -- 0 to 100
  risk_level        risk_level NOT NULL DEFAULT 'Low',
  last_calculated   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(student_id)
);

CREATE INDEX idx_student_risk_scores_school_id ON student_risk_scores(school_id);
CREATE INDEX idx_student_risk_scores_risk_level ON student_risk_scores(risk_level);
