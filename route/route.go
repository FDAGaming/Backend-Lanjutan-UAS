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
	// 5.1 Authentication (FR-001)
	// =================================================================
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)
	auth.Post("/logout", authService.Logout)
	
	// Profile (Butuh Token)
	auth.Get("/profile", authMiddleware.AuthRequired(), authService.GetProfile)

	// =================================================================
	// 5.2 Users Management (FR-009: Admin Only)
	// =================================================================
	// Endpoint ini dilindungi permission "user:manage" (Admin)
	users := api.Group("/users",
		authMiddleware.AuthRequired(),
		authMiddleware.PermissionRequired("user:manage"),
	)
	users.Get("/", authService.GetAllUsers)         // List All Users
	users.Post("/", authService.CreateUser)         // Create User
	users.Get("/:id", authService.GetUserDetail)    // Detail User
	users.Put("/:id", authService.UpdateUser)       // Update General Info
	users.Delete("/:id", authService.DeleteUser)    // Delete User
	users.Put("/:id/role", authService.UpdateUserRole) // Assign Role

	// =================================================================
	// 5.4 Achievements (FR-003, FR-004, FR-005, FR-007, FR-008, FR-010)
	// =================================================================
	ach := api.Group("/achievements", authMiddleware.AuthRequired())

	// Public (Authenticated) - List & Detail (FR-010 Admin View All included here)
	ach.Get("/", achService.GetAll)
	ach.Get("/:id", achService.GetDetail)
	ach.Get("/:id/history", achService.GetHistory)

	// Mahasiswa Actions (Butuh permission spesifik)
	ach.Post("/", authMiddleware.PermissionRequired("achievement:create"), achService.Submit, // FR-003 Submit Draft
	)
	ach.Put("/:id", authMiddleware.PermissionRequired("achievement:update"), achService.Update,
	)
	ach.Delete("/:id", authMiddleware.PermissionRequired("achievement:delete"), achService.Delete, // FR-005 Delete Draft
	)
	ach.Post("/:id/submit", authMiddleware.PermissionRequired("achievement:create"), achService.RequestVerification, // FR-004 Submit for Verification
	)
	ach.Post("/:id/attachments", authMiddleware.PermissionRequired("achievement:update"), achService.UploadAttachment,
	)

	// Dosen Actions (Verify/Reject)
	ach.Post("/:id/verify", authMiddleware.PermissionRequired("achievement:verify"), achService.Verify, // FR-007 Verify
	)
	ach.Post("/:id/reject", authMiddleware.PermissionRequired("achievement:verify"), achService.Reject, // FR-008 Reject
	)

	// =================================================================
	// 5.5 Students & Lecturers (Manage Profiles - FR-009)
	// =================================================================
	// Group Students
	std := api.Group("/students", authMiddleware.AuthRequired())
	
	// Create/Update Student Profile (Admin Only)
	std.Post("/", authMiddleware.PermissionRequired("user:manage"), authService.SetStudentProfile)
	
	// Set Advisor (Admin Only)
	std.Put("/:id/advisor", authMiddleware.PermissionRequired("user:manage"), authService.UpdateStudentAdvisor)
	
	std.Get("/", authService.GetAllStudents)
	std.Get("/:id", authService.GetStudentDetail)
	std.Get("/:id/achievements", achService.GetStudentAchievements)

	// Group Lecturers
	lec := api.Group("/lecturers", authMiddleware.AuthRequired())
	
	// Create/Update Lecturer Profile (Admin Only)
	lec.Post("/", authMiddleware.PermissionRequired("user:manage"), authService.SetLecturerProfile)
	
	lec.Get("/", authService.GetAllLecturers)
	
	// Get Advisee List (FR-006: View Prestasi Bimbingan)
	// Bisa diakses Dosen untuk melihat mahasiswa bimbingannya sendiri
	lec.Get("/:id/advisees", achService.GetAdviseeAchievements)

	// =================================================================
	// 5.8 Reports & Analytics (FR-011)
	// =================================================================
	report := api.Group("/reports", authMiddleware.AuthRequired())
	
	// General Stats (Admin/Dosen/All)
	report.Get("/statistics", achService.GetStatistics)
	
	// Student Specific Stats
	report.Get("/student/:id", achService.GetStudentStatistics)
}