package repository

import (
	"context"
	"errors"
	"uas/app/model"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository struct {
	pgDB      *gorm.DB
	mongoColl *mongo.Collection
}

func NewAchievementRepository(pg *gorm.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		pgDB:      pg,
		mongoColl: mongoDB.Collection("achievements"), // Sesuai SRS 3.2.1
	}
}

// --- CREATE (HYBRID TRANSACTION) ---

func (r *AchievementRepository) Create(ctx context.Context, content *model.Achievement, ref *model.AchievementReference) error {
	// 1. Set Timestamp
	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now
	ref.CreatedAt = now
	ref.UpdatedAt = now

	// 2. Insert ke MongoDB
	res, err := r.mongoColl.InsertOne(ctx, content)
	if err != nil {
		return err
	}

	// 3. Ambil ID dari Mongo, masukkan ke field Reference Postgres
	oid, _ := res.InsertedID.(primitive.ObjectID)
	ref.MongoAchievementID = oid.Hex()

	// 4. Insert ke PostgreSQL
	if err := r.pgDB.Create(ref).Error; err != nil {
		// KOMPENSASI (ROLLBACK MANUAL):
		// Jika simpan ke Postgres gagal, hapus data sampah di Mongo
		_, _ = r.mongoColl.DeleteOne(ctx, bson.M{"_id": oid})
		return errors.New("failed to save reference to postgres: " + err.Error())
	}

	return nil
}

// --- FIND ALL (PAGINATION, SORT, SEARCH - MODUL 6) ---

func (r *AchievementRepository) FindAll(param model.PaginationParam, studentID string, advisorID string) ([]model.AchievementReference, int64, error) {
	var achievements []model.AchievementReference
	var total int64

	// 1. Start Query
	query := r.pgDB.Model(&model.AchievementReference{}).Preload("Student.User")

	// 2. Filter Logic (RBAC Data Level)
	if studentID != "" {
		// Jika Mahasiswa, hanya lihat punya sendiri
		query = query.Where("student_id = ?", studentID) // student_id disini adalah UUID (referensi ke tabel student)
	}
	if advisorID != "" {
		// Jika Dosen Wali, hanya lihat mahasiswa bimbingannya
		// Join ke tabel student untuk cek advisor_id
		query = query.Joins("JOIN students ON students.id = achievement_references.student_id").
			Where("students.advisor_id = ?", advisorID)
	}

	// 3. Search (Search By Title OR Status) - Case Insensitive
	if param.Search != "" {
		searchLower := "%" + strings.ToLower(param.Search) + "%"
		// Postgres ILIKE or LOWER()
		query = query.Where("LOWER(achievement_references.title) LIKE ? OR LOWER(achievement_references.status) LIKE ?", searchLower, searchLower)
	}

	// 4. Count Total (Sebelum Limit/Offset)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. Sorting
	orderBy := "created_at" // Default
	if param.SortBy != "" {
		orderBy = param.SortBy
	}
	
	orderDir := "DESC" // Default
	if strings.ToUpper(param.Order) == "ASC" {
		orderDir = "ASC"
	}
	query = query.Order(orderBy + " " + orderDir)

	// 6. Pagination (Limit & Offset)
	offset := (param.Page - 1) * param.Limit
	query = query.Limit(param.Limit).Offset(offset)

	// 7. Execute
	err := query.Find(&achievements).Error
	return achievements, total, err
}

// --- FIND DETAIL (HYBRID FETCH) ---

func (r *AchievementRepository) FindDetail(ctx context.Context, id string) (*model.AchievementReference, *model.Achievement, error) {
	// 1. Ambil data Metadata dari Postgres
	var ref model.AchievementReference
	if err := r.pgDB.Preload("Student.User").Preload("Verifier").First(&ref, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}

	// 2. Ambil data Detail dari MongoDB menggunakan ID yang tersimpan di Postgres
	var content model.Achievement
	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return &ref, nil, errors.New("invalid mongo id format")
	}

	err = r.mongoColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
	if err != nil {
		// Jika data mongo hilang (inkonsistensi), return meta saja dengan error note
		return &ref, nil, errors.New("detail data not found in mongo")
	}

	return &ref, &content, nil
}

// --- UPDATE STATUS (VERIFIKASI DOSEN) ---

func (r *AchievementRepository) UpdateStatus(id string, status string, verifiedBy string, note string, points int) error {
	updates := map[string]interface{}{
		"status":      status,
		"updated_at":  time.Now(),
		"verified_by": verifiedBy,
		"verified_at": time.Now(),
	}
	
	if note != "" {
		updates["rejection_note"] = note
	}
	
	// Jika verified, tambahkan poin
	if status == "verified" && points > 0 {
		updates["points"] = points
	}

	return r.pgDB.Model(&model.AchievementReference{}).Where("id = ?", id).Updates(updates).Error
}

// --- DELETE (SOFT DELETE / HARD DELETE) ---
// Sesuai FR-005, mahasiswa bisa hapus draft
func (r *AchievementRepository) Delete(ctx context.Context, id string) error {
	// 1. Cari dulu datanya untuk dapatkan MongoID
	var ref model.AchievementReference
	if err := r.pgDB.First(&ref, "id = ?", id).Error; err != nil {
		return err
	}

	// Cek status: Hanya boleh hapus jika Draft
	if ref.Status != "draft" {
		return errors.New("cannot delete submitted or verified achievement")
	}

	// 2. Hapus dari Postgres
	if err := r.pgDB.Delete(&ref).Error; err != nil {
		return err
	}

	// 3. Hapus dari Mongo (Cleanup)
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	_, err := r.mongoColl.DeleteOne(ctx, bson.M{"_id": objID})
	
	return err
}