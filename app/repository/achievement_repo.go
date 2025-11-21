package repository

import (
	"context"
	"errors"
	"uas/app/model"
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

func NewAchievementRepo(pg *gorm.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		pgDB:      pg,
		mongoColl: mongoDB.Collection("achievements"),
	}
}

// Create: Insert Mongo -> Insert Postgres
func (r *AchievementRepository) Create(ctx context.Context, content *model.AchievementContent, ref *model.AchievementReference) error {
	// 1. Insert ke MongoDB
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()
	res, err := r.mongoColl.InsertOne(ctx, content)
	if err != nil {
		return err
	}

	// 2. Ambil ID Mongo dan set ke Referensi Postgres
	ref.MongoAchievementID = res.InsertedID.(primitive.ObjectID).Hex()
	ref.CreatedAt = time.Now()
	ref.UpdatedAt = time.Now()
	
	// 3. Insert ke Postgres
	if err := r.pgDB.Create(ref).Error; err != nil {
		// KOMPENSASI: Hapus data di Mongo jika Postgres gagal
		_, _ = r.mongoColl.DeleteOne(ctx, bson.M{"_id": res.InsertedID})
		return errors.New("failed to create reference: " + err.Error())
	}

	return nil
}

// FindAllByAdvisor: Ambil list prestasi mahasiswa bimbingan (Hanya dari Postgres agar cepat)
func (r *AchievementRepository) FindAllByStudentIDs(studentIDs []string) ([]model.AchievementReference, error) {
	var refs []model.AchievementReference
	err := r.pgDB.Where("student_id IN ?", studentIDs).Order("created_at desc").Find(&refs).Error
	return refs, err
}

// FindDetail: Ambil referensi Postgres + Detail Mongo
func (r *AchievementRepository) FindDetail(ctx context.Context, id string) (*model.AchievementReference, *model.AchievementContent, error) {
	// 1. Ambil Postgres Data
	var ref model.AchievementReference
	if err := r.pgDB.First(&ref, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}

	// 2. Ambil Mongo Data berdasarkan ID yang disimpan di Postgres
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	var content model.AchievementContent
	
	err := r.mongoColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
	if err != nil {
		return &ref, nil, errors.New("detail data not found in mongo")
	}

	return &ref, &content, nil
}

// UpdateStatus: Verifikasi Dosen
func (r *AchievementRepository) UpdateStatus(id string, status string, verifiedBy string, note string) error {
	updates := map[string]interface{}{
		"status":      status,
		"updated_at":  time.Now(),
		"verified_by": verifiedBy,
		"verified_at": time.Now(),
	}
	if note != "" {
		updates["rejection_note"] = note
	}

	return r.pgDB.Model(&model.AchievementReference{}).Where("id = ?", id).Updates(updates).Error
}