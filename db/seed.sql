-- eportalgo/db/seed.sql

-- Clear existing data (Optional, use with caution)
-- DELETE FROM saas_company_profiles;
-- DELETE FROM users;
-- DELETE FROM roles;

-- 1. Seed Roles
INSERT INTO roles (role_name, description, is_school_role) VALUES
('Executive Administrator', 'Platform-level administrator with full access to all schools and platform settings.', FALSE),
('Developer', 'Platform-level developer with technical access.', FALSE),
('DB Manager', 'Platform-level database manager.', FALSE),
('Academic Administrator', 'School-level administrator focused on academic management.', TRUE),
('Finance Administrator', 'School-level administrator focused on billing and finance.', TRUE),
('IT Administrator', 'School-level administrator for technical support and settings.', TRUE),
('Teacher', 'School-level teaching staff.', TRUE),
('Student', 'School-level student.', TRUE),
('Parent', 'Parent or guardian of a student.', TRUE)
ON CONFLICT (role_name) DO UPDATE SET 
    description = EXCLUDED.description,
    is_school_role = EXCLUDED.is_school_role;

-- 2. Seed Executive Administrator User
-- Note: Replace firebase_uid and password_hash with real values when deploying
-- For now, we use placeholders.
WITH admin_role AS (
    SELECT role_id FROM roles WHERE role_name = 'Executive Administrator' LIMIT 1
)
INSERT INTO users (
    school_id, 
    role_id, 
    first_name, 
    last_name, 
    email, 
    firebase_uid, 
    password_hash, 
    is_active
) 
SELECT 
    NULL, 
    role_id, 
    'Platform', 
    'Owner', 
    'admin@eportal.com', 
    'executive-admin-placeholder-uid', 
    '$2a$10$YourBcryptHashHere', -- Replace with a real bcrypt hash
    TRUE
FROM admin_role
ON CONFLICT (email) DO NOTHING;

-- 3. Seed SaaS Company Profile for the Executive Administrator
WITH admin_user AS (
    SELECT user_id FROM users WHERE email = 'admin@eportal.com' LIMIT 1
)
INSERT INTO saas_company_profiles (
    user_id, 
    job_title, 
    department_name_text, 
    internal_employee_id, 
    hire_date
)
SELECT 
    user_id, 
    'Chief Executive Officer', 
    'Executive Management', 
    'EMP-001', 
    CURRENT_DATE
FROM admin_user
ON CONFLICT (user_id) DO NOTHING;

-- 4. Seed some default subscription tiers (optional but useful)
INSERT INTO subscription_tiers (tier_name, description, price_per_quarter, currency, features) VALUES
('Basic', 'Essential features for small schools.', 150.00, 'USD', '{"max_students": 500, "features": ["attendance", "grading", "basic_reporting"]}'),
('Professional', 'Advanced features for growing institutions.', 450.00, 'USD', '{"max_students": 2000, "features": ["attendance", "grading", "finance", "parent_portal"]}'),
('Enterprise', 'Full suite of features for large campuses.', 1200.00, 'USD', '{"max_students": -1, "features": ["all"]}')
ON CONFLICT (tier_name) DO UPDATE SET
    description = EXCLUDED.description,
    price_per_quarter = EXCLUDED.price_per_quarter,
    features = EXCLUDED.features;
