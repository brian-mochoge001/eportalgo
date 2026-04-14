package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/handlers"
	custom_mw "github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/brian-mochoge001/eportalgo/worker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	queries := db.New(conn)

	// JWKS URL for BetterAuth JWT verification
	jwksURL := os.Getenv("JWKS_URL")
	if jwksURL == "" {
		jwksURL = "http://localhost:3001/api/auth/jwks"
	}
	fmt.Printf("Using JWKS URL: %s\n", jwksURL)

	// Initialize Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	fmt.Printf("DEBUG: Using Redis URL: %s\n", redisURL)

	optRedis, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Error parsing Redis URL: %v", err)
	}
	fmt.Printf("DEBUG: Redis Addr: %s\n", optRedis.Addr)
	redisClient := redis.NewClient(optRedis)

	// Initialize Asynq client and server
	redisConnOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		log.Fatalf("Error parsing Redis URI for Asynq: %v", err)
	}
	
	// Type assert to see the address for debugging
	if clientOpt, ok := redisConnOpt.(asynq.RedisClientOpt); ok {
		fmt.Printf("DEBUG: Asynq Redis Addr: %s\n", clientOpt.Addr)
	} else if clusterOpt, ok := redisConnOpt.(asynq.RedisClusterClientOpt); ok {
		fmt.Printf("DEBUG: Asynq Redis Cluster Addrs: %v\n", clusterOpt.Addrs)
	}

	asynqClient := asynq.NewClient(redisConnOpt)
	defer asynqClient.Close()

	asynqServer := asynq.NewServer(
		redisConnOpt,
		asynq.Config{Concurrency: 10},
	)

	// Security headers (Helmet equivalent)
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		IsDevelopment:         os.Getenv("NODE_ENV") != "production",
	})

	schoolService := services.NewSchoolService(queries, redisClient)
	authService := services.NewAuthService(queries)
	studentService := services.NewStudentService(queries, conn)
	assignmentService := services.NewAssignmentService(queries, asynqClient)
	timetableService := services.NewTimetableService(queries, conn)
	userService := services.NewUserService(queries, conn)
	meetingService := services.NewMeetingService(queries, conn)
	courseService := services.NewCourseService(queries, conn)
	eventService := services.NewEventService(queries, conn)
	lessonPlanService := services.NewLessonPlanService(queries, conn)
	classService := services.NewClassService(queries, conn)
	attendanceService := services.NewAttendanceService(queries)
	financeService := services.NewFinanceService(queries, conn)
	quizService := services.NewQuizService(queries, conn)
	reportingService := services.NewReportingService(queries)

	schoolHandler := handlers.NewSchoolHandler(queries, schoolService, redisClient)
	authHandler := handlers.NewAuthHandler(queries, authService)
	userHandler := handlers.NewUserHandler(queries, userService)
	courseHandler := handlers.NewCourseHandler(queries, courseService)
	departmentHandler := handlers.NewDepartmentHandler(queries)
	classHandler := handlers.NewClassHandler(queries, classService)
	subjectHandler := handlers.NewSubjectHandler(queries)
	enrollmentHandler := handlers.NewEnrollmentHandler(queries, studentService)
	attendanceHandler := handlers.NewAttendanceHandler(queries, attendanceService)
	assignmentHandler := handlers.NewAssignmentHandler(queries, assignmentService)
	auditLogHandler := handlers.NewAuditLogHandler(queries)
	studentRiskHandler := handlers.NewStudentRiskHandler(queries, asynqClient)
	badgeHandler := handlers.NewBadgeHandler(queries)
	billingContactHandler := handlers.NewBillingContactHandler(queries, conn)
	chatHandler := handlers.NewChatHandler(queries)
	classRepresentativeHandler := handlers.NewClassRepresentativeHandler(queries, conn)
	eventHandler := handlers.NewEventHandler(queries, eventService)
	externalCertificationHandler := handlers.NewExternalCertificationHandler(queries)
	feeHandler := handlers.NewFeeHandler(queries)
	feedbackHandler := handlers.NewFeedbackHandler(queries)
	gradeHandler := handlers.NewGradeHandler(queries)
	groupHandler := handlers.NewGroupHandler(queries, conn)
	learningMaterialHandler := handlers.NewLearningMaterialHandler(queries)
	lessonPlanHandler := handlers.NewLessonPlanHandler(queries, lessonPlanService)
	meetingHandler := handlers.NewMeetingHandler(queries, meetingService)
	newsletterHandler := handlers.NewNewsletterHandler(queries)
	notificationHandler := handlers.NewNotificationHandler(queries)
	onlineClassSessionHandler := handlers.NewOnlineClassSessionHandler(queries)
	parentHandler := handlers.NewParentHandler(queries)
	paymentHandler := handlers.NewPaymentHandler(queries, financeService)
	quizHandler := handlers.NewQuizHandler(queries, conn)
	quizSubmissionHandler := handlers.NewQuizSubmissionHandler(queries, quizService)
	roomHandler := handlers.NewRoomHandler(queries)
	schoolSettingHandler := handlers.NewSchoolSettingHandler(queries)
	shortCourseGradeHandler := handlers.NewShortCourseGradeHandler(queries)
	studentHandler := handlers.NewStudentHandler(queries)
	studentCourseProgressHandler := handlers.NewStudentCourseProgressHandler(queries)
	submissionHandler := handlers.NewSubmissionHandler(queries, conn)
	subscriptionTierHandler := handlers.NewSubscriptionTierHandler(queries)
	teacherAvailabilityHandler := handlers.NewTeacherAvailabilityHandler(queries)
	teacherHandler := handlers.NewTeacherHandler(queries)
	teacherWorkloadHandler := handlers.NewTeacherWorkloadHandler(queries)
	timetableHandler := handlers.NewTimetableHandler(queries, timetableService)
	transcriptHandler := handlers.NewTranscriptHandler(queries, reportingService)
	transferRequestHandler := handlers.NewTransferRequestHandler(queries)
	bannerHandler := handlers.NewBannerHandler(queries)
	reminderHandler := handlers.NewReminderHandler(queries)

	// Set up router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(func(next http.Handler) http.Handler { return secureMiddleware.Handler(next) })
	r.Use(middleware.Compress(5))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID", "X-School-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler)

	r.Use(custom_mw.TenantMiddleware)
	r.Use(custom_mw.RequestLogger)
	r.Use(custom_mw.ErrorHandler)
	r.Use(middleware.Recoverer)

	// Asynq server
	taskHandler := worker.NewTaskHandler(queries)
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeAssignmentNotification, taskHandler.HandleAssignmentNotification)
	mux.HandleFunc(worker.TypeAuditLog, taskHandler.HandleAuditLog)
	mux.HandleFunc(worker.TypeCalculateRiskScores, taskHandler.HandleCalculateRiskScores)

	go func() {
		if err := asynqServer.Run(mux); err != nil {
			log.Fatalf("could not run asynq server: %v", err)
		}
	}()

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.RegisterUser)

			// Protected auth routes (require valid JWT)
			r.Group(func(r chi.Router) {
				r.Use(custom_mw.AuthMiddleware(jwksURL, queries))
				r.Get("/me", authHandler.GetMe)
				r.Post("/login", authHandler.LoginUser)
			})
		})

		r.Route("/schools", func(r chi.Router) {
			r.Post("/register", schoolHandler.RegisterSchool)
			r.Group(func(r chi.Router) {
				r.Use(custom_mw.AuthMiddleware(jwksURL, queries))
				r.Put("/{schoolId}/verify", schoolHandler.VerifySchool)
				r.Get("/{schoolId}/settings", schoolHandler.GetSchoolSettings)
				r.Put("/{schoolId}/settings", schoolHandler.UpdateSchoolSettings)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(custom_mw.AuthMiddleware(jwksURL, queries))
			r.Use(custom_mw.RLSMiddleware(conn))
			r.Use(custom_mw.AuditMiddleware(asynqClient))

			r.Route("/users", func(r chi.Router) {
				r.Post("/add", userHandler.AddUser)
				r.Get("/school/{schoolId}", userHandler.GetUsersBySchool)
				r.Post("/{userId}/student-profile", userHandler.AddStudentProfile)
				r.Post("/{userId}/parent-profile", userHandler.AddParentProfile)
				r.Get("/profile", userHandler.GetFullProfile)
				r.Get("/detailed-grades", userHandler.GetDetailedGrades)
			})

			r.Route("/reminders", func(r chi.Router) {
				r.Get("/lists", reminderHandler.ListReminderLists)
				r.Post("/lists", reminderHandler.CreateReminderList)
				r.Get("/{listId}", reminderHandler.ListReminders)
				r.Post("/", reminderHandler.CreateReminder)
				r.Put("/{id}/status", reminderHandler.UpdateReminderStatus)
				r.Delete("/{id}", reminderHandler.DeleteReminder)
			})

			r.Route("/courses", func(r chi.Router) {
				r.Get("/", courseHandler.GetCourses)
				r.Post("/", courseHandler.CreateCourse)
				r.Post("/{course_id}/enroll", courseHandler.EnrollShortCourse)
				r.Delete("/{course_id}/unenroll/{student_id}", courseHandler.UnenrollShortCourse)
			})

			r.Route("/departments", func(r chi.Router) {
				r.Get("/", departmentHandler.GetDepartments)
				r.Post("/", departmentHandler.CreateDepartment)
				r.Put("/{departmentId}", departmentHandler.UpdateDepartment)
				r.Delete("/{departmentId}", departmentHandler.DeleteDepartment)
			})

			r.Route("/subjects", func(r chi.Router) {
				r.Get("/", subjectHandler.GetSubjects)
				r.Post("/", subjectHandler.CreateSubject)
				r.Get("/{id}", subjectHandler.GetSubjectByID)
				r.Put("/{id}", subjectHandler.UpdateSubject)
				r.Delete("/{id}", subjectHandler.DeleteSubject)
				r.Get("/{id}/alerts", subjectHandler.GetSubjectAlerts)
				r.Get("/{id}/materials", subjectHandler.GetSubjectMaterials)
				r.Get("/{id}/assignments", subjectHandler.GetSubjectAssignments)
			})

			r.Route("/classes", func(r chi.Router) {
				r.Get("/", classHandler.GetClasses)
				r.Post("/", classHandler.CreateClass)
				r.Post("/{class_id}/students", classHandler.AddStudentsToClass)
			})

			r.Route("/enrollments", func(r chi.Router) {
				r.Get("/", enrollmentHandler.GetEnrollments)
				r.Post("/onboard-student", enrollmentHandler.OnboardNewStudent)
				r.Post("/transfer/initiate", enrollmentHandler.InitiateStudentTransfer)
				r.Put("/transfer/{transferRequestId}/process", enrollmentHandler.ProcessIncomingTransfer)
			})

			r.Route("/attendance", func(r chi.Router) {
				r.Post("/mark", attendanceHandler.MarkAttendance)
				r.Get("/class/{class_id}", attendanceHandler.GetAttendanceByClass)
				r.Get("/student/{student_id}", attendanceHandler.GetStudentAttendance)
			})

			r.Route("/assignments", func(r chi.Router) {
				r.Get("/my", assignmentHandler.GetMyAssignments)
				r.Get("/{id}", assignmentHandler.GetAssignmentByID)
				r.Get("/class/{class_id}", assignmentHandler.GetAssignments)
				r.Post("/", assignmentHandler.CreateAssignment)
				r.Put("/{id}", assignmentHandler.UpdateAssignment)
				r.Delete("/{id}", assignmentHandler.DeleteAssignment)
			})

			r.Route("/banners", func(r chi.Router) {
				r.Get("/active", bannerHandler.GetActiveBanners)
				r.Group(func(r chi.Router) {
					r.Use(custom_mw.Authorize("Executive Administrator", "Developer"))
					r.Get("/", bannerHandler.ListBanners)
					r.Post("/", bannerHandler.CreateBanner)
					r.Put("/{id}", bannerHandler.UpdateBanner)
					r.Delete("/{id}", bannerHandler.DeleteBanner)
				})
			})

			r.Route("/audit-logs", func(r chi.Router) {
				r.Get("/", auditLogHandler.ListAuditLogs)
				r.Get("/{id}", auditLogHandler.GetAuditLog)
			})

			r.Route("/ews", func(r chi.Router) {
				r.Get("/at-risk", studentRiskHandler.ListAtRiskStudents)
				r.Post("/calculate", studentRiskHandler.TriggerRiskCalculation)
			})

			r.Route("/badges", func(r chi.Router) {
				r.Get("/", badgeHandler.ListBadgesBySchool)
				r.Post("/", badgeHandler.CreateBadge)
				r.Get("/{id}", badgeHandler.GetBadgeByID)
				r.Put("/{id}", badgeHandler.UpdateBadge)
				r.Delete("/{id}", badgeHandler.DeleteBadge)
				r.Post("/{badgeId}/award", badgeHandler.AwardBadge)
				r.Delete("/{badgeId}/revoke/{studentId}", badgeHandler.RevokeBadge)
			})

			r.Route("/billing-contacts", func(r chi.Router) {
				r.Get("/", billingContactHandler.GetBillingContacts)
				r.Post("/", billingContactHandler.CreateBillingContact)
				r.Get("/{id}", billingContactHandler.GetBillingContactByID)
				r.Put("/{id}", billingContactHandler.UpdateBillingContact)
				r.Delete("/{id}", billingContactHandler.DeleteBillingContact)
			})

			r.Route("/chat", func(r chi.Router) {
				r.Get("/rooms", chatHandler.GetChatRooms)
				r.Get("/rooms/{chat_room_id}/messages", chatHandler.GetChatMessages)
				r.Post("/message", chatHandler.SendMessage)
			})

			r.Route("/class-representatives", func(r chi.Router) {
				r.Get("/", classRepresentativeHandler.GetClassRepresentatives)
				r.Post("/", classRepresentativeHandler.CreateClassRepresentative)
				r.Get("/{id}", classRepresentativeHandler.GetClassRepresentativeByID)
				r.Put("/{id}", classRepresentativeHandler.UpdateClassRepresentative)
				r.Delete("/{id}", classRepresentativeHandler.DeleteClassRepresentative)
			})

			// Handlers for Batch 1
			r.Route("/events", func(r chi.Router) {
				r.Get("/", eventHandler.GetEvents)
				r.Post("/", eventHandler.CreateEvent)
				r.Get("/{id}", eventHandler.GetEventByID)
				r.Put("/{id}", eventHandler.UpdateEvent)
				r.Delete("/{id}", eventHandler.DeleteEvent)
			})
			r.Route("/learning-materials", func(r chi.Router) {
				r.Get("/", learningMaterialHandler.GetLearningMaterials)
				r.Post("/", learningMaterialHandler.CreateLearningMaterial)
				r.Get("/{id}", learningMaterialHandler.GetLearningMaterialByID)
				r.Put("/{id}", learningMaterialHandler.UpdateLearningMaterial)
				r.Delete("/{id}", learningMaterialHandler.DeleteLearningMaterial)
			})
			r.Route("/lesson-plans", func(r chi.Router) {
				r.Get("/", lessonPlanHandler.GetLessonPlans)
				r.Post("/", lessonPlanHandler.CreateLessonPlan)
				r.Get("/{id}", lessonPlanHandler.GetLessonPlanByID)
				r.Put("/{id}", lessonPlanHandler.UpdateLessonPlan)
				r.Delete("/{id}", lessonPlanHandler.DeleteLessonPlan)
			})
			r.Route("/meetings", func(r chi.Router) {
				r.Get("/", meetingHandler.GetMeetings)
				r.Post("/", meetingHandler.CreateMeeting)
				r.Get("/{id}", meetingHandler.GetMeetingByID)
				r.Put("/{id}", meetingHandler.UpdateMeeting)
				r.Delete("/{id}", meetingHandler.DeleteMeeting)
				r.Post("/{id}/attendees", meetingHandler.AddMeetingAttendees)
				r.Delete("/{id}/attendees", meetingHandler.RemoveMeetingAttendees)
			})
			r.Route("/newsletters", func(r chi.Router) {
				r.Get("/", newsletterHandler.GetNewsletters)
				r.Post("/", newsletterHandler.CreateNewsletter)
				r.Get("/{id}", newsletterHandler.GetNewsletterByID)
				r.Put("/{id}", newsletterHandler.UpdateNewsletter)
				r.Delete("/{id}", newsletterHandler.DeleteNewsletter)
			})
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notificationHandler.GetNotifications)
				r.Post("/", notificationHandler.CreateNotification)
				r.Put("/{id}/read", notificationHandler.MarkAsRead)
			})
			r.Route("/online-class-sessions", func(r chi.Router) {
				r.Get("/", onlineClassSessionHandler.GetOnlineClassSessions)
				r.Post("/", onlineClassSessionHandler.CreateOnlineClassSession)
				r.Get("/{id}", onlineClassSessionHandler.GetOnlineClassSessionByID)
				r.Put("/{id}", onlineClassSessionHandler.UpdateOnlineClassSession)
				r.Delete("/{id}", onlineClassSessionHandler.DeleteOnlineClassSession)
			})

			// Handlers for Batch 2
			r.Route("/external-certifications", func(r chi.Router) {
				r.Get("/", externalCertificationHandler.GetExternalCertifications)
				r.Post("/", externalCertificationHandler.CreateExternalCertification)
				r.Get("/{id}", externalCertificationHandler.GetExternalCertificationByID)
				r.Put("/{id}", externalCertificationHandler.UpdateExternalCertification)
				r.Delete("/{id}", externalCertificationHandler.DeleteExternalCertification)
			})
			r.Route("/fees", func(r chi.Router) {
				r.Get("/structures", feeHandler.GetFeeStructures)
				r.Post("/structures", feeHandler.CreateFeeStructure)
				r.Get("/student", feeHandler.GetStudentFees)
				r.Post("/student", feeHandler.CreateStudentFee)
			})
			r.Route("/feedback", func(r chi.Router) {
				r.Get("/", feedbackHandler.ListFeedbacks)
				r.Post("/", feedbackHandler.CreateFeedback)
				r.Get("/{id}", feedbackHandler.GetFeedbackByID)
				r.Put("/{id}", feedbackHandler.UpdateFeedback)
				r.Delete("/{id}", feedbackHandler.DeleteFeedback)
			})
			r.Route("/grades", func(r chi.Router) {
				r.Get("/submission/{submission_id}", gradeHandler.GetGradesBySubmission)
				r.Post("/", gradeHandler.CreateGrade)
			})
			r.Route("/groups", func(r chi.Router) {
				r.Post("/", groupHandler.CreateGroup)
				r.Post("/teacher", groupHandler.TeacherCreateGroup)
				r.Post("/invitations", groupHandler.RespondToGroupInvitation)
				r.Get("/{id}/members", groupHandler.GetGroupMembers)
			})
			r.Route("/parents", func(r chi.Router) {
				r.Get("/", parentHandler.GetParents)
				r.Get("/{id}", parentHandler.GetParentByID)
				r.Get("/children", parentHandler.GetChildren)

				// Parent monitoring routes (parent-only)
				r.Route("/children/{childId}", func(r chi.Router) {
					r.Use(custom_mw.Authorize("Parent"))
					r.Get("/attendance", parentHandler.GetChildAttendance)
					r.Get("/grades", parentHandler.GetChildGrades)
					r.Get("/assignments", parentHandler.GetChildAssignments)
					r.Get("/fees", parentHandler.GetChildFees)
				})
			})
			r.Route("/payments", func(r chi.Router) {
				r.Get("/", paymentHandler.GetPayments)
				r.Post("/", paymentHandler.CreatePayment)
			})
			r.Route("/quizzes", func(r chi.Router) {
				r.Get("/", quizHandler.GetQuizzes)
				r.Post("/", quizHandler.CreateQuiz)
				r.Get("/{id}", quizHandler.GetQuizByID)
				r.Put("/{id}", quizHandler.UpdateQuiz)
				r.Delete("/{id}", quizHandler.DeleteQuiz)
			})
			r.Route("/quiz-submissions", func(r chi.Router) {
				r.Get("/", quizSubmissionHandler.GetQuizSubmissions)
				r.Post("/", quizSubmissionHandler.CreateQuizSubmission)
				r.Get("/{id}", quizSubmissionHandler.GetQuizSubmissionByID)
				r.Patch("/{id}/grade", quizSubmissionHandler.GradeQuizSubmission)
			})
			r.Route("/rooms", func(r chi.Router) {
				r.Get("/", roomHandler.GetRooms)
				r.Post("/", roomHandler.CreateRoom)
				r.Get("/{id}", roomHandler.GetRoomByID)
				r.Put("/{id}", roomHandler.UpdateRoom)
				r.Delete("/{id}", roomHandler.DeleteRoom)
			})
			r.Route("/school-settings", func(r chi.Router) {
				r.Get("/", schoolSettingHandler.GetSchoolSettings)
				r.Put("/", schoolSettingHandler.UpdateSchoolSettings)
			})
			r.Route("/short-course-grades", func(r chi.Router) {
				r.Post("/", shortCourseGradeHandler.GradeShortCourse)
				r.Get("/", shortCourseGradeHandler.GetShortCourseGrades)
				r.Get("/{id}", shortCourseGradeHandler.GetShortCourseGradeByID)
			})
			r.Route("/students", func(r chi.Router) {
				r.Get("/", studentHandler.GetStudents)
				r.Get("/{id}", studentHandler.GetStudentByID)
			})
			r.Route("/student-course-progress", func(r chi.Router) {
				r.Get("/", studentCourseProgressHandler.GetStudentCourseProgresses)
				r.Post("/", studentCourseProgressHandler.CreateStudentCourseProgress)
				r.Get("/{id}", studentCourseProgressHandler.GetStudentCourseProgressByID)
				r.Put("/{id}", studentCourseProgressHandler.UpdateStudentCourseProgress)
				r.Delete("/{id}", studentCourseProgressHandler.DeleteStudentCourseProgress)
			})
			r.Route("/submissions", func(r chi.Router) {
				r.Get("/", submissionHandler.GetSubmissions)
				r.Post("/", submissionHandler.CreateSubmission)
				r.Get("/{id}", submissionHandler.GetSubmissionByID)
				r.Put("/{id}/status", submissionHandler.UpdateSubmissionStatus)
			})
			r.Route("/subscription-tiers", func(r chi.Router) {
				r.Get("/", subscriptionTierHandler.GetSubscriptionTiers)
				r.Get("/{id}", subscriptionTierHandler.GetSubscriptionTierByID)
			})
			r.Route("/teacher-availability", func(r chi.Router) {
				r.Get("/", teacherAvailabilityHandler.GetTeacherAvailabilities)
				r.Post("/", teacherAvailabilityHandler.CreateTeacherAvailability)
				r.Get("/{id}", teacherAvailabilityHandler.GetTeacherAvailabilityByID)
				r.Put("/{id}", teacherAvailabilityHandler.UpdateTeacherAvailability)
				r.Delete("/{id}", teacherAvailabilityHandler.DeleteTeacherAvailability)
			})
			r.Route("/teachers", func(r chi.Router) {
				r.Get("/", teacherHandler.GetTeachers)
				r.Get("/{id}", teacherHandler.GetTeacherByID)
			})
			r.Route("/teacher-workloads", func(r chi.Router) {
				r.Get("/", teacherWorkloadHandler.GetTeacherWorkloads)
				r.Post("/", teacherWorkloadHandler.CreateTeacherWorkload)
				r.Get("/{id}", teacherWorkloadHandler.GetTeacherWorkloadByID)
				r.Put("/{id}", teacherWorkloadHandler.UpdateTeacherWorkload)
				r.Delete("/{id}", teacherWorkloadHandler.DeleteTeacherWorkload)
			})
			r.Route("/timetables", func(r chi.Router) {
				r.Get("/", timetableHandler.GetTimetables)
				r.Post("/", timetableHandler.CreateTimetable)
				r.Post("/generate", timetableHandler.GenerateTimetable)
				r.Get("/entries", timetableHandler.GetTimetableEntries)
			})
			r.Route("/transcripts", func(r chi.Router) {
				r.Get("/", transcriptHandler.GetTranscripts)
				r.Post("/", transcriptHandler.CreateTranscript)
				r.Get("/{id}", transcriptHandler.GetTranscriptByID)
				r.Put("/{id}", transcriptHandler.UpdateTranscript)
				r.Delete("/{id}", transcriptHandler.DeleteTranscript)
			})
			r.Route("/transfer-requests", func(r chi.Router) {
				r.Get("/", transferRequestHandler.GetTransferRequests)
				r.Post("/", transferRequestHandler.CreateTransferRequest)
				r.Get("/{id}", transferRequestHandler.GetTransferRequestByID)
				r.Put("/{id}", transferRequestHandler.UpdateTransferRequest)
				r.Delete("/{id}", transferRequestHandler.DeleteTransferRequest)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
