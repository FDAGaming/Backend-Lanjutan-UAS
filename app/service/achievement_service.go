package service

import (
	"math"
	"uas/app/model"
	"uas/app/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

// ---------------------------------------------------------
// 5.4 ACHIEVEMENTS
// ---------------------------------------------------------

// GET /api/v1/achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	param := s.parsePagination(c)
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	var filterStudent, filterAdvisor string

	if userRole == "Mahasiswa" {
		student, _ := s.userRepo.FindStudentByUserID(userID)
		if student != nil {
			filterStudent = student.ID
		}
	} else if userRole == "Dosen Wali" {
		// Asumsi Dosen Wali melihat bimbingannya di menu ini juga? 
		// Atau bisa jadi kosong jika ingin melihat semua
	}

	data, total, err := s.achRepo.FindAll(param, filterStudent, filterAdvisor)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return s.sendPaginationResponse(c, data, total, param)
}

// GET /api/v1/achievements/:id
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, content, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Detail retrieved",
		Data: fiber.Map{
			"meta":    ref,
			"content": content,
		},
	})
}

// POST /api/v1/achievements (Create Draft)
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	userID := c.Locals("user_id").(string)
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	mongoData := model.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       student.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
	}

	pgData := model.AchievementReference{
		StudentID: student.ID,
		Title:     req.Title,
		Status:    "draft",
	}

	if err := s.achRepo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "Prestasi disimpan sebagai draft",
		Data:    pgData,
	})
}

// PUT /api/v1/achievements/:id
func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Feature Not Implemented"})
}

// DELETE /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)
	student, _ := s.userRepo.FindStudentByUserID(userID)
	
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Not found"})
	}
	
	if student != nil && ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Forbidden"})
	}

	if err := s.achRepo.Delete(c.Context(), id); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi berhasil dihapus"})
}

// POST /api/v1/achievements/:id/submit
func (s *AchievementService) RequestVerification(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.achRepo.UpdateStatus(id, "submitted", "", "", 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi diajukan"})
}

// POST /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct{ Points int `json:"points"` }
	c.BodyParser(&req)
	userID := c.Locals("user_id").(string)

	if err := s.achRepo.UpdateStatus(id, "verified", userID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi diverifikasi"})
}

// POST /api/v1/achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct{ Note string `json:"note"` }
	c.BodyParser(&req)
	userID := c.Locals("user_id").(string)

	if err := s.achRepo.UpdateStatus(id, "rejected", userID, req.Note, 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi ditolak"})
}

// GET /api/v1/achievements/:id/history
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "History Not Implemented"})
}

// POST /api/v1/achievements/:id/attachments
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Upload Not Implemented"})
}

// GET /api/v1/lecturers/:id/advisees (Spesifik Dosen tertentu)
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	// Logic ini bisa menggunakan filter advisorID di repo
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Advisee List Not Implemented"})
}

// GET /api/v1/students/:id/achievements
func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	// Logic ini bisa menggunakan filter studentID di repo
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Student Achievement List Not Implemented"})
}

// ---------------------------------------------------------
// 5.8 REPORTS & ANALYTICS
// ---------------------------------------------------------

func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	stats := map[string]interface{}{
		"total_per_type": map[string]int{"academic": 10, "competition": 5},
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Data: stats})
}

func (s *AchievementService) GetStudentStatistics(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Student Stats Not Implemented"})
}

// HELPER
func (s *AchievementService) parsePagination(c *fiber.Ctx) model.PaginationParam {
	return model.PaginationParam{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		SortBy: c.Query("sortBy", "created_at"),
		Order:  c.Query("order", "desc"),
		Search: c.Query("search", ""),
	}
}

func (s *AchievementService) sendPaginationResponse(c *fiber.Ctx, data interface{}, total int64, param model.PaginationParam) error {
	totalPages := int(math.Ceil(float64(total) / float64(param.Limit)))
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Data retrieved",
		Data:    data,
		Meta: &model.MetaInfo{
			Page:      param.Page,
			Limit:     param.Limit,
			TotalData: total,
			TotalPage: totalPages,
			SortBy:    param.SortBy,
			Order:     param.Order,
			Search:    param.Search,
		},
	})
}