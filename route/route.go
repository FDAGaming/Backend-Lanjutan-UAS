package route

import (
	"uas/middleware"
	"uas/app/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authService *service.AuthService,
	achService *service.AchievementService,
	authMiddleware *middleware.AuthMiddleware,
) {
	api := app.Group("/api/v1")

	// =================================================================
	// 5.1 Authentication [cite: 723-727]
	// =================================================================
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken) // Logic di service
	auth.Post("/logout", authService.Logout)        // Logic di service
	
	// Profile (Butuh Token)
	auth.Get("/profile", 
		authMiddleware.AuthRequired(), 
		authService.GetProfile,
	)

	// =================================================================
	// 5.2 Users (Admin Only) [cite: 728-734]
	// =================================================================
	// Karena kita belum memisahkan UserService, kita gunakan AuthService sementara
	// atau method Not Implemented di AuthService
	users := api.Group("/users", 
		authMiddleware.AuthRequired(), 
		authMiddleware.PermissionRequired("user:manage"),
	)
	
	users.Get("/", authService.GetAllUsers)
	users.Get("/:id", authService.GetUserDetail)
	users.Post("/", authService.CreateUser)
	users.Put("/:id", authService.UpdateUser)
	users.Delete("/:id", authService.DeleteUser)
	users.Put("/:id/role", authService.UpdateUserRole)

	// =================================================================
	// 5.4 Achievements [cite: 735-746]
	// =================================================================
	ach := api.Group("/achievements", authMiddleware.AuthRequired())

	// List (filtered by role logic inside service)
	ach.Get("/", achService.GetAll)

	// Detail
	ach.Get("/:id", achService.GetDetail)

	// Create (Mahasiswa)
	ach.Post("/", 
		authMiddleware.PermissionRequired("achievement:create"), 
		achService.Submit,
	)

	// Update (Mahasiswa)
	ach.Put("/:id", 
		authMiddleware.PermissionRequired("achievement:update"), 
		achService.Update,
	)

	// Delete (Mahasiswa)
	ach.Delete("/:id", 
		authMiddleware.PermissionRequired("achievement:delete"), 
		achService.Delete,
	)

	// Submit for verification
	ach.Post("/:id/submit", 
		authMiddleware.PermissionRequired("achievement:create"), 
		achService.RequestVerification,
	)

	// Verify (Dosen Wali)
	ach.Post("/:id/verify", 
		authMiddleware.PermissionRequired("achievement:verify"), 
		achService.Verify,
	)

	// Reject (Dosen Wali)
	ach.Post("/:id/reject", 
		authMiddleware.PermissionRequired("achievement:verify"), 
		achService.Reject,
	)

	// Status history
	ach.Get("/:id/history", achService.GetHistory)

	// Upload files
	ach.Post("/:id/attachments", achService.UploadAttachment)

	// =================================================================
	// 5.5 Students & Lecturers [cite: 747-753]
	// =================================================================
	std := api.Group("/students", authMiddleware.AuthRequired())
	std.Get("/", authService.GetAllStudents)
	std.Get("/:id", authService.GetStudentDetail)
	std.Get("/:id/achievements", achService.GetStudentAchievements) // Logic di AchievementService
	std.Put("/:id/advisor", 
		authMiddleware.PermissionRequired("user:manage"), 
		authService.UpdateStudentAdvisor,
	)

	lec := api.Group("/lecturers", authMiddleware.AuthRequired())
	lec.Get("/", authService.GetAllLecturers)
	lec.Get("/:id/advisees", 
		// Ini mirip GetAdviseeAchievements tapi spesifik ID dosen tertentu
		achService.GetAdviseeAchievements, 
	)

	// =================================================================
	// 5.8 Reports & Analytics [cite: 754-756]
	// =================================================================
	report := api.Group("/reports", authMiddleware.AuthRequired())
	report.Get("/statistics", achService.GetStatistics)
	report.Get("/student/:id", achService.GetStudentStatistics)
}