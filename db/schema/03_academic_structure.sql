-- eportalgo/db/schema/03_academic_structure.sql

CREATE TABLE departments (
  department_id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id                    UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  department_name              TEXT NOT NULL,
  head_of_department_id        UUID REFERENCES users(user_id) ON DELETE SET NULL,
  deputy_head_of_department_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  created_at                   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, department_name)
);

CREATE TABLE rooms (
  room_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  room_name     TEXT NOT NULL,
  capacity      INT NOT NULL,
  room_type     room_type NOT NULL,
  department_id UUID REFERENCES departments(department_id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, room_name)
);

CREATE TABLE subjects (
  subject_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id              UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  subject_name           TEXT NOT NULL,
  description            TEXT,
  double_period_required BOOLEAN NOT NULL DEFAULT FALSE,
  lab_period_required    BOOLEAN NOT NULL DEFAULT FALSE,
  max_online_percentage  DECIMAL(5, 2),
  created_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, subject_name)
);

CREATE TABLE courses (
  course_id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id                 UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  course_code               TEXT NOT NULL,
  course_name               TEXT NOT NULL,
  description               TEXT,
  is_short_course           BOOLEAN NOT NULL DEFAULT FALSE,
  price                     DECIMAL(10, 2),
  is_graded_independently   BOOLEAN NOT NULL DEFAULT FALSE,
  requires_all_units_passed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(school_id, course_code)
);

CREATE TABLE academic_classes (
  class_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id             UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  course_id             UUID NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
  teacher_id            UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  class_name            TEXT NOT NULL,
  academic_year         TEXT NOT NULL,
  semester              TEXT,
  start_date            DATE,
  end_date              DATE,
  attendance_priority   TEXT NOT NULL DEFAULT 'normal',
  created_at            TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at            TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Many-to-Many Relationships
CREATE TABLE teacher_subjects (
  teacher_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  subject_id UUID NOT NULL REFERENCES subjects(subject_id) ON DELETE CASCADE,
  PRIMARY KEY (teacher_id, subject_id)
);

CREATE TABLE department_subjects (
  department_id UUID NOT NULL REFERENCES departments(department_id) ON DELETE CASCADE,
  subject_id    UUID NOT NULL REFERENCES subjects(subject_id) ON DELETE CASCADE,
  PRIMARY KEY (department_id, subject_id)
);

CREATE TABLE course_subjects (
  course_id  UUID NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
  subject_id UUID NOT NULL REFERENCES subjects(subject_id) ON DELETE CASCADE,
  PRIMARY KEY (course_id, subject_id)
);

-- Back-references from 02_profiles.sql
ALTER TABLE school_staff_profiles ADD CONSTRAINT fk_staff_department FOREIGN KEY (department_id) REFERENCES departments(department_id) ON DELETE SET NULL;
ALTER TABLE teacher_profiles ADD CONSTRAINT fk_teacher_department FOREIGN KEY (department_id) REFERENCES departments(department_id) ON DELETE SET NULL;
ALTER TABLE teacher_profiles ADD CONSTRAINT fk_teacher_class FOREIGN KEY (class_teacher_of_class_id) REFERENCES academic_classes(class_id) ON DELETE SET NULL;
ALTER TABLE student_profiles ADD CONSTRAINT fk_student_class FOREIGN KEY (current_class_id) REFERENCES academic_classes(class_id) ON DELETE SET NULL;

CREATE TABLE class_representatives (
  class_rep_id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_user_id                 UUID NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
  academic_class_id               UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  can_communicate_teacher         BOOLEAN NOT NULL DEFAULT FALSE,
  can_communicate_department_head BOOLEAN NOT NULL DEFAULT FALSE,
  created_at                      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
