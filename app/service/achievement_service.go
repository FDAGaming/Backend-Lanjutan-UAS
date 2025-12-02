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
	// 1. Parse Input
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input: " + err.Error()})
	}

	// 2. Ambil User ID dari Token (Middleware)
	userID := c.Locals("user_id").(string)

	// Validasi: Pastikan User adalah Mahasiswa dan punya profil Student
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		// Error 404 ini muncul jika user login tapi tidak ada di tabel 'students'
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found. Are you a student?"})
	}

	// 3. Siapkan Data MongoDB (Konten Detail)
	mongoData := model.Achievement{
		ID:              primitive.NewObjectID(), // Generate ID baru untuk Mongo
		StudentID:       student.ID,              // Link ke ID Student (Postgres UUID)
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,     // Field dinamis (Juara, Lomba, dll)
		Attachments:     req.Attachments, // File bukti
		Tags:            req.Tags,
		Points:          req.Points, // [UPDATED] Mengambil poin dari input user (sebelumnya 0)
	}

	// 4. Siapkan Data PostgreSQL (Referensi Status)
	pgData := model.AchievementReference{
		StudentID: student.ID,
		Title:     req.Title,   // Disimpan juga di SQL untuk searching/sorting
		Status:    "draft",     // Status Awal sesuai FR-003
	}

	// 5. Simpan ke Database (Hybrid Transaction di Repository)
	if err := s.achRepo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to submit achievement: " + err.Error()})
	}

	// 6. Return Success Response
	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "Prestasi berhasil disimpan sebagai draft",
		Data: fiber.Map{
			"referenceId": pgData.ID,                 // ID dari Postgres
			"mongoId":     pgData.MongoAchievementID, // ID dari Mongo
			"status":      pgData.Status,
			"points":      mongoData.Points,          // Tampilkan poin yang tersimpan
		},
	})
}

// GET /api/v1/achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	param := s.parsePagination(c)
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	var filterStudent, filterAdvisor string

	// Logic Filter Berdasarkan Role (RBAC Data Level)
	if userRole == "Mahasiswa" {
		// Mahasiswa HANYA boleh melihat prestasi miliknya sendiri
		student, _ := s.userRepo.FindStudentByUserID(userID)
		if student != nil {
			filterStudent = student.ID
		}
	} else if userRole == "Dosen Wali" {
		// Dosen Wali HANYA boleh melihat prestasi mahasiswa bimbingannya (FR-006)
		lecturer, _ := s.userRepo.FindLecturerByUserID(userID)
		if lecturer != nil {
			filterAdvisor = lecturer.ID
		}
	}
	// Admin melihat semua (filter kosong)

	data, total, err := s.achRepo.FindAll(param, filterStudent, filterAdvisor)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return s.sendPaginationResponse(c, data, total, param)
}

// GET /api/v1/achievements/:id
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Validasi Kepemilikan (Optional tapi disarankan)
	// Idealnya dicek apakah user berhak melihat detail ini
	
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

// POST /api/v1/achievements/:id/submit (FR-004: Submit untuk Verifikasi)
func (s *AchievementService) RequestVerification(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Ambil Profil Mahasiswa
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	// 2. Cek Existensi Achievement
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// 3. Validasi Kepemilikan (Harus milik mahasiswa yang login)
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You do not own this achievement"})
	}

	// 4. Validasi Status (Hanya 'draft' yang bisa disubmit)
	if ref.Status != "draft" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only draft achievement can be submitted for verification"})
	}

	// 5. Update Status menjadi 'submitted'
	if err := s.achRepo.UpdateStatus(id, "submitted", "", "", 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to update status: " + err.Error()})
	}

	// 6. Create Notification (Simulasi Log)
	// TODO: Integrasi dengan Notification Service jika ada
	// log.Printf("Notifikasi dikirim ke Dosen Wali ID: %v", student.AdvisorID)

	// 7. Return Updated Status
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

// POST /api/v1/achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct{ Points int `json:"points"` }
	
	// Tambahkan error handling parsing body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

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

// DELETE /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)
	student, _ := s.userRepo.FindStudentByUserID(userID)
	
	// Cek kepemilikan sebelum hapus
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

// ... (Method stub/placeholder lain seperti Update, GetHistory, UploadAttachment biarkan saja seperti sebelumnya) ...

func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Feature Not Implemented"})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "History Not Implemented"})
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Upload Not Implemented"})
}

func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Advisee List Not Implemented"})
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