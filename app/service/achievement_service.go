package service

import (
	"math"
	"uas/app/model"
	"uas/app/repository"
	"github.com/gofiber/fiber/v2"
)

// FR-010 & Modul 6: Get All Achievements (Pagination, Sort, Search)
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	// 1. Parsing Query Parameter (Modul 6)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	// Masukkan ke struct param
	param := model.PaginationParam{
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  order,
		Search: search,
	}

	// 2. Cek Role Login (RBAC Logic)
	// Jika Mahasiswa -> Hanya lihat data sendiri
	// Jika Admin/Dosen -> Bisa lihat semua (logic disederhanakan)
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)
	
	filterStudentID := ""
	if userRole == "Mahasiswa" {
		// Cari studentID berdasarkan userID
		student, err := s.userRepo.FindStudentByUserID(userID)
		if err == nil {
			filterStudentID = student.ID
		}
	}

	// 3. Panggil Repository
	data, total, err := s.achRepo.FindAll(param, filterStudentID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{
			Code:    500,
			Status:  "error",
			Message: err.Error(),
		})
	}

	// 4. Hitung Metadata Pagination (Modul 6)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return Response
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Data berhasil diambil",
		Data:    data,
		Meta: &model.MetaInfo{
			Page:      page,
			Limit:     limit,
			TotalData: total,
			TotalPage: totalPages,
			SortBy:    sortBy,
			Order:     order,
			Search:    search,
		},
	})
}

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
        Title:     req.Title, // PENTING: Simpan judul di Postgres untuk sorting!
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