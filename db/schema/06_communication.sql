-- eportalgo/db/schema/06_communication.sql

CREATE TABLE chat_rooms (
  chat_room_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  chat_name           TEXT NOT NULL,
  chat_type           TEXT NOT NULL,
  description         TEXT,
  is_active           BOOLEAN NOT NULL DEFAULT TRUE,
  created_by_user_id  UUID REFERENCES users(user_id) ON DELETE SET NULL,
  associated_class_id UUID REFERENCES academic_classes(class_id) ON DELETE SET NULL,
  allowed_file_types  TEXT[] DEFAULT '{}',
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_participants (
  participant_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id          UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  chat_room_id       UUID NOT NULL REFERENCES chat_rooms(chat_room_id) ON DELETE CASCADE,
  user_id            UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  joined_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at            TIMESTAMPTZ(6),
  status             TEXT NOT NULL DEFAULT 'active',
  invited_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(chat_room_id, user_id)
);

CREATE TABLE chat_messages (
  message_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id      UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  chat_room_id   UUID NOT NULL REFERENCES chat_rooms(chat_room_id) ON DELETE CASCADE,
  sender_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  message_text   TEXT NOT NULL, -- Should be encrypted at app level
  attachment_url TEXT,
  sent_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
  group_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id          UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  name               TEXT NOT NULL,
  description        TEXT,
  created_by_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  is_teacher_created BOOLEAN NOT NULL DEFAULT FALSE,
  max_members        INT,
  chat_room_id       UUID UNIQUE REFERENCES chat_rooms(chat_room_id) ON DELETE SET NULL
);

CREATE TABLE group_members (
  group_member_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id        UUID NOT NULL REFERENCES groups(group_id) ON DELETE CASCADE,
  user_id         UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  status          TEXT NOT NULL DEFAULT 'pending',
  UNIQUE(group_id, user_id)
);

CREATE TABLE notifications (
  notification_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id         UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  sender_id         UUID REFERENCES users(user_id) ON DELETE SET NULL,
  notification_type notification_type NOT NULL,
  title             TEXT NOT NULL,
  message           TEXT NOT NULL,
  link_url          TEXT,
  entity_type       TEXT,
  entity_id         TEXT,
  sent_at           TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_recipients (
  notification_id UUID NOT NULL REFERENCES notifications(notification_id) ON DELETE CASCADE,
  recipient_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  is_read         BOOLEAN NOT NULL DEFAULT FALSE,
  read_at         TIMESTAMPTZ(6),
  PRIMARY KEY (notification_id, recipient_id)
);

CREATE TABLE newsletters (
  newsletter_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title           TEXT NOT NULL,
  content         TEXT NOT NULL,
  sent_at         TIMESTAMPTZ(6) DEFAULT CURRENT_TIMESTAMP,
  sent_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  target_schools  JSONB DEFAULT '[]',
  attachments     JSONB,
  created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);

CREATE TABLE feedback (
  feedback_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID REFERENCES schools(school_id) ON DELETE CASCADE,
  user_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  subject       TEXT,
  message       TEXT NOT NULL,
  rating        INT,
  feedback_type TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'New',
  submitted_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at      TIMESTAMPTZ(6)
);
