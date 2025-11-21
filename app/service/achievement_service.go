package service

import (
	"uas/app/model"
	"uas/app/repository"
	"github.com/gofiber/fiber/v2"
)

type AchievementService struct {
	achRepo  *repository.AchievementRepository
	userRepo *repository.UserRepository
}

func NewAchievementService(achRepo *repository.AchievementRepository, userRepo *repository.UserRepository) *AchievementService {
	return &AchievementService{
		achRepo:  achRepo,
		userRepo: userRepo,
	}
}

// FR-003: Submit Prestasi (Mahasiswa)
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	// 1. Parse Body
	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input data"})
	}

	// 2. Ambil User ID dari Context (set via Middleware JWT)
	userID := c.Locals("user_id").(string)

	// 3. Cari Student ID berdasarkan User ID
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"})
	}

	// 4. Mapping ke Model Mongo
	content := model.AchievementContent{
		StudentID:       student.ID, // Pakai Student ID, bukan User ID
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
	}

	// 5. Mapping ke Model Postgres
	ref := model.AchievementReference{
		StudentID: student.ID,
		Status:    "draft",
	}

	// 6. Simpan ke Database (Repo Hybrid)
	if err := s.achRepo.Create(c.Context(), &content, &ref); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Achievement saved successfully",
		"data":    fiber.Map{"id": ref.ID, "status": ref.Status},
	})
}

// FR-007 & FR-008: Verify / Reject Prestasi (Dosen Wali)
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	achID := c.Params("id")
	
	var req model.VerifyAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Ambil User ID Dosen dari Token
	userID := c.Locals("user_id").(string)
	
	// Validasi apakah user ini benar dosen? (Logic simplifikasi)
	// Sebaiknya cek via repo apakah dosen ini adalah wali mahasiswa tsb
	
	if err := s.achRepo.UpdateStatus(achID, req.Status, userID, req.RejectionNote); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Achievement status updated to " + req.Status})
}

// FR-006: Get Detail (Gabungan SQL + Mongo)
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	achID := c.Params("id")

	ref, content, err := s.achRepo.FindDetail(c.Context(), achID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Merge Response Manual
	return c.JSON(fiber.Map{
		"meta":    ref,
		"content": content,
	})
}