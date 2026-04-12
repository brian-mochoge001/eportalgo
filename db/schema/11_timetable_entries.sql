-- eportalgo/db/schema/11_timetable_entries.sql

CREATE TABLE timetable_entries (
  entry_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  timetable_id  UUID NOT NULL REFERENCES timetables(timetable_id) ON DELETE CASCADE,
  class_id      UUID NOT NULL REFERENCES academic_classes(class_id) ON DELETE CASCADE,
  subject_id    UUID NOT NULL REFERENCES subjects(subject_id) ON DELETE CASCADE,
  teacher_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  room_id       UUID NOT NULL REFERENCES rooms(room_id) ON DELETE CASCADE,
  day_of_week   INT NOT NULL, -- 0 (Sunday) to 6 (Saturday)
  start_time    TIME NOT NULL,
  end_time      TIME NOT NULL,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_timetable_entries_timetable_id ON timetable_entries(timetable_id);
CREATE INDEX idx_timetable_entries_teacher_id ON timetable_entries(teacher_id);
CREATE INDEX idx_timetable_entries_class_id ON timetable_entries(class_id);
CREATE INDEX idx_timetable_entries_room_id ON timetable_entries(room_id);
