package service

import (
	"math"
	"uas/app/model"
	"uas/app/repository"
	// "strings"
	// "time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// ==========================================
// 4.2 MANAJEMEN PRESTASI (MAHASISWA)
// ==========================================

// FR-003: Submit Prestasi
// Desc: Mahasiswa mengisi data, simpan ke Mongo & Postgres dengan status 'draft'
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	// 1. Parse Input
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	// 2. Ambil User ID (Mahasiswa) dari Token
	userID := c.Locals("user_id").(string)
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	// 3. Mapping Data ke MongoDB 
	mongoData := model.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       student.ID, // UUID Student
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details, // Field dinamis
		Attachments:     req.Attachments,
		Tags:            req.Tags,
	}

	// 4. Mapping Data ke PostgreSQL 
	pgData := model.AchievementReference{
		StudentID: student.ID,
		Title:     req.Title,   // Untuk keperluan Search/Sort (Modul 6)
		Status:    "draft",     // Status Awal
	}

	// 5. Simpan (Hybrid Transaction)
	if err := s.achRepo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "Prestasi berhasil disimpan sebagai draft",
		Data:    pgData,
	})
}

// FR-004: Submit untuk Verifikasi
// Desc: Mengubah status 'draft' -> 'submitted'
func (s *AchievementService) RequestVerification(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Validasi Kepemilikan (Apakah user ini pemilik prestasi?)
	// Logic sederhana: Cek di repo atau ambil studentID user ini
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized"})
	}

	// 2. Cek Status Sekarang (Harus Draft)
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Not your achievement"})
	}
	if ref.Status != "draft" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only draft can be submitted"})
	}

	// 3. Update Status ke 'submitted'
	// Note: 'submitted_at' diupdate di repo
	if err := s.achRepo.UpdateStatus(id, "submitted", "", "", 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	// 4. (Optional) Create Notification untuk Dosen Wali (TODO: Implement Notification Service)

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi diajukan untuk verifikasi"})
}

// FR-005: Hapus Prestasi
// Desc: Hapus data jika status masih 'draft'
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Validasi Kepemilikan
	student, _ := s.userRepo.FindStudentByUserID(userID)
	
	// 2. Cek Detail & Status
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Not found"})
	}
	
	// Pastikan yang menghapus adalah pemilik
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Forbidden"})
	}

	// 3. Hapus (Repo akan validasi status 'draft')
	if err := s.achRepo.Delete(c.Context(), id); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi draft berhasil dihapus"})
}

// ==========================================
// 4.3 VERIFIKASI PRESTASI (DOSEN WALI)
// ==========================================

// FR-006: View Prestasi Mahasiswa Bimbingan
// Desc: Dosen melihat list prestasi mahasiswa bimbingannya saja
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	// 1. Ambil Data Dosen
	userID := c.Locals("user_id").(string)
	lecturer, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// 2. Parse Parameter Pagination (Modul 6)
	param := s.parsePagination(c)

	// 3. Get Data dengan Filter AdvisorID
	data, total, err := s.achRepo.FindAll(param, "", lecturer.ID) // Filter by Advisor ID
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return s.sendPaginationResponse(c, data, total, param)
}

// FR-007: Verify Prestasi
// Desc: Dosen approve prestasi, status -> 'verified', set points
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Input Body: Bisa jadi dosen ingin memberi poin spesifik
	var req struct {
		Points int `json:"points"`
	}
	c.BodyParser(&req)

	userID := c.Locals("user_id").(string) // ID User Dosen

	// 1. Update Status 
	// Status: verified, VerifiedBy: userID, Points: req.Points
	if err := s.achRepo.UpdateStatus(id, "verified", userID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi berhasil diverifikasi"})
}

// FR-008: Reject Prestasi
// Desc: Dosen tolak prestasi dengan catatan
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Parse Rejection Note [cite: 1723]
	var req struct {
		Note string `json:"note" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil || req.Note == "" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Rejection note is required"})
	}

	// 2. Update Status [cite: 1726-1727]
	// Status: rejected, VerifiedBy: userID, Note: req.Note
	if err := s.achRepo.UpdateStatus(id, "rejected", userID, req.Note, 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	// 3. (Optional) Notify Mahasiswa [cite: 1728]

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Prestasi ditolak"})
}

// ==========================================
// 4.4 MANAJEMEN SISTEM (ADMIN)
// ==========================================

// FR-010: View All Achievements
// Desc: Admin melihat SEMUA prestasi dengan filter/sorting
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	// 1. Parse Pagination (Modul 6)
	param := s.parsePagination(c)

	// 2. Logic Filter Berdasarkan Role Login (Reuse Logic)
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	var filterStudent, filterAdvisor string

	if userRole == "Mahasiswa" {
		// Mahasiswa lihat punya sendiri
		student, _ := s.userRepo.FindStudentByUserID(userID)
		filterStudent = student.ID
	} 
	// Jika Admin, filter kosong (lihat semua)

	// 3. Get Data
	data, total, err := s.achRepo.FindAll(param, filterStudent, filterAdvisor)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return s.sendPaginationResponse(c, data, total, param)
}

// ==========================================
// 4.5 REPORTING & ANALYTICS
// ==========================================

// FR-011: Achievement Statistics
// Desc: Menampilkan statistik sederhana
func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	// Logic ini bisa dikembangkan lebih lanjut dengan Aggregation Query di Repo
	// Disini kita buat contoh response statis/mockup sesuai requirement FR-011
	
	// [cite: 1750] Output yang diminta
	stats := map[string]interface{}{
		"total_per_type": map[string]int{
			"academic":    10, // Nanti ganti dengan Count DB
			"competition": 5,
		},
		"total_per_period": map[string]int{
			"2024": 20,
			"2025": 15,
		},
		"top_students": []string{"Mahasiswa A", "Mahasiswa B"}, 
		"status_distribution": map[string]int{
			"verified": 50,
			"pending":  10,
		},
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Statistics generated",
		Data:    stats,
	})
}

// ==========================================
// HELPER FUNCTIONS (Untuk Modul 6)
// ==========================================

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
		Message: "Data retrieved successfully",
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

// --- Placeholder Features ---

func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, content, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Not found"})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Data: fiber.Map{"meta": ref, "content": content}})
}

func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update feature not implemented"})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "History not implemented"})
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Upload not implemented"})
}

func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	// Logic: Get achievements by Student ID parameter
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Not implemented"})
}

func (s *AchievementService) GetStudentStatistics(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Student stats not implemented"})
}