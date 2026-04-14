-- eportalgo/db/schema/16_subject_filtering.sql

ALTER TABLE assignments ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects(subject_id) ON DELETE SET NULL;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects(subject_id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_assignments_subject_id ON assignments(subject_id);
CREATE INDEX IF NOT EXISTS idx_notifications_subject_id ON notifications(subject_id);
ALTER TABLE learning_materials ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects(subject_id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_learning_materials_subject_id ON learning_materials(subject_id);
