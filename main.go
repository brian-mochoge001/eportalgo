package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
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
	"google.golang.org/api/option"
)

func main() {
	// Load environment variables from .env file
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

	// Initialize Firebase Admin SDK
	credsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	
	// If env var is missing, try common locations
	if credsPath == "" {
		possiblePaths := []string{
			"/etc/secrets/eschool-infinnitydevelopers-firebase-adminsdk.json",
			"eschool-infinnitydevelopers-firebase-adminsdk.json",
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				credsPath = p
				fmt.Printf("Auto-detected Firebase credentials at: %s\n", p)
				break
			}
		}
	}

	if credsPath == "" {
		fmt.Println("DEBUG: FIREBASE_CREDENTIALS_PATH is empty and no fallback files found. Available environment variables (keys only):")
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			fmt.Printf("- %s\n", pair[0])
		}
		log.Fatal("FIREBASE_CREDENTIALS_PATH environment variable is required or credentials must exist at /etc/secrets/eschool-infinnitydevelopers-firebase-adminsdk.json")
	}

	// Verify file exists
	info, err := os.Stat(credsPath)
	if err != nil {
		log.Fatalf("Firebase credentials file not found at %s: %v", credsPath, err)
	}
	fmt.Printf("Firebase credentials file found. Size: %d bytes\n", info.Size())

	// Force GOOGLE_APPLICATION_CREDENTIALS for underlying libraries
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsPath)

	opt := option.WithCredentialsFile(credsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase App: %v", err)
	}

	firebaseAuth, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v", err)
	}

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

	// Initialize Services
	schoolService := services.NewSchoolService(queries, firebaseAuth, redisClient)
	authService := services.NewAuthService(queries, firebaseAuth)
	studentService := services.NewStudentService(queries, conn)
	assignmentService := services.NewAssignmentService(queries, asynqClient)
	timetableService := services.NewTimetableService(queries, conn)
	userService := services.NewUserService(queries, conn, firebaseAuth)
	meetingService := services.NewMeetingService(queries, conn)
	courseService := services.NewCourseService(queries, conn)
	eventService := services.NewEventService(queries, conn)
	lessonPlanService := services.NewLessonPlanService(queries, conn)
	classService := services.NewClassService(queries, conn)
	attendanceService := services.NewAttendanceService(queries)
	financeService := services.NewFinanceService(queries, conn)
	quizService := services.NewQuizService(queries, conn)
	reportingService := services.NewReportingService(queries)

	// Initialize Handlers
	schoolHandler := handlers.NewSchoolHandler(queries, schoolService, redisClient)
	authHandler := handlers.NewAuthHandler(queries, authService, firebaseAuth)
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

	// Start Asynq server
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
			r.Post("/login", authHandler.LoginUser)
		})

		r.Route("/schools", func(r chi.Router) {
			r.Post("/register", schoolHandler.RegisterSchool)
			r.Group(func(r chi.Router) {
				r.Use(custom_mw.AuthMiddleware(firebaseAuth, queries))
				r.Put("/{schoolId}/verify", schoolHandler.VerifySchool)
				r.Get("/{schoolId}/settings", schoolHandler.GetSchoolSettings)
				r.Put("/{schoolId}/settings", schoolHandler.UpdateSchoolSettings)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(custom_mw.AuthMiddleware(firebaseAuth, queries))
			r.Use(custom_mw.RLSMiddleware(conn))
			r.Use(custom_mw.AuditMiddleware(asynqClient))

			r.Route("/users", func(r chi.Router) {
				r.Post("/add", userHandler.AddUser)
				r.Get("/school/{schoolId}", userHandler.GetUsersBySchool)
				r.Post("/{userId}/student-profile", userHandler.AddStudentProfile)
				r.Post("/{userId}/parent-profile", userHandler.AddParentProfile)
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
				r.Get("/class/{class_id}", assignmentHandler.GetAssignments)
				r.Post("/", assignmentHandler.CreateAssignment)
				r.Put("/{id}", assignmentHandler.UpdateAssignment)
				r.Delete("/{id}", assignmentHandler.DeleteAssignment)
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
				r.Get("/", badgeHandler.GetBadges)
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
