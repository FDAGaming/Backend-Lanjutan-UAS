package service

import (
	"project-uas/app/model"
	"project-uas/app/repository"
	"github.com/gofiber/fiber/v2" // Asumsi pakai Fiber
)

type AchievementService struct {
	repo *repository.AchievementRepository
}

func NewAchievementService(repo *repository.AchievementRepository) *AchievementService {
	return &AchievementService{repo: repo}
}

[cite_start]// FR-003: Submit Prestasi [cite: 178]
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	// 1. Parsing Input JSON
	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// 2. Ambil User ID dari Token (Context Middleware)
	// Pastikan middleware JWT sudah menyimpan 'user_id' di Locals
	userID := c.Locals("user_id").(string) 

	// 3. Siapkan Object untuk MongoDB
	mongoData := model.AchievementContent{
		StudentID:       userID,
		Title:           req.Title,
		AchievementType: req.AchievementType,
		Description:     req.Description,
		Details:         req.Details, // Field dinamis masuk sini
	}

	// 4. Siapkan Object untuk PostgreSQL
	pgData := model.AchievementMeta{
		StudentID: userID,
		// MongoAchievementID diisi di repository nanti
	}

	// 5. Panggil Repo untuk simpan ke dua DB
	if err := s.repo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 6. Return Success
	return c.Status(201).JSON(fiber.Map{
		"status": "success",
		"message": "Prestasi berhasil disimpan sebagai draft",
		"data": fiber.Map{
			"metaId": pgData.ID, // ID Postgres
			"mongoId": pgData.MongoAchievementID, // ID Mongo
		},
	})
}