package main

import (
	"log"
	"os"

	"uas/middleware"
	"uas/app/repository"
	"uas/route"
	"uas/app/service"
	"uas/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using system environment variables")
	}

	// 2. Initialize Database (Hybrid: Postgres & Mongo)
	// Config ini otomatis melakukan AutoMigrate untuk Postgres
	db := config.InitDB()

	// 3. Setup Repositories (Data Access Layer)
	// ---------------------------------------------------------
	// RoleRepo: Untuk akses data Role & Permissions (RBAC)
	roleRepo := repository.NewRoleRepository(db.Postgres)
	
	// UserRepo: Untuk akses data User, Mahasiswa, Dosen
	userRepo := repository.NewUserRepository(db.Postgres)
	
	// AchRepo: Butuh DUA koneksi (Postgres untuk relasi, Mongo untuk data dinamis)
	achRepo := repository.NewAchievementRepository(db.Postgres, db.Mongo)

	// 4. Setup Services (Business Logic Layer)
	// ---------------------------------------------------------
	// AuthService: Butuh UserRepo & RoleRepo (untuk inject permissions ke token saat login)
	authService := service.NewAuthService(userRepo, roleRepo)
	
	// AchService: Butuh AchRepo & UserRepo (untuk validasi profil mahasiswa/dosen)
	achService := service.NewAchievementService(achRepo, userRepo)

	// 5. Setup Middleware
	// ---------------------------------------------------------
	// AuthMiddleware: Butuh RoleRepo (jika ingin validasi permission level DB strict)
	authMiddleware := middleware.NewAuthMiddleware(roleRepo)

	// 6. Initialize Fiber App
	// ---------------------------------------------------------
	app := fiber.New(fiber.Config{
		// Custom Error Handler agar response JSON rapi jika terjadi panic/error framework
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"code":    code,
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// Default Middlewares
	app.Use(logger.New()) // Log request ke terminal
	app.Use(cors.New())   // Enable CORS untuk akses dari Frontend

	// 7. Setup Routes (Wiring Semua Komponen)
	// ---------------------------------------------------------
	// Kita kirimkan app, services, dan middleware ke file route
	route.SetupRoutes(app, authService, achService, authMiddleware)

	// 8. Start Server
	// ---------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	log.Println("🚀 Server running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}