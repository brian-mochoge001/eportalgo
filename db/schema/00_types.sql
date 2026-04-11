-- eportalgo/db/schema/00_types.sql

CREATE TYPE material_type AS ENUM (
  'FILE',
  'URL',
  'CODE_SNIPPET',
  'GOOGLE_BOOK'
);

CREATE TYPE notification_type AS ENUM (
  'SYSTEM_UPDATE',
  'ANNOUNCEMENT',
  'MESSAGE',
  'ASSIGNMENT_SUBMITTED',
  'ASSIGNMENT_GRADED',
  'SUBMISSION_FAILED',
  'NEW_EVENT'
);

CREATE TYPE room_type AS ENUM (
  'Classroom',
  'Lab',
  'LectureHall',
  'Departmental',
  'Other'
);

CREATE TYPE exam_type AS ENUM (
  'FINAL_EXAM',
  'CAT'
);

CREATE TYPE online_exam_status AS ENUM (
  'PENDING_APPROVAL',
  'APPROVED',
  'REJECTED',
  'NOT_APPLICABLE'
);
