package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"uas/app/model" // Ganti 'project-uas' sesuai nama module di go.mod Anda

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseInstances menampung koneksi kedua DB agar mudah di-inject
type DatabaseInstances struct {
	Postgres *gorm.DB
	Mongo    *mongo.Database
}

var DB *DatabaseInstances // Global variable (opsional, tapi berguna untuk akses cepat)

func InitDB() *DatabaseInstances {
	pgDB := connectPostgres()
	mongoDB := connectMongo()

	DB = &DatabaseInstances{
		Postgres: pgDB,
		Mongo:    mongoDB,
	}
	
	return DB
}

// --- KONEKSI POSTGRESQL (Relational Data) ---
func connectPostgres() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
    // Tambahkan baris ini:
    DisableForeignKeyConstraintWhenMigrating: true, 
})

	if err != nil {
		log.Fatal("❌ Gagal koneksi ke PostgreSQL:", err)
	}

	log.Println("✅ Terhubung ke PostgreSQL")

	// AUTO MIGRATION: Membuat tabel otomatis berdasarkan Struct
	// Sesuai SRS Hal 4-5
	log.Println("🔄 Menjalankan Auto Migration...")
	err = db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.Student{},
		&model.Lecturer{},
		&model.AchievementReference{}, // Tabel referensi prestasi
	)

	if err != nil {
		log.Fatal("❌ Gagal migrasi database:", err)
	}

	return db
}

// --- KONEKSI MONGODB (Dynamic Data) ---
func connectMongo() *mongo.Database {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal membuat client MongoDB:", err)
	}

	// Cek koneksi dengan Ping
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Gagal ping ke MongoDB:", err)
	}

	log.Println("✅ Terhubung ke MongoDB")

	return client.Database(dbName)
}