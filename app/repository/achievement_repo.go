package repository

import (
	"context"
	"errors"
	"project-uas/app/model"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository struct {
	pgDB    *gorm.DB
	mongoColl *mongo.Collection
}

func NewAchievementRepo(pg *gorm.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		pgDB:    pg,
		mongoColl: mongoDB.Collection("achievements"),
	}
}

// Fungsi Create menyimpan ke DUA database
func (r *AchievementRepository) Create(ctx context.Context, content *model.AchievementContent, meta *model.AchievementMeta) error {
	[cite_start]// LANGKAH 1: Simpan Data Detail ke MongoDB dulu [cite: 184]
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()
	
	res, err := r.mongoColl.InsertOne(ctx, content)
	if err != nil {
		return err
	}

	// Ambil ID yang baru dibuat oleh Mongo
	mongoID := res.InsertedID.(primitive.ObjectID).Hex()

	[cite_start]// LANGKAH 2: Simpan Referensi ID tersebut ke PostgreSQL [cite: 184]
	meta.MongoAchievementID = mongoID // Link ID Mongo ke Postgres
	meta.CreatedAt = time.Now()
	meta.UpdatedAt = time.Now()
	[cite_start]meta.Status = "draft" // Status awal sesuai SRS FR-003 [cite: 185]

	if err := r.pgDB.Create(meta).Error; err != nil {
		// ROLLBACK MANUAL: Jika simpan ke Postgres gagal, hapus data di Mongo agar tidak jadi sampah
		_, _ = r.mongoColl.DeleteOne(ctx, primitive.M{"_id": res.InsertedID})
		return errors.New("failed to save reference to postgres: " + err.Error())
	}

	return nil
}