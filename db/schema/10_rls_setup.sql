-- eportalgo/db/schema/10_rls_setup.sql

-- 1. Helper functions for RLS
CREATE OR REPLACE FUNCTION get_current_school_id() RETURNS UUID AS $$
BEGIN
    -- Use current_setting with missing_ok = true to avoid errors if not set
    RETURN NULLIF(current_setting('app.current_school_id', TRUE), '')::UUID;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION is_platform_admin() RETURNS BOOLEAN AS $$
BEGIN
    RETURN current_setting('app.current_role', TRUE) = 'Executive Administrator';
EXCEPTION WHEN OTHERS THEN
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 2. Function to enable RLS and add policy to a table
-- This helps automate the process for many tables
CREATE OR REPLACE PROCEDURE enable_rls_for_tenant(table_name_text TEXT) AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', table_name_text);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation_policy ON %I', table_name_text);
    EXECUTE format(
        'CREATE POLICY tenant_isolation_policy ON %I USING (school_id = get_current_school_id() OR is_platform_admin())',
        table_name_text
    );
END;
$$ LANGUAGE plpgsql;

-- 3. Apply RLS to tables
DO $$
DECLARE
    t TEXT;
BEGIN
    -- List of tables that have a school_id column and need isolation
    FOR t IN SELECT table_name 
             FROM information_schema.columns 
             WHERE column_name = 'school_id' 
               AND table_schema = 'public'
               AND table_name NOT IN ('schools') -- schools table handled specially
    LOOP
        CALL enable_rls_for_tenant(t);
    END;
END $$;

-- 4. Special case for schools table
-- A school should only be able to see its own record, unless it's an admin
ALTER TABLE schools ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS school_self_isolation_policy ON schools;
CREATE POLICY school_self_isolation_policy ON schools
    USING (school_id = get_current_school_id() OR is_platform_admin());

-- 5. Special case for roles
-- Roles are currently global (is_school_role = FALSE) or school-specific (is_school_role = TRUE)
-- For now, roles remain accessible to all to avoid chicken-and-egg problems during login
-- but we could restrict school-specific roles later.
