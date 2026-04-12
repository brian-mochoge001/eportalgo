-- eportalgo/db/schema/12_full_text_search.sql

-- Add search vector to audit_logs
ALTER TABLE audit_logs ADD COLUMN search_vector tsvector;
CREATE INDEX idx_audit_logs_search_vector ON audit_logs USING GIN(search_vector);

CREATE OR REPLACE FUNCTION audit_logs_search_vector_update() RETURNS trigger AS $$
BEGIN
  new.search_vector := to_tsvector('english', coalesce(new.action, '') || ' ' || coalesce(new.entity_type, ''));
  RETURN new;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_logs_search_vector_update
BEFORE INSERT OR UPDATE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION audit_logs_search_vector_update();

-- Update existing rows
UPDATE audit_logs SET search_vector = to_tsvector('english', coalesce(action, '') || ' ' || coalesce(entity_type, ''));

-- Add search vector to users
ALTER TABLE users ADD COLUMN search_vector tsvector;
CREATE INDEX idx_users_search_vector ON users USING GIN(search_vector);

CREATE OR REPLACE FUNCTION users_search_vector_update() RETURNS trigger AS $$
BEGIN
  new.search_vector := to_tsvector('english', coalesce(new.first_name, '') || ' ' || coalesce(new.last_name, '') || ' ' || coalesce(new.email, ''));
  RETURN new;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_search_vector_update
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION users_search_vector_update();

-- Update existing rows
UPDATE users SET search_vector = to_tsvector('english', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(email, ''));
