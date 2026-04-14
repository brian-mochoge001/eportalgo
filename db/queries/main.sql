-- eportalgo/db/queries/main.sql

-- Schools
-- name: GetSchool :one
SELECT * FROM schools
WHERE school_id = $1 LIMIT 1;

-- name: ListSchools :many
SELECT * FROM schools
ORDER BY school_name;

-- name: CreateSchool :one
INSERT INTO schools (
  school_name, subdomain, status, school_initial, address, phone_number, email, logo_url, primary_color, secondary_color
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- Users
-- name: GetUserByFirebaseUID :one
SELECT u.*, r.role_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
WHERE u.firebase_uid = $1 LIMIT 1;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 AND school_id = $2 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND school_id = $2 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  school_id, role_id, first_name, last_name, email, firebase_uid, password_hash, phone_number, date_of_birth, gender, is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE role_name = $1 LIMIT 1;

-- Profiles
-- name: CreateStudentProfile :one
INSERT INTO student_profiles (
  user_id, school_id, enrollment_number, current_grade_level, admission_date, current_class_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: CreateParentProfile :one
INSERT INTO parent_profiles (
  user_id, school_id, home_address, occupation, emergency_contact_name, emergency_contact_phone
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- Transfers
-- name: CreateTransferRequest :one
INSERT INTO transfer_requests (
  entity_type, entity_id, source_school_id, destination_school_id, initiated_by_user_id, notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetTransferRequestByID :one
SELECT * FROM transfer_requests
WHERE transfer_id = $1 LIMIT 1;

-- name: UpdateTransferRequestStatus :one
UPDATE transfer_requests
SET status = $2, completion_date = $3, notes = $4, updated_at = CURRENT_TIMESTAMP
WHERE transfer_id = $1
RETURNING *;

-- Audit Logs
-- name: CreateAuditLog :one
INSERT INTO audit_logs (
  school_id, user_id, action, entity_type, entity_id, old_value, new_value, ip_address, user_agent
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListAuditLogs :many
SELECT al.*, u.first_name, u.last_name, u.email, s.school_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.user_id
LEFT JOIN schools s ON al.school_id = s.school_id
WHERE (al.school_id = sqlc.narg('school_id') OR sqlc.arg('is_super_admin')::boolean = true)
  AND (sqlc.narg('user_id')::uuid IS NULL OR al.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('entity_type')::text IS NULL OR al.entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('entity_id')::uuid IS NULL OR al.entity_id = sqlc.narg('entity_id'))
  AND (sqlc.narg('query')::text IS NULL OR al.search_vector @@ websearch_to_tsquery('english', sqlc.narg('query')))
ORDER BY al.logged_at DESC;

-- Academic Structure
-- name: GetClassesBySchool :many
SELECT c.*, t.first_name as teacher_first_name, t.last_name as teacher_last_name, co.course_name,
       (SELECT COUNT(*) FROM enrollments e WHERE e.class_id = c.class_id) as enrollment_count
FROM academic_classes c
JOIN users t ON c.teacher_id = t.user_id
JOIN courses co ON c.course_id = co.course_id
WHERE c.school_id = $1;

-- name: CreateAcademicClass :one
INSERT INTO academic_classes (
  school_id, course_id, teacher_id, class_name, academic_year, semester, start_date, end_date
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetClassByID :one
SELECT * FROM academic_classes
WHERE class_id = $1 AND school_id = $2 LIMIT 1;

-- name: GetClassWithDetails :one
SELECT ac.*, t.user_id as teacher_user_id,
       hod.user_id as head_of_department_id
FROM academic_classes ac
LEFT JOIN users t ON ac.teacher_id = t.user_id
LEFT JOIN courses co ON ac.course_id = co.course_id
LEFT JOIN departments d ON co.department_id = d.department_id
LEFT JOIN users hod ON d.head_of_department_id = hod.user_id
WHERE ac.class_id = $1 AND ac.school_id = $2
LIMIT 1;

-- Subjects
-- name: GetSubjectsBySchool :many
SELECT * FROM subjects
WHERE school_id = $1;

-- name: CreateSubject :one
INSERT INTO subjects (
  school_id, subject_name, description, double_period_required, lab_period_required, max_online_percentage
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetSubjectByID :one
SELECT * FROM subjects
WHERE subject_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateSubject :one
UPDATE subjects
SET subject_name = $2, description = $3, double_period_required = $4, lab_period_required = $5, max_online_percentage = $6, updated_at = CURRENT_TIMESTAMP
WHERE subject_id = $1 AND school_id = $7
RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subjects
WHERE subject_id = $1 AND school_id = $2;

-- Courses
-- name: GetCoursesBySchool :many
SELECT * FROM courses
WHERE school_id = $1;

-- name: CreateCourse :one
INSERT INTO courses (
  school_id, course_code, course_name, description, is_short_course, price, is_graded_independently, requires_all_units_passed
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses
WHERE course_id = $1 AND school_id = $2 LIMIT 1;

-- Enrollments
-- name: GetEnrollmentsBySchool :many
SELECT * FROM enrollments
WHERE school_id = $1;

-- name: CreateEnrollment :one
INSERT INTO enrollments (
  school_id, student_id, class_id, enrollment_date, status
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetEnrollmentsByClass :many
SELECT student_id FROM enrollments
WHERE class_id = $1;

-- name: GetEnrollmentByStudentAndClass :one
SELECT * FROM enrollments
WHERE student_id = $1 AND class_id = $2 LIMIT 1;

-- name: UpdateEnrollmentStatus :exec
UPDATE enrollments
SET status = $3, updated_at = CURRENT_TIMESTAMP
WHERE student_id = $1 AND school_id = $2;

-- Short Courses
-- name: CheckShortCourseEnrollment :one
SELECT * FROM short_course_enrollments
WHERE student_id = $1 AND course_id = $2 LIMIT 1;

-- name: EnrollShortCourse :one
INSERT INTO short_course_enrollments (
  school_id, student_id, course_id, status
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UnenrollShortCourse :exec
DELETE FROM short_course_enrollments
WHERE student_id = $1 AND course_id = $2 AND school_id = $3;

-- Attendance
-- name: CreateAttendanceRecord :one
INSERT INTO attendance_records (
  school_id, student_id, class_id, attendance_date, status, notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetAttendanceByClass :many
SELECT a.*, u.first_name, u.last_name, u.email
FROM attendance_records a
JOIN users u ON a.student_id = u.user_id
WHERE a.class_id = $1 AND a.school_id = $2
ORDER BY a.attendance_date DESC;

-- name: GetAttendanceRecordByUnique :one
SELECT * FROM attendance_records
WHERE school_id = $1 AND student_id = $2 AND class_id = $3 AND attendance_date = $4
LIMIT 1;

-- name: UpdateAttendanceRecord :one
UPDATE attendance_records
SET status = $2, notes = $3, updated_at = CURRENT_TIMESTAMP
WHERE attendance_id = $1 AND school_id = $4
RETURNING *;

-- Assignments
-- name: CreateAssignment :one
INSERT INTO assignments (
  school_id, class_id, teacher_id, title, description, due_date, max_score, assignment_type, file_url
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetAssignmentByID :one
SELECT * FROM assignments
WHERE assignment_id = $1 AND school_id = $2 LIMIT 1;

-- name: GetAssignmentsByClass :many
SELECT * FROM assignments
WHERE class_id = $1 AND school_id = $2;

-- name: GetMyAssignments :many
SELECT a.*, ac.class_name
FROM assignments a
JOIN academic_classes ac ON a.class_id = ac.class_id
JOIN enrollments e ON ac.class_id = e.class_id
WHERE e.student_id = $1 AND a.school_id = $2
ORDER BY a.due_date ASC;

-- name: DeleteAssignment :exec
DELETE FROM assignments
WHERE assignment_id = $1 AND school_id = $2 AND teacher_id = $3;

-- Submissions
-- name: CreateSubmission :one
INSERT INTO submissions (
  school_id, student_id, assignment_id, submission_content, status
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- Quizzes
-- name: CreateQuiz :one
INSERT INTO quizzes (
  school_id, teacher_id, assignment_id, subject_id, title, description, quiz_type, duration_minutes, start_time, end_time
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetQuizByID :one
SELECT * FROM quizzes
WHERE quiz_id = $1 AND school_id = $2 LIMIT 1;

-- name: GetQuizzes :many
SELECT * FROM quizzes
WHERE school_id = $1;

-- Quiz Submissions
-- name: CreateQuizSubmission :one
INSERT INTO quiz_submissions (
  quiz_id, student_id, score, status
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetQuizSubmissions :many
SELECT qs.*, q.title as quiz_title, u.first_name, u.last_name, u.email
FROM quiz_submissions qs
JOIN quizzes q ON qs.quiz_id = q.quiz_id
JOIN users u ON qs.student_id = u.user_id
WHERE q.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('quiz_id')::uuid IS NULL OR qs.quiz_id = sqlc.narg('quiz_id'))
  AND (sqlc.narg('student_id')::uuid IS NULL OR qs.student_id = sqlc.narg('student_id'))
ORDER BY qs.submitted_at DESC;

-- name: CreateQuizAnswer :one
INSERT INTO quiz_answers (
  quiz_submission_id, question_id, student_answer_text, selected_option_id, is_correct
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- Notifications
-- name: CreateNotification :one
WITH new_notification AS (
  INSERT INTO notifications (
    school_id, sender_id, notification_type, title, message, link_url, sent_at
  ) VALUES (
    sqlc.arg('school_id'), sqlc.arg('sender_id'), sqlc.arg('notification_type'), sqlc.arg('title'), sqlc.arg('message'), sqlc.arg('link_url'), CURRENT_TIMESTAMP
  )
  RETURNING *
)
SELECT * FROM new_notification;

-- name: CreateNotificationRecipient :one
INSERT INTO notification_recipients (
  notification_id, recipient_id, is_read
) VALUES (
  $1, $2, false
)
RETURNING *;

-- name: GetNotificationsByRecipient :many
SELECT n.*, u.first_name as sender_first_name, u.last_name as sender_last_name, nr.is_read, nr.read_at
FROM notifications n
JOIN notification_recipients nr ON n.notification_id = nr.notification_id
LEFT JOIN users u ON n.sender_id = u.user_id
WHERE nr.recipient_id = $1
ORDER BY n.sent_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notification_recipients
SET is_read = true, read_at = CURRENT_TIMESTAMP
WHERE notification_id = $1 AND recipient_id = $2;

-- Badges
-- name: CreateBadge :one
INSERT INTO badges (
  school_id, badge_name, description, icon_url, criteria
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetBadgesBySchool :many
SELECT * FROM badges
WHERE school_id = $1
ORDER BY badge_name ASC;

-- Billing Contacts
-- name: CreateBillingContact :one
INSERT INTO billing_contacts (
  school_id, name, email, phone_number, role, is_primary
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ResetPrimaryBillingContact :exec
UPDATE billing_contacts
SET is_primary = false
WHERE school_id = sqlc.arg('school_id') AND is_primary = true AND (sqlc.narg('billing_contact_id')::uuid IS NULL OR billing_contact_id != sqlc.narg('billing_contact_id'));

-- Chat
-- name: CreateChatRoom :one
INSERT INTO chat_rooms (
  school_id, chat_name, chat_type, created_by_user_id, is_active
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: AddChatParticipant :exec
INSERT INTO chat_participants (
  school_id, chat_room_id, user_id
) VALUES (
  $1, $2, $3
);

-- name: CreateChatMessage :one
INSERT INTO chat_messages (
  chat_room_id, sender_id, school_id, message_text, attachment_url
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- Class Representatives
-- name: CreateClassRepresentative :one
INSERT INTO class_representatives (
  student_user_id, academic_class_id, can_communicate_teacher, can_communicate_department_head
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetClassRepresentativesBySchool :many
SELECT cr.*, u.first_name, u.last_name, u.email, ac.class_name
FROM class_representatives cr
JOIN users u ON cr.student_user_id = u.user_id
JOIN academic_classes ac ON cr.academic_class_id = ac.class_id
WHERE ac.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('academic_class_id')::uuid IS NULL OR cr.academic_class_id = sqlc.narg('academic_class_id'))
ORDER BY cr.created_at DESC;

-- Events
-- name: CreateEvent :one
INSERT INTO events (
  school_id, title, description, event_date, end_date, location, event_type, organizer_id, is_public
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetEventsBySchool :many
SELECT * FROM events
WHERE school_id = $1
ORDER BY event_date ASC;

-- Meetings
-- name: CreateMeeting :one
INSERT INTO meetings (
  school_id, title, agenda, meeting_date, duration_minutes, location, meeting_type, organizer_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetMeetingsBySchool :many
SELECT m.*, u.first_name as organizer_first_name, u.last_name as organizer_last_name
FROM meetings m
LEFT JOIN users u ON m.organizer_id = u.user_id
WHERE m.school_id = $1
ORDER BY m.meeting_date DESC;

-- Online Classes
-- name: CreateOnlineClassSession :one
INSERT INTO online_class_sessions (
  school_id, class_id, teacher_id, session_title, start_time, end_time, meeting_link, description, recording_link
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetOnlineClassSessions :many
SELECT ocs.*, ac.class_name, u.first_name as teacher_first_name, u.last_name as teacher_last_name
FROM online_class_sessions ocs
JOIN academic_classes ac ON ocs.class_id = ac.class_id
JOIN users u ON ocs.teacher_id = u.user_id
WHERE ocs.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('class_id')::uuid IS NULL OR ocs.class_id = sqlc.narg('class_id'))
  AND (sqlc.narg('teacher_id')::uuid IS NULL OR ocs.teacher_id = sqlc.narg('teacher_id'))
ORDER BY ocs.start_time DESC;

-- Learning Materials
-- name: CreateLearningMaterial :one
INSERT INTO learning_materials (
  school_id, uploaded_by_user_id, class_id, course_id, title, description, file_url, material_type
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetLearningMaterials :many
SELECT lm.*, u.first_name as uploader_first_name, u.last_name as uploader_last_name, ac.class_name, co.course_name
FROM learning_materials lm
JOIN users u ON lm.uploaded_by_user_id = u.user_id
LEFT JOIN academic_classes ac ON lm.class_id = ac.class_id
LEFT JOIN courses co ON lm.course_id = co.course_id
WHERE lm.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('class_id')::uuid IS NULL OR lm.class_id = sqlc.narg('class_id'))
  AND (sqlc.narg('course_id')::uuid IS NULL OR lm.course_id = sqlc.narg('course_id'))
ORDER BY lm.uploaded_at DESC;

-- Lesson Plans
-- name: CreateLessonPlan :one
INSERT INTO lesson_plans (
  school_id, teacher_id, class_id, title, content, date_covered
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetLessonPlans :many
SELECT lp.*, u.first_name as teacher_first_name, u.last_name as teacher_last_name, ac.class_name
FROM lesson_plans lp
JOIN users u ON lp.teacher_id = u.user_id
LEFT JOIN academic_classes ac ON lp.class_id = ac.class_id
WHERE lp.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('teacher_id')::uuid IS NULL OR lp.teacher_id = sqlc.narg('teacher_id'))
  AND (sqlc.narg('class_id')::uuid IS NULL OR lp.class_id = sqlc.narg('class_id'))
ORDER BY lp.created_at DESC;

-- Feedback
-- name: CreateFeedback :one
INSERT INTO feedback (
  school_id, user_id, subject, message, rating, feedback_type, status, submitted_at
) VALUES (
  $1, $2, $3, $4, $5, $6, 'New', CURRENT_TIMESTAMP
)
RETURNING *;

-- name: ListFeedbacks :many
SELECT f.*, u.first_name, u.last_name, u.email, s.school_name
FROM feedback f
JOIN users u ON f.user_id = u.user_id
LEFT JOIN schools s ON f.school_id = s.school_id
WHERE (f.school_id = sqlc.narg('school_id') OR sqlc.arg('is_super_admin')::boolean = true)
ORDER BY f.submitted_at DESC;

-- name: GetFeedbackByID :one
SELECT f.*, u.first_name, u.last_name, u.email, s.school_name
FROM feedback f
JOIN users u ON f.user_id = u.user_id
LEFT JOIN schools s ON f.school_id = s.school_id
WHERE f.feedback_id = $1 LIMIT 1;

-- Rooms & Departments
-- name: CreateRoom :one
INSERT INTO rooms (
  school_id, room_name, capacity, room_type, department_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRoomsBySchool :many
SELECT r.*, d.department_name
FROM rooms r
LEFT JOIN departments d ON r.department_id = d.department_id
WHERE r.school_id = $1;

-- name: GetDepartmentByName :one
SELECT * FROM departments
WHERE school_id = $1 AND department_name = $2 LIMIT 1;

-- name: CreateDepartment :one
INSERT INTO departments (
  school_id, department_name, head_of_department_id, deputy_head_of_department_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments
WHERE department_id = $1 AND school_id = $2;

-- Payments
-- name: CreatePayment :one
INSERT INTO payments (
  school_id, student_fee_id, amount, payment_method, transaction_id, recorded_by_user_id, notes, receipt_number
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListPayments :many
SELECT p.*, u.first_name, u.last_name, sf.amount_due, fs.fee_name
FROM payments p
JOIN student_fees sf ON p.student_fee_id = sf.student_fee_id
JOIN users u ON sf.student_id = u.user_id
JOIN fee_structures fs ON sf.fee_structure_id = fs.fee_structure_id
WHERE p.school_id = $1
ORDER BY p.payment_date DESC;

-- Teacher Workloads
-- name: GetTeacherWorkloads :many
SELECT tw.*, u.first_name, u.last_name, u.email
FROM teacher_workloads tw
JOIN users u ON tw.teacher_id = u.user_id
WHERE u.school_id = $1
  AND (sqlc.narg('teacher_id')::uuid IS NULL OR tw.teacher_id = sqlc.narg('teacher_id'))
ORDER BY tw.created_at DESC;

-- Grades
-- name: CreateGrade :one
INSERT INTO grades (
  school_id, submission_id, graded_by_user_id, score, feedback, graded_at
) VALUES (
  $1, $2, $3, $4, $5, CURRENT_TIMESTAMP
)
RETURNING *;

-- External Certifications
-- name: CreateExternalCertification :one
INSERT INTO external_certifications (
  student_id, name, issuer, credential_id, verification_url, issue_date, expiry_date, is_verified
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetExternalCertifications :many
SELECT ec.*, u.first_name, u.last_name, u.email
FROM external_certifications ec
JOIN users u ON ec.student_id = u.user_id
WHERE ($1::uuid IS NULL OR ec.student_id = $1)
ORDER BY ec.issue_date DESC;

-- =============================================
-- MISSING QUERIES (added to fix build errors)
-- =============================================

-- Assignments: Update
-- name: UpdateAssignment :one
UPDATE assignments
SET title = $3, description = $4, due_date = $5, max_score = $6, assignment_type = $7, file_url = $8, updated_at = CURRENT_TIMESTAMP
WHERE assignment_id = $1 AND school_id = $2
RETURNING *;

-- Attendance: GetStudentAttendance
-- name: GetStudentAttendance :many
SELECT a.*, ac.class_name
FROM attendance_records a
JOIN academic_classes ac ON a.class_id = ac.class_id
WHERE a.student_id = $1 AND a.school_id = $2
ORDER BY a.attendance_date DESC;

-- Audit Logs: GetAuditLog
-- name: GetAuditLog :one
SELECT al.*, u.first_name, u.last_name, u.email, s.school_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.user_id
LEFT JOIN schools s ON al.school_id = s.school_id
WHERE al.log_id = $1 LIMIT 1;

-- Badges: GetByID, Update, Delete, Award, Revoke
-- name: GetBadgeByID :one
SELECT * FROM badges
WHERE badge_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateBadge :one
UPDATE badges
SET badge_name = $3, description = $4, icon_url = $5, criteria = $6, updated_at = CURRENT_TIMESTAMP
WHERE badge_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteBadge :exec
DELETE FROM badges
WHERE badge_id = $1 AND school_id = $2;

-- name: AwardBadge :one
INSERT INTO student_badges (
  school_id, student_id, badge_id, awarded_by_user_id, notes
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: RevokeBadge :exec
DELETE FROM student_badges
WHERE badge_id = $1 AND student_id = $2 AND school_id = $3;

-- name: GetStudentBadges :many
SELECT sb.*, b.badge_name, b.description, b.icon_url, b.criteria
FROM student_badges sb
JOIN badges b ON sb.badge_id = b.badge_id
WHERE sb.student_id = $1 AND sb.school_id = $2
ORDER BY sb.awarded_at DESC;

-- Billing Contacts: GetBySchool, GetByID, Update, Delete
-- name: GetBillingContactsBySchool :many
SELECT * FROM billing_contacts
WHERE school_id = $1
ORDER BY is_primary DESC, name ASC;

-- name: GetBillingContactByID :one
SELECT * FROM billing_contacts
WHERE billing_contact_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateBillingContact :one
UPDATE billing_contacts
SET name = $3, email = $4, phone_number = $5, role = $6, is_primary = $7, updated_at = CURRENT_TIMESTAMP
WHERE billing_contact_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteBillingContact :exec
DELETE FROM billing_contacts
WHERE billing_contact_id = $1 AND school_id = $2;

-- Chat: GetRoomsByUser, GetMessages, GetParticipants, GetRoom
-- name: GetChatRoomsByUser :many
SELECT cr.*
FROM chat_rooms cr
JOIN chat_participants cp ON cr.chat_room_id = cp.chat_room_id
WHERE cp.user_id = $1 AND cr.is_active = true
ORDER BY cr.updated_at DESC;

-- name: GetChatMessagesByRoom :many
SELECT cm.*, u.first_name as sender_first_name, u.last_name as sender_last_name
FROM chat_messages cm
JOIN users u ON cm.sender_id = u.user_id
WHERE cm.chat_room_id = $1
ORDER BY cm.sent_at ASC;

-- name: GetChatParticipants :many
SELECT cp.*, u.first_name, u.last_name, u.email, u.profile_picture_url
FROM chat_participants cp
JOIN users u ON cp.user_id = u.user_id
WHERE cp.chat_room_id = $1 AND cp.status = 'active';

-- name: GetChatRoom :one
SELECT * FROM chat_rooms
WHERE chat_room_id = $1 LIMIT 1;

-- Class Representatives: GetByID, Update, Delete
-- name: GetClassRepresentativeByID :one
SELECT cr.*, u.first_name, u.last_name, u.email, ac.class_name, ac.school_id
FROM class_representatives cr
JOIN users u ON cr.student_user_id = u.user_id
JOIN academic_classes ac ON cr.academic_class_id = ac.class_id
WHERE cr.class_rep_id = $1 LIMIT 1;

-- name: UpdateClassRepresentative :one
UPDATE class_representatives
SET can_communicate_teacher = $2, can_communicate_department_head = $3, updated_at = CURRENT_TIMESTAMP
WHERE class_rep_id = $1
RETURNING *;

-- name: DeleteClassRepresentative :exec
DELETE FROM class_representatives
WHERE class_rep_id = $1;

-- name: DeactivateChatRoomsByParticipant :exec
UPDATE chat_rooms
SET is_active = false, updated_at = CURRENT_TIMESTAMP
WHERE chat_room_id IN (
  SELECT chat_room_id FROM chat_participants WHERE user_id = $1
);

-- Departments: GetBySchool, GetByID, Update, AddSubject, ClearSubjects
-- name: GetDepartmentsBySchool :many
SELECT d.*, u1.first_name as head_first_name, u1.last_name as head_last_name,
       u2.first_name as deputy_first_name, u2.last_name as deputy_last_name
FROM departments d
LEFT JOIN users u1 ON d.head_of_department_id = u1.user_id
LEFT JOIN users u2 ON d.deputy_head_of_department_id = u2.user_id
WHERE d.school_id = $1
ORDER BY d.department_name ASC;

-- name: GetDepartmentByID :one
SELECT * FROM departments
WHERE department_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateDepartment :one
UPDATE departments
SET department_name = $3, head_of_department_id = $4, deputy_head_of_department_id = $5, updated_at = CURRENT_TIMESTAMP
WHERE department_id = $1 AND school_id = $2
RETURNING *;

-- name: AddDepartmentSubject :exec
INSERT INTO department_subjects (department_id, subject_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ClearDepartmentSubjects :exec
DELETE FROM department_subjects
WHERE department_id = $1;

-- Events: GetByID, Update, Delete
-- name: GetEventByID :one
SELECT * FROM events
WHERE event_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateEvent :one
UPDATE events
SET title = $3, description = $4, event_date = $5, end_date = $6, location = $7, event_type = $8, organizer_id = $9, is_public = $10, updated_at = CURRENT_TIMESTAMP
WHERE event_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE event_id = $1 AND school_id = $2;

-- External Certifications: GetByID, Update, Delete
-- name: GetExternalCertificationByID :one
SELECT * FROM external_certifications
WHERE cert_id = $1 LIMIT 1;

-- name: UpdateExternalCertification :one
UPDATE external_certifications
SET name = $2, issuer = $3, credential_id = $4, verification_url = $5, issue_date = $6, expiry_date = $7, is_verified = $8, updated_at = CURRENT_TIMESTAMP
WHERE cert_id = $1
RETURNING *;

-- name: DeleteExternalCertification :exec
DELETE FROM external_certifications
WHERE cert_id = $1;

-- Feedback: Update, Delete
-- name: UpdateFeedback :one
UPDATE feedback
SET subject = $2, message = $3, rating = $4, feedback_type = $5, status = $6, updated_at = CURRENT_TIMESTAMP
WHERE feedback_id = $1
RETURNING *;

-- name: DeleteFeedback :exec
DELETE FROM feedback
WHERE feedback_id = $1;

-- Fee Structures
-- name: GetFeeStructuresBySchool :many
SELECT * FROM fee_structures
WHERE school_id = $1
ORDER BY academic_year DESC, fee_name ASC;

-- name: CreateFeeStructure :one
INSERT INTO fee_structures (
  school_id, fee_name, amount, currency, academic_year, description, is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- Banners
-- name: CreateBanner :one
INSERT INTO banners (
  school_id, title, image_url, target_url, is_active, "order"
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetActiveBanners :many
SELECT * FROM banners
WHERE (school_id = $1 OR school_id IS NULL) AND is_active = true
ORDER BY "order" ASC, created_at DESC;

-- name: ListAllBanners :many
SELECT * FROM banners
WHERE (school_id = $1 OR sqlc.arg('is_super_admin')::boolean = true)
ORDER BY created_at DESC;

-- name: UpdateBanner :one
UPDATE banners
SET title = $2, image_url = $3, target_url = $4, is_active = $5, "order" = $6, updated_at = CURRENT_TIMESTAMP
WHERE banner_id = $1
RETURNING *;

-- name: DeleteBanner :exec
DELETE FROM banners
WHERE banner_id = $1;

-- Student Fees
-- name: GetStudentFeesBySchool :many
SELECT sf.*, u.first_name, u.last_name, u.email, fs.fee_name, fs.amount as fee_amount
FROM student_fees sf
JOIN users u ON sf.student_id = u.user_id
JOIN fee_structures fs ON sf.fee_structure_id = fs.fee_structure_id
WHERE sf.school_id = $1
ORDER BY sf.created_at DESC;

-- name: GetStudentFeesByStudent :many
SELECT sf.*, fs.fee_name, fs.amount as fee_amount
FROM student_fees sf
JOIN fee_structures fs ON sf.fee_structure_id = fs.fee_structure_id
WHERE sf.student_id = $1 AND sf.school_id = $2
ORDER BY sf.created_at DESC;

-- name: CreateStudentFee :one
INSERT INTO student_fees (
  school_id, student_id, fee_structure_id, amount_due, due_date, notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateStudentFeeAmountPaid :one
UPDATE student_fees
SET amount_paid = amount_paid + $2::decimal, updated_at = CURRENT_TIMESTAMP
WHERE student_fee_id = $1
RETURNING *;

-- name: GetChildrenForParent :many
SELECT u.user_id, u.first_name, u.last_name, u.profile_picture_url, sp.current_grade_level, s.school_name
FROM parent_student_relationships psr
JOIN users u ON psr.student_user_id = u.user_id
JOIN student_profiles sp ON u.user_id = sp.user_id
JOIN schools s ON psr.school_id = s.school_id
WHERE psr.parent_user_id = $1 AND psr.school_id = $2;

-- Grades: GetBySubmission
-- name: GetGradesBySubmission :many
SELECT g.*, u.first_name as grader_first_name, u.last_name as grader_last_name
FROM grades g
LEFT JOIN users u ON g.graded_by_user_id = u.user_id
WHERE g.submission_id = $1 AND g.school_id = $2
ORDER BY g.graded_at DESC;

-- Groups: Create, AddMember
-- name: CreateGroup :one
INSERT INTO groups (
  school_id, name, description, created_by_user_id, is_teacher_created, chat_room_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: AddGroupMember :one
INSERT INTO group_members (
  group_id, user_id, status
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- Learning Materials: GetByID, Update, Delete
-- name: GetLearningMaterialByID :one
SELECT * FROM learning_materials
WHERE material_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateLearningMaterial :one
UPDATE learning_materials
SET title = $3, description = $4, file_url = $5, material_type = $6, class_id = $7, course_id = $8, updated_at = CURRENT_TIMESTAMP
WHERE material_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteLearningMaterial :exec
DELETE FROM learning_materials
WHERE material_id = $1 AND school_id = $2;

-- Lesson Plans: GetByID, Update, Delete
-- name: GetLessonPlanByID :one
SELECT * FROM lesson_plans
WHERE lesson_plan_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateLessonPlan :one
UPDATE lesson_plans
SET teacher_id = $3, title = $4, content = $5, class_id = $6, date_covered = $7, updated_at = CURRENT_TIMESTAMP
WHERE lesson_plan_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteLessonPlan :exec
DELETE FROM lesson_plans
WHERE lesson_plan_id = $1 AND school_id = $2 AND teacher_id = $3;

-- Meetings: GetByID, Update, Delete, Attendees
-- name: GetMeetingByID :one
SELECT m.*, u.first_name as organizer_first_name, u.last_name as organizer_last_name
FROM meetings m
LEFT JOIN users u ON m.organizer_id = u.user_id
WHERE m.meeting_id = $1 AND m.school_id = $2 LIMIT 1;

-- name: UpdateMeeting :one
UPDATE meetings
SET title = $3, agenda = $4, meeting_date = $5, duration_minutes = $6, location = $7, meeting_type = $8, organizer_id = $9, updated_at = CURRENT_TIMESTAMP
WHERE meeting_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteMeeting :exec
DELETE FROM meetings
WHERE meeting_id = $1 AND school_id = $2;

-- name: AddMeetingAttendee :exec
INSERT INTO meeting_attendees (meeting_id, user_id, school_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetMeetingAttendees :many
SELECT ma.*, u.first_name, u.last_name, u.email
FROM meeting_attendees ma
JOIN users u ON ma.user_id = u.user_id
WHERE ma.meeting_id = $1;

-- name: RemoveMeetingAttendee :exec
DELETE FROM meeting_attendees
WHERE meeting_id = $1 AND user_id = $2 AND school_id = $3;

-- Newsletters: Create, Get, GetByID, Update, Delete
-- name: CreateNewsletter :one
INSERT INTO newsletters (
  title, content, sent_by_user_id, target_schools
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetNewsletters :many
SELECT n.*, u.first_name as sender_first_name, u.last_name as sender_last_name
FROM newsletters n
LEFT JOIN users u ON n.sent_by_user_id = u.user_id
WHERE (n.target_schools @> $1::jsonb OR n.target_schools = '[]'::jsonb OR n.target_schools IS NULL)
ORDER BY n.created_at DESC;

-- name: GetNewsletterByID :one
SELECT * FROM newsletters
WHERE newsletter_id = $1 LIMIT 1;

-- name: UpdateNewsletter :one
UPDATE newsletters
SET title = $2, content = $3, target_schools = $4, updated_at = CURRENT_TIMESTAMP
WHERE newsletter_id = $1
RETURNING *;

-- name: DeleteNewsletter :exec
DELETE FROM newsletters
WHERE newsletter_id = $1;

-- Online Class Sessions: GetByID, Update, Delete
-- name: GetOnlineClassSessionByID :one
SELECT * FROM online_class_sessions
WHERE session_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateOnlineClassSession :one
UPDATE online_class_sessions
SET teacher_id = $3, session_title = $4, start_time = $5, end_time = $6, meeting_link = $7, description = $8, recording_link = $9, updated_at = CURRENT_TIMESTAMP
WHERE session_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteOnlineClassSession :exec
DELETE FROM online_class_sessions
WHERE session_id = $1 AND school_id = $2 AND teacher_id = $3;

-- Parents
-- name: GetParentsBySchool :many
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, pp.*
FROM users u
JOIN parent_profiles pp ON u.user_id = pp.user_id
WHERE pp.school_id = $1
ORDER BY u.last_name, u.first_name;

-- name: GetParentByUserID :one
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, pp.*
FROM users u
JOIN parent_profiles pp ON u.user_id = pp.user_id
WHERE u.user_id = $1 AND pp.school_id = $2 LIMIT 1;

-- Questions & Options
-- name: CreateQuestion :one
INSERT INTO questions (
  quiz_id, question_text, question_type, "order"
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: CreateOption :one
INSERT INTO options (
  question_id, option_text, is_correct
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- Quiz Submissions: GetByID
-- name: GetQuizSubmissionByID :one
SELECT qs.*, q.title as quiz_title, q.teacher_id,
       u.first_name as student_first_name, u.last_name as student_last_name
FROM quiz_submissions qs
JOIN quizzes q ON qs.quiz_id = q.quiz_id
JOIN users u ON qs.student_id = u.user_id
WHERE qs.submission_id = $1 LIMIT 1;

-- Rooms: GetByID, Update, Delete
-- name: GetRoomByID :one
SELECT * FROM rooms
WHERE room_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateRoom :one
UPDATE rooms
SET room_name = $3, capacity = $4, room_type = $5, department_id = $6, updated_at = CURRENT_TIMESTAMP
WHERE room_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE room_id = $1 AND school_id = $2;

-- School Settings
-- name: GetSchoolSettings :one
SELECT * FROM school_settings
WHERE school_id = $1 LIMIT 1;

-- name: UpdateSchoolSettings :one
UPDATE school_settings
SET branding_logo_url = $2, branding_colors = $3, timezone = $4, preferences = $5, email_template_configs = $6, payment_providers = $7, updated_at = CURRENT_TIMESTAMP
WHERE school_id = $1
RETURNING *;

-- Short Course Grades
-- name: GradeShortCourse :one
INSERT INTO short_course_grades (
  enrollment_id, course_id, student_id, score, feedback, graded_by_user_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetShortCourseGrades :many
SELECT scg.*, u.first_name, u.last_name, u.email, c.course_name, e.school_id
FROM short_course_grades scg
JOIN users u ON scg.student_id = u.user_id
JOIN courses c ON scg.course_id = c.course_id
JOIN short_course_enrollments e ON scg.enrollment_id = e.enrollment_id
WHERE e.school_id = $1
  AND (sqlc.narg('course_id')::uuid IS NULL OR scg.course_id = sqlc.narg('course_id'))
  AND (sqlc.narg('student_id')::uuid IS NULL OR scg.student_id = sqlc.narg('student_id'))
ORDER BY scg.graded_at DESC;

-- name: GetShortCourseGradeByID :one
SELECT scg.*, u.first_name, u.last_name, u.email, c.course_name, e.school_id
FROM short_course_grades scg
JOIN users u ON scg.student_id = u.user_id
JOIN courses c ON scg.course_id = c.course_id
JOIN short_course_enrollments e ON scg.enrollment_id = e.enrollment_id
WHERE scg.grade_id = $1 LIMIT 1;

-- name: GetEnrollmentByID :one
SELECT sce.*, c.course_name
FROM short_course_enrollments sce
JOIN courses c ON sce.course_id = c.course_id
WHERE sce.enrollment_id = $1 LIMIT 1;

-- Student Course Progress
-- name: CreateStudentCourseProgress :one
INSERT INTO student_course_progress (
  enrollment_id, progress_percentage
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetStudentCourseProgresses :many
SELECT scp.*, e.student_id, e.school_id, e.class_id, u.first_name, u.last_name
FROM student_course_progress scp
JOIN enrollments e ON scp.enrollment_id = e.enrollment_id
JOIN users u ON e.student_id = u.user_id
WHERE e.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('enrollment_id')::uuid IS NULL OR scp.enrollment_id = sqlc.narg('enrollment_id'))
  AND (sqlc.narg('student_id')::uuid IS NULL OR e.student_id = sqlc.narg('student_id'))
ORDER BY scp.updated_at DESC;

-- name: GetStudentCourseProgressByID :one
SELECT scp.*, e.student_id, e.school_id, e.class_id
FROM student_course_progress scp
JOIN enrollments e ON scp.enrollment_id = e.enrollment_id
WHERE scp.progress_id = $1 LIMIT 1;

-- name: UpdateStudentCourseProgress :one
UPDATE student_course_progress
SET progress_percentage = $2, last_activity_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE progress_id = $1
RETURNING *;

-- name: DeleteStudentCourseProgress :exec
DELETE FROM student_course_progress
WHERE progress_id = $1;

-- Students
-- name: GetStudentsBySchool :many
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, u.is_active, sp.*
FROM users u
JOIN student_profiles sp ON u.user_id = sp.user_id
WHERE sp.school_id = $1
ORDER BY u.last_name, u.first_name;

-- name: GetStudentByUserID :one
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, u.is_active, sp.*
FROM users u
JOIN student_profiles sp ON u.user_id = sp.user_id
WHERE u.user_id = $1 AND sp.school_id = $2 LIMIT 1;

-- Submissions: Get, GetByID, UpdateStatus, GetByStudentAndAssignment
-- name: GetSubmissions :many
SELECT s.*, u.first_name as student_first_name, u.last_name as student_last_name, a.title as assignment_title
FROM submissions s
JOIN users u ON s.student_id = u.user_id
JOIN assignments a ON s.assignment_id = a.assignment_id
WHERE s.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('assignment_id')::uuid IS NULL OR s.assignment_id = sqlc.narg('assignment_id'))
  AND (sqlc.narg('student_id')::uuid IS NULL OR s.student_id = sqlc.narg('student_id'))
ORDER BY s.submitted_at DESC;

-- name: GetSubmissionByID :one
SELECT s.*, a.teacher_id, a.title as assignment_title
FROM submissions s
JOIN assignments a ON s.assignment_id = a.assignment_id
WHERE s.submission_id = $1 AND s.school_id = $2 LIMIT 1;

-- name: UpdateSubmissionStatus :one
UPDATE submissions
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE submission_id = $1 AND school_id = $3
RETURNING *;

-- name: GetSubmissionByStudentAndAssignment :one
SELECT * FROM submissions
WHERE student_id = $1 AND assignment_id = $2 LIMIT 1;

-- Subscription Tiers
-- name: GetSubscriptionTiers :many
SELECT * FROM subscription_tiers
ORDER BY price_per_quarter ASC;

-- name: GetSubscriptionTierByID :one
SELECT * FROM subscription_tiers
WHERE tier_id = $1 LIMIT 1;

-- Teacher Availability
-- name: CreateTeacherAvailability :one
INSERT INTO teacher_availability (
  teacher_id, day_of_week, start_time, end_time, is_recurring, notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetTeacherAvailabilities :many
SELECT * FROM teacher_availability
WHERE (sqlc.narg('teacher_id')::uuid IS NULL OR teacher_id = sqlc.narg('teacher_id'))
ORDER BY day_of_week, start_time;

-- name: GetTeacherAvailabilityByID :one
SELECT * FROM teacher_availability
WHERE availability_id = $1 LIMIT 1;

-- name: UpdateTeacherAvailability :one
UPDATE teacher_availability
SET teacher_id = $2, day_of_week = $3, start_time = $4, end_time = $5, is_recurring = $6, notes = $7, updated_at = CURRENT_TIMESTAMP
WHERE availability_id = $1
RETURNING *;

-- name: DeleteTeacherAvailability :exec
DELETE FROM teacher_availability
WHERE availability_id = $1 AND teacher_id = $2;

-- Teacher Workloads: Create, GetByID, Update, Delete
-- name: CreateTeacherWorkload :one
INSERT INTO teacher_workloads (
  teacher_id, max_hours_per_week, current_hours_per_week
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetTeacherWorkloadByID :one
SELECT * FROM teacher_workloads
WHERE workload_id = $1 LIMIT 1;

-- name: UpdateTeacherWorkload :one
UPDATE teacher_workloads
SET max_hours_per_week = $2, current_hours_per_week = $3, updated_at = CURRENT_TIMESTAMP
WHERE workload_id = $1
RETURNING *;

-- name: DeleteTeacherWorkload :exec
DELETE FROM teacher_workloads
WHERE workload_id = $1;

-- Teachers
-- name: GetTeachersBySchool :many
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, u.is_active, tp.*
FROM users u
JOIN teacher_profiles tp ON u.user_id = tp.user_id
WHERE tp.school_id = $1
ORDER BY u.last_name, u.first_name;

-- name: GetTeacherByUserID :one
SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone_number, u.profile_picture_url, u.is_active, tp.*
FROM users u
JOIN teacher_profiles tp ON u.user_id = tp.user_id
WHERE u.user_id = $1 AND tp.school_id = $2 LIMIT 1;

-- Timetables
-- name: GetTimetables :many
SELECT * FROM timetables
WHERE school_id = $1
ORDER BY created_at DESC;

-- name: GetTimetableByID :one
SELECT * FROM timetables
WHERE timetable_id = $1 AND school_id = $2 LIMIT 1;

-- name: CreateTimetable :one
INSERT INTO timetables (
  school_id, academic_year, semester, title, description, is_active
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- Scheduling Engine Queries
-- name: GetCourseSubjects :many
SELECT s.*
FROM subjects s
JOIN course_subjects cs ON s.subject_id = cs.subject_id
WHERE cs.course_id = $1;

-- name: GetTeacherSubjects :many
SELECT s.*, ts.teacher_id
FROM subjects s
JOIN teacher_subjects ts ON s.subject_id = ts.subject_id
WHERE ts.teacher_id = $1;

-- name: ListTeacherSubjectsBySchool :many
SELECT ts.teacher_id, s.*
FROM teacher_subjects ts
JOIN subjects s ON ts.subject_id = s.subject_id
WHERE s.school_id = $1;

-- name: GetClassesForScheduling :many
SELECT ac.*, c.course_name,
       (SELECT COUNT(*) FROM enrollments e WHERE e.class_id = ac.class_id) as enrollment_count
FROM academic_classes ac
JOIN courses c ON ac.course_id = c.course_id
WHERE ac.school_id = $1 AND ac.academic_year = $2 AND (ac.semester = $3 OR ac.semester IS NULL);
-- name: GetTimetableEntries :many
SELECT te.*, ac.class_name, s.subject_name, u.first_name as teacher_first_name, u.last_name as teacher_last_name, r.room_name
FROM timetable_entries te
JOIN academic_classes ac ON te.class_id = ac.class_id
JOIN subjects s ON te.subject_id = s.subject_id
JOIN users u ON te.teacher_id = u.user_id
JOIN rooms r ON te.room_id = r.room_id
WHERE te.timetable_id = $1
ORDER BY te.day_of_week, te.start_time;

-- name: CreateTimetableEntry :one
INSERT INTO timetable_entries (
  timetable_id, class_id, subject_id, teacher_id, room_id, day_of_week, start_time, end_time
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: DeleteTimetableEntriesByTimetable :exec
DELETE FROM timetable_entries
WHERE timetable_id = $1;

-- Transcripts
-- name: CreateTranscript :one
INSERT INTO transcripts (
  school_id, student_id, academic_year, cumulative_gpa, transcript_data, issued_by_user_id
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetTranscripts :many
SELECT t.*, u.first_name as student_first_name, u.last_name as student_last_name
FROM transcripts t
JOIN users u ON t.student_id = u.user_id
WHERE t.school_id = sqlc.arg('school_id')
  AND (sqlc.narg('student_id')::uuid IS NULL OR t.student_id = sqlc.narg('student_id'))
ORDER BY t.issued_at DESC;

-- name: GetTranscriptByID :one
SELECT * FROM transcripts
WHERE transcript_id = $1 AND school_id = $2 LIMIT 1;

-- name: UpdateTranscript :one
UPDATE transcripts
SET academic_year = $3, cumulative_gpa = $4, transcript_data = $5, updated_at = CURRENT_TIMESTAMP
WHERE transcript_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteTranscript :exec
DELETE FROM transcripts
WHERE transcript_id = $1 AND school_id = $2;

-- Transfer Requests: List, Delete
-- name: ListTransferRequests :many
SELECT tr.*, u.first_name as entity_first_name, u.last_name as entity_last_name,
       ss.school_name as source_school_name, ds.school_name as destination_school_name,
       iu.first_name as initiated_first_name, iu.last_name as initiated_last_name
FROM transfer_requests tr
JOIN users u ON tr.entity_id = u.user_id
JOIN schools ss ON tr.source_school_id = ss.school_id
JOIN schools ds ON tr.destination_school_id = ds.school_id
JOIN users iu ON tr.initiated_by_user_id = iu.user_id
WHERE (sqlc.narg('status')::text IS NULL OR tr.status = sqlc.narg('status'))
  AND (sqlc.narg('entity_type')::text IS NULL OR tr.entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('source_school_id')::uuid IS NULL OR tr.source_school_id = sqlc.narg('source_school_id'))
  AND (sqlc.narg('destination_school_id')::uuid IS NULL OR tr.destination_school_id = sqlc.narg('destination_school_id'))
  AND (sqlc.narg('school_id')::uuid IS NULL OR tr.source_school_id = sqlc.narg('school_id') OR tr.destination_school_id = sqlc.narg('school_id'))
ORDER BY tr.request_date DESC;

-- name: DeleteTransferRequest :exec
DELETE FROM transfer_requests
WHERE transfer_id = $1 AND (source_school_id = $2 OR destination_school_id = $2);

-- Users: ListBySchool, GetStudentProfileByUserID, GetParentProfileByUserID
-- name: ListUsersBySchool :many
SELECT u.*, r.role_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
WHERE u.school_id = $1
  AND (sqlc.narg('query')::text IS NULL OR u.search_vector @@ websearch_to_tsquery('english', sqlc.narg('query')))
ORDER BY u.last_name, u.first_name;

-- name: GetStudentProfileByUserID :one
SELECT * FROM student_profiles
WHERE user_id = $1 AND school_id = $2 LIMIT 1;

-- name: GetParentProfileByUserID :one
SELECT * FROM parent_profiles
WHERE user_id = $1 AND school_id = $2 LIMIT 1;

-- =============================================
-- MISSING QUERIES (added to fix build errors)
-- =============================================

-- Quizzes: Update, Delete
-- name: UpdateQuiz :one
UPDATE quizzes
SET title = $3, description = $4, quiz_type = $5, duration_minutes = $6, start_time = $7, end_time = $8, updated_at = CURRENT_TIMESTAMP
WHERE quiz_id = $1 AND school_id = $2
RETURNING *;

-- name: DeleteQuiz :exec
DELETE FROM quizzes
WHERE quiz_id = $1 AND school_id = $2;

-- Early Warning System
-- name: CalculateStudentMetrics :many
SELECT
    sp.user_id,
    sp.school_id,
    COALESCE(
        (SELECT (COUNT(CASE WHEN ar.status = 'Present' THEN 1 END)::DECIMAL / NULLIF(COUNT(*), 0)) * 100
         FROM attendance_records ar
         WHERE ar.student_id = sp.user_id),
        100.00
    )::DECIMAL(5,2) as attendance_rate,
    COALESCE(
        (SELECT AVG(g.score)
         FROM grades g
         JOIN submissions s ON g.submission_id = s.submission_id
         WHERE s.student_id = sp.user_id),
        0.00
    )::DECIMAL(5,2) as average_grade
FROM student_profiles sp
WHERE sp.school_id = $1;

-- name: UpsertStudentRiskScore :one
INSERT INTO student_risk_scores (
    school_id, student_id, attendance_rate, average_grade, risk_score, risk_level, last_calculated
) VALUES (
    $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP
)
ON CONFLICT (student_id) DO UPDATE SET
    attendance_rate = EXCLUDED.attendance_rate,
    average_grade = EXCLUDED.average_grade,
    risk_score = EXCLUDED.risk_score,
    risk_level = EXCLUDED.risk_level,
    last_calculated = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListAtRiskStudents :many
SELECT
    srs.*,
    u.first_name,
    u.last_name,
    u.email,
    sp.enrollment_number,
    sp.current_grade_level
FROM student_risk_scores srs
JOIN users u ON srs.student_id = u.user_id
JOIN student_profiles sp ON srs.student_id = sp.user_id
WHERE srs.school_id = $1
  AND (sqlc.narg('risk_level')::risk_level IS NULL OR srs.risk_level = sqlc.narg('risk_level'))
  AND srs.risk_level != 'Low'
ORDER BY srs.risk_score DESC;
-- name: GetSchoolByNameOrSubdomain :one
SELECT * FROM schools
WHERE school_name = sqlc.arg('school_name') OR subdomain = sqlc.arg('subdomain')
LIMIT 1;

-- name: GetSchoolByInitial :one
SELECT * FROM schools
WHERE school_initial = $1 LIMIT 1;

-- name: GetSchoolWithAdmin :one
SELECT s.*, u.firebase_uid as admin_firebase_uid, r.role_name as admin_role_name
FROM schools s
JOIN users u ON s.school_id = u.school_id
JOIN roles r ON u.role_id = r.role_id
WHERE s.school_id = $1
LIMIT 1;

-- name: UpdateSchoolStatus :one
UPDATE schools
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE school_id = $1
RETURNING *;

-- name: CreateSchoolSetting :one
INSERT INTO school_settings (school_id)
VALUES ($1)
RETURNING *;

-- =============================================
-- PARENT MONITORING & AUTH QUERIES
-- =============================================

-- Auth: Get user by email without school_id constraint (for JWT lookup)
-- name: GetUserByEmailOnly :one
SELECT u.*, r.role_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
WHERE u.email = $1 LIMIT 1;

-- Auth: Get user by ID (for JWT sub claim lookup)
-- name: GetUserByID :one
SELECT u.*, r.role_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
WHERE u.user_id = $1 LIMIT 1;

-- Parent-Child: Validate relationship
-- name: ValidateParentChildRelationship :one
SELECT * FROM parent_student_relationships
WHERE parent_user_id = $1 AND student_user_id = $2 AND school_id = $3
LIMIT 1;

-- Parent-Child: Create relationship
-- name: CreateParentStudentRelationship :one
INSERT INTO parent_student_relationships (
  parent_user_id, student_user_id, school_id, relationship_type
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- Parent Monitoring: Get child attendance records
-- name: GetChildAttendance :many
SELECT a.*, ac.class_name
FROM attendance_records a
JOIN academic_classes ac ON a.class_id = ac.class_id
WHERE a.student_id = $1 AND a.school_id = $2
ORDER BY a.attendance_date DESC;

-- Parent Monitoring: Get child grades
-- name: GetChildGrades :many
SELECT g.*, s.submission_id, a.title as assignment_title, a.max_score,
       u.first_name as grader_first_name, u.last_name as grader_last_name
FROM grades g
JOIN submissions s ON g.submission_id = s.submission_id
JOIN assignments a ON s.assignment_id = a.assignment_id
LEFT JOIN users u ON g.graded_by_user_id = u.user_id
WHERE s.student_id = $1 AND g.school_id = $2
ORDER BY g.graded_at DESC;

-- Parent Monitoring: Get child assignments
-- name: GetChildAssignments :many
SELECT a.*, ac.class_name,
       (SELECT s.status FROM submissions s WHERE s.student_id = $1 AND s.assignment_id = a.assignment_id LIMIT 1) as submission_status
FROM assignments a
JOIN academic_classes ac ON a.class_id = ac.class_id
JOIN enrollments e ON ac.class_id = e.class_id
WHERE e.student_id = $1 AND a.school_id = $2
ORDER BY a.due_date DESC;

-- Parent Monitoring: Get child fees
-- name: GetChildFees :many
SELECT sf.*, fs.fee_name, fs.amount as fee_amount,
       COALESCE(sf.amount_paid, 0) as total_paid,
       (sf.amount_due - COALESCE(sf.amount_paid, 0)) as balance
FROM student_fees sf
JOIN fee_structures fs ON sf.fee_structure_id = fs.fee_structure_id
WHERE sf.student_id = $1 AND sf.school_id = $2
ORDER BY sf.due_date DESC;

-- Reminders
-- name: ListReminderLists :many
SELECT * FROM reminder_lists
WHERE user_id = $1 AND (school_id = $2 OR school_id IS NULL)
ORDER BY title;

-- name: CreateReminderList :one
INSERT INTO reminder_lists (
  school_id, user_id, title, color
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: ListRemindersByList :many
SELECT * FROM reminders
WHERE list_id = $1 AND user_id = $2
ORDER BY due_date ASC, created_at DESC;

-- name: CreateReminder :one
INSERT INTO reminders (
  list_id, user_id, title, notes, due_date, priority
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateReminderStatus :one
UPDATE reminders
SET is_completed = $2, updated_at = CURRENT_TIMESTAMP
WHERE reminder_id = $1 AND user_id = $3
RETURNING *;

-- name: DeleteReminder :exec
DELETE FROM reminders
WHERE reminder_id = $1 AND user_id = $2;

-- Subject Specific Data
-- name: GetAssignmentsBySubject :many
SELECT a.*, ac.class_name
FROM assignments a
JOIN academic_classes ac ON a.class_id = ac.class_id
WHERE a.subject_id = $1 AND a.school_id = $2
ORDER BY a.due_date DESC;

-- name: GetMaterialsBySubject :many
SELECT lm.*, u.first_name as uploader_first_name, u.last_name as uploader_last_name
FROM learning_materials lm
JOIN users u ON lm.uploaded_by_user_id = u.user_id
WHERE lm.subject_id = $1 AND lm.school_id = $2
ORDER BY lm.uploaded_at DESC;

-- name: GetNotificationsBySubject :many
SELECT n.*, u.first_name as sender_first_name, u.last_name as sender_last_name, nr.is_read, nr.read_at
FROM notifications n
JOIN notification_recipients nr ON n.notification_id = nr.notification_id
LEFT JOIN users u ON n.sender_id = u.user_id
WHERE n.subject_id = $1 AND nr.recipient_id = $2
ORDER BY n.sent_at DESC;

-- Full Profile
-- name: GetStudentFullProfile :one
SELECT u.*, r.role_name, sp.enrollment_number, sp.current_grade_level, sp.admission_date, ac.class_name as current_class_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
JOIN student_profiles sp ON u.user_id = sp.user_id
LEFT JOIN academic_classes ac ON sp.current_class_id = ac.class_id
WHERE u.user_id = $1 LIMIT 1;

-- name: GetParentFullProfile :one
SELECT u.*, r.role_name, pp.home_address, pp.occupation, pp.emergency_contact_name, pp.emergency_contact_phone
FROM users u
JOIN roles r ON u.role_id = r.role_id
JOIN parent_profiles pp ON u.user_id = pp.user_id
WHERE u.user_id = $1 LIMIT 1;

-- Academic History
-- name: GetDetailedGrades :many
SELECT g.*, a.title as assignment_title, a.max_score, ac.class_name, s.subject_name
FROM grades g
JOIN submissions sub ON g.submission_id = sub.submission_id
JOIN assignments a ON sub.assignment_id = a.assignment_id
JOIN academic_classes ac ON a.class_id = ac.class_id
JOIN subjects s ON a.subject_id = s.subject_id
WHERE sub.student_id = $1 AND g.school_id = $2
ORDER BY ac.academic_year DESC, ac.semester DESC, g.graded_at DESC;
