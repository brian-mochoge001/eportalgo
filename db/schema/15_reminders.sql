-- eportalgo/db/schema/15_reminders.sql

CREATE TABLE reminder_lists (
  list_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id  UUID REFERENCES schools(school_id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  title      TEXT NOT NULL,
  color      TEXT DEFAULT '#007AFF',
  created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reminders (
  reminder_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  list_id      UUID NOT NULL REFERENCES reminder_lists(list_id) ON DELETE CASCADE,
  user_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  title        TEXT NOT NULL,
  notes        TEXT,
  due_date     TIMESTAMPTZ(6),
  priority     TEXT DEFAULT 'medium', -- 'low', 'medium', 'high'
  is_completed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reminder_lists_user_id ON reminder_lists(user_id);
CREATE INDEX idx_reminders_list_id ON reminders(list_id);
CREATE INDEX idx_reminders_user_id ON reminders(user_id);
