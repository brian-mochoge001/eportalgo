-- eportalgo/db/schema/04_remainder.sql

-- 04_enrollment_attendance.sql
CREATE TABLE parent_student_relationships (
  relationship_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  parent_user_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  student_user_id   UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  relationship_type TEXT NOT NULL DEFAULT 'Parent/Guardian',
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(parent_user_id, student_user_id)
);

CREATE TABLE enrollments (
  enrollment_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id       UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  class_id        UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  enrollment_date DATE NOT NULL DEFAULT CURRENT_DATE,
  status          TEXT NOT NULL DEFAULT 'Enrolled',
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(student_id, class_id)
);

CREATE TABLE attendance_records (
  attendance_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id       UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  class_id        UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  attendance_date DATE NOT NULL,
  status          TEXT NOT NULL,
  notes           TEXT,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, student_id, class_id, attendance_date)
);

-- 05_curriculum_materials.sql
CREATE TABLE lesson_plans (
  lesson_plan_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id            UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  teacher_id           UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  class_id             UUID REFERENCES academic_classes(class_id) ON DELETE SET NULL,
  subject_id           UUID REFERENCES subjects(subject_id) ON DELETE SET NULL,
  title                TEXT NOT NULL,
  topic                TEXT,
  lesson_number        INT,
  objectives           TEXT[],
  content              TEXT,
  assessment_questions TEXT[],
  online_meeting_link  TEXT,
  date_covered         DATE,
  created_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at           TIMESTAMPTZ(6)
);

CREATE TABLE learning_materials (
  material_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  uploaded_by_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  class_id            UUID REFERENCES academic_classes(class_id) ON DELETE SET NULL,
  course_id           UUID REFERENCES courses(course_id) ON DELETE SET NULL,
  title               TEXT NOT NULL,
  description         TEXT,
  file_url            TEXT,
  content             TEXT,
  material_type       material_type NOT NULL,
  external_data       JSONB,
  uploaded_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lesson_plan_learning_materials (
  lesson_plan_id UUID REFERENCES lesson_plans(lesson_plan_id) ON DELETE CASCADE,
  material_id    UUID REFERENCES learning_materials(material_id) ON DELETE CASCADE,
  PRIMARY KEY (lesson_plan_id, material_id)
);

-- 06_assessments.sql
CREATE TABLE grading_systems (
  grading_system_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  class_id          UUID UNIQUE NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  description       TEXT,
  created_by        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE grade_categories (
  grade_category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  grading_system_id UUID NOT NULL REFERENCES grading_systems(grading_system_id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  weight            DECIMAL(5, 2) NOT NULL,
  UNIQUE(grading_system_id, name)
);

CREATE TABLE assignments (
  assignment_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  class_id          UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  teacher_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  title             TEXT NOT NULL,
  description       TEXT,
  due_date          TIMESTAMPTZ(6),
  max_score         DECIMAL(5, 2) NOT NULL,
  assignment_type   TEXT NOT NULL,
  file_url          TEXT,
  quiz_id           UUID UNIQUE, -- Referenced in 07_exams_quizzes.sql
  grade_category_id UUID REFERENCES grade_categories(grade_category_id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE submissions (
  submission_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id          UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id         UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  assignment_id      UUID NOT NULL REFERENCES assignments(assignment_id) ON DELETE CASCADE,
  submission_content TEXT,
  submitted_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status             TEXT NOT NULL DEFAULT 'Submitted',
  created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(student_id, assignment_id)
);

CREATE TABLE grades (
  grade_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  submission_id     UUID UNIQUE NOT NULL REFERENCES submissions(submission_id) ON DELETE CASCADE,
  graded_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  score             DECIMAL(5, 2) NOT NULL,
  feedback          TEXT,
  graded_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 07_exams_quizzes.sql
CREATE TABLE quizzes (
  quiz_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id    UUID UNIQUE REFERENCES assignments(assignment_id) ON DELETE CASCADE,
  school_id        UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  teacher_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  subject_id       UUID REFERENCES subjects(subject_id) ON DELETE SET NULL,
  title            TEXT NOT NULL,
  description      TEXT,
  quiz_type        TEXT NOT NULL,
  duration_minutes INT,
  start_time       TIMESTAMPTZ(6),
  end_time         TIMESTAMPTZ(6),
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE exams (
  exam_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  subject_id          UUID NOT NULL REFERENCES subjects(subject_id) ON DELETE CASCADE,
  teacher_id          UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  title               TEXT NOT NULL,
  description         TEXT,
  exam_type           exam_type NOT NULL,
  is_online           BOOLEAN NOT NULL DEFAULT FALSE,
  online_exam_status  online_exam_status NOT NULL DEFAULT 'NOT_APPLICABLE',
  approved_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  duration_minutes    INT,
  start_time          TIMESTAMPTZ(6),
  end_time            TIMESTAMPTZ(6),
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE questions (
  question_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quiz_id       UUID REFERENCES quizzes(quiz_id) ON DELETE CASCADE,
  exam_id       UUID REFERENCES exams(exam_id) ON DELETE CASCADE,
  question_text TEXT NOT NULL,
  question_type TEXT NOT NULL,
  "order"       INT NOT NULL,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE options (
  option_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id UUID NOT NULL REFERENCES questions(question_id) ON DELETE CASCADE,
  option_text TEXT NOT NULL,
  is_correct  BOOLEAN NOT NULL,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 08_finance.sql
CREATE TABLE fee_structures (
  fee_structure_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id        UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  fee_name         TEXT NOT NULL,
  amount           DECIMAL(10, 2) NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'USD',
  academic_year    TEXT NOT NULL,
  description      TEXT,
  is_active        BOOLEAN NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, fee_name, academic_year)
);

CREATE TABLE student_fees (
  student_fee_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id        UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  fee_structure_id UUID NOT NULL REFERENCES fee_structures(fee_structure_id) ON DELETE CASCADE,
  amount_due       DECIMAL(10, 2) NOT NULL,
  amount_paid      DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
  due_date         DATE,
  status           TEXT NOT NULL DEFAULT 'Pending',
  notes            TEXT,
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payments (
  payment_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  student_fee_id      UUID NOT NULL REFERENCES student_fees(student_fee_id) ON DELETE CASCADE,
  amount              DECIMAL(10, 2) NOT NULL,
  payment_date        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  payment_method      TEXT,
  transaction_id      TEXT UNIQUE,
  recorded_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  notes               TEXT,
  receipt_number      TEXT UNIQUE,
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
