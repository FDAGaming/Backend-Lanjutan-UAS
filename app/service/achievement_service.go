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

// POST /api/v1/achievements (FR-003: Submit Prestasi - Draft)
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input: " + err.Error()})
	}

	userID := c.Locals("user_id").(string)
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found. Are you a student?"})
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
		Points:          req.Points,
	}

	pgData := model.AchievementReference{
		StudentID: student.ID,
		Title:     req.Title,
		Status:    "draft",
	}

	if err := s.achRepo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to submit achievement: " + err.Error()})
	}

	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "Prestasi berhasil disimpan sebagai draft",
		Data: fiber.Map{
			"referenceId": pgData.ID,
			"mongoId":     pgData.MongoAchievementID,
			"status":      pgData.Status,
			"points":      mongoData.Points,
		},
	})
}

// GET /api/v1/achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
    param := s.parsePagination(c)
    userRole := c.Locals("role").(string) // Ambil Role dari Token
    userID := c.Locals("user_id").(string)

    var filterStudent, filterAdvisor string

    // Logic Filter Berdasarkan Role (RBAC Data Level)
    if userRole == "Mahasiswa" {
        student, _ := s.userRepo.FindStudentByUserID(userID)
        if student != nil {
            filterStudent = student.ID
        }
    } else if userRole == "Dosen Wali" {
        lecturer, _ := s.userRepo.FindLecturerByUserID(userID)
        if lecturer != nil {
            filterAdvisor = lecturer.ID
        }
    }
    // JIKA ADMIN: Kode akan melewati if/else di atas,
    // sehingga filterStudent & filterAdvisor tetap string kosong ("").
    // Akibatnya, Repository akan mengambil SEMUA data tanpa filter WHERE.

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

// POST /api/v1/achievements/:id/submit
func (s *AchievementService) RequestVerification(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You do not own this achievement"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only draft achievement can be submitted"})
	}

	if err := s.achRepo.UpdateStatus(id, "submitted", "", "", 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to update status: " + err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil diajukan untuk verifikasi",
		Data: fiber.Map{
			"id":     id,
			"status": "submitted",
		},
	})
}

// ---------------------------------------------------------
// FR-007: Verify Prestasi
// ---------------------------------------------------------
// POST /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// 1. Dosen Input Poin (Opsional, jika ada revisi poin)
	var req struct{ Points int `json:"points"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	// 2. Ambil ID Dosen (Verifier)
	userID := c.Locals("user_id").(string)

	// Validasi apakah user benar-benar Dosen (Optional, krn Middleware RBAC sudah cek role)
	_, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: Only Lecturer can verify"})
	}

	// 3. Cek Data Prestasi & Precondition Status
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// CHECK PRECONDITION: Status 'submitted'
	if ref.Status != "submitted" {
		return c.Status(400).JSON(model.WebResponse{
			Code:    400, 
			Status:  "error", 
			Message: "Cannot verify. Achievement status must be 'submitted', current status is: " + ref.Status,
		})
	}

	// 4. Update Status -> 'verified' & Set Points
	if err := s.achRepo.UpdateStatus(id, "verified", userID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	// 5. Return Updated Status
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil diverifikasi",
		Data: fiber.Map{
			"id":     id,
			"status": "verified",
			"points": req.Points,
		},
	})
}

// ---------------------------------------------------------
// FR-008: Reject Prestasi
// ---------------------------------------------------------
// POST /api/v1/achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// 1. Dosen Input Rejection Note (Wajib)
	var req struct{ Note string `json:"note"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}
	if req.Note == "" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Rejection note is required"})
	}

	userID := c.Locals("user_id").(string)

	// 2. Cek Data & Precondition
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	if ref.Status != "submitted" {
		return c.Status(400).JSON(model.WebResponse{
			Code:    400, 
			Status:  "error", 
			Message: "Cannot reject. Achievement status must be 'submitted', current status is: " + ref.Status,
		})
	}

	// 3. Update Status -> 'rejected'
	if err := s.achRepo.UpdateStatus(id, "rejected", userID, req.Note, 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi ditolak",
		Data: fiber.Map{
			"id":     id,
			"status": "rejected",
		},
	})
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

func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Feature Not Implemented"})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "History Not Implemented"})
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Upload Not Implemented"})
}

// GET /api/v1/lecturers/:id/advisees
// FR-006: View Prestasi Mahasiswa Bimbingan
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	// 1. Ambil User ID Dosen dari Token
	userID := c.Locals("user_id").(string)

	// 2. Cari Profil Dosen (Lecturer) berdasarkan User ID
	lecturer, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// 3. Parse Parameter Pagination (Page, Limit, Search, Sort)
	param := s.parsePagination(c)

	// 4. Panggil Repo FindAll dengan Filter AdvisorID
	// Parameter ke-2 (studentID) kosong karena kita mau semua mahasiswa bimbingan, bukan spesifik satu
	// Parameter ke-3 (advisorID) diisi ID Dosen yang sedang login
	data, total, err := s.achRepo.FindAll(param, "", lecturer.ID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	// 5. Return Response dengan Pagination
	return s.sendPaginationResponse(c, data, total, param)
}	

func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Student Achievement List Not Implemented"})
}

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