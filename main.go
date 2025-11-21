package main

import (
	"log"
	"uas/app/repository"
	"uas/app/service"
	"uas/database"
	
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
    // 1. Load ENV
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // 2. Init Database (Otomatis connect PG & Mongo)
    db := config.InitDB()

	// 3. Setup Repository (Inject kedua DB)
	userRepo := repository.NewUserRepository(db.Postgres)
	// Repo Achievement butuh DUA database
	achRepo := repository.NewAchievementRepo(db.Postgres, db.Mongo)

	// 4. Setup Service (Inject Repo)
	authService := service.NewAuthService(userRepo)
	achService := service.NewAchievementService(achRepo, userRepo)

	// 5. Setup Fiber & Routes
	app := fiber.New()
	
	api := app.Group("/api/v1")
	api.Post("/auth/login", authService.Login)
	api.Post("/achievements", achService.Submit) // Middleware JWT perlu ditambahkan

	// Start Server
	log.Fatal(app.Listen(":8080"))
}