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
	// 5.1 Authentication
	// =================================================================
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)
	auth.Post("/logout", authService.Logout)
	// Profile (Butuh Token)
	auth.Get("/profile", authMiddleware.AuthRequired(), authService.GetProfile)

	// =================================================================
	// 5.2 Users (Admin Only)
	// =================================================================
	// Endpoint ini hanya bisa diakses oleh user dengan permission "user:manage" (biasanya Admin)
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
	// 5.4 Achievements
	// =================================================================
	ach := api.Group("/achievements", authMiddleware.AuthRequired())

	// Public (Authenticated) - List & Detail
	ach.Get("/", achService.GetAll)
	ach.Get("/:id", achService.GetDetail)
	ach.Get("/:id/history", achService.GetHistory)

	// Mahasiswa Actions (Butuh permission spesifik)
	ach.Post("/", authMiddleware.PermissionRequired("achievement:create"), achService.Submit)
	ach.Put("/:id", authMiddleware.PermissionRequired("achievement:update"), achService.Update)
	ach.Delete("/:id", authMiddleware.PermissionRequired("achievement:delete"), achService.Delete)
	ach.Post("/:id/submit", authMiddleware.PermissionRequired("achievement:create"), achService.RequestVerification)
	ach.Post("/:id/attachments", authMiddleware.PermissionRequired("achievement:update"), achService.UploadAttachment)

	// Dosen Actions (Verify/Reject)
	ach.Post("/:id/verify", authMiddleware.PermissionRequired("achievement:verify"), achService.Verify)
	ach.Post("/:id/reject", authMiddleware.PermissionRequired("achievement:verify"), achService.Reject)

	// =================================================================
	// 5.5 Students & Lecturers
	// =================================================================
	// Group Students
	std := api.Group("/students", authMiddleware.AuthRequired())
	std.Get("/", authService.GetAllStudents)
	
	// Create/Update Student Profile (Admin Only)
	std.Post("/", authMiddleware.PermissionRequired("user:manage"), authService.SetStudentProfile)
	
	std.Get("/:id", authService.GetStudentDetail)
	std.Get("/:id/achievements", achService.GetStudentAchievements)
	
	// Set Advisor (Admin Only)
	std.Put("/:id/advisor", authMiddleware.PermissionRequired("user:manage"), authService.UpdateStudentAdvisor)

	// Group Lecturers
	lec := api.Group("/lecturers", authMiddleware.AuthRequired())
	lec.Get("/", authService.GetAllLecturers)
	
	// Create/Update Lecturer Profile (Admin Only)
	lec.Post("/", authMiddleware.PermissionRequired("user:manage"), authService.SetLecturerProfile)
	
	// Get Advisee List (Untuk Dosen melihat mahasiswa bimbingannya)
	// Kita gunakan 'me' atau ID spesifik, logic di service akan menangani
	lec.Get("/:id/advisees", achService.GetAdviseeAchievements)

	// =================================================================
	// 5.8 Reports & Analytics
	// =================================================================
	report := api.Group("/reports", authMiddleware.AuthRequired())
	report.Get("/statistics", achService.GetStatistics)
	report.Get("/student/:id", achService.GetStudentStatistics)
}