package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // Driver PostgreSQL standar
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInstances menampung koneksi kedua DB agar mudah di-inject
type DatabaseInstances struct {
	Postgres *sql.DB // Menggunakan *sql.DB standar
	Mongo    *mongo.Database
}

var DB *DatabaseInstances

func InitDB() *DatabaseInstances {
	pgDB := connectPostgres()
	mongoDB := connectMongo()

	DB = &DatabaseInstances{
		Postgres: pgDB,
		Mongo:    mongoDB,
	}

	return DB
}

// --- KONEKSI POSTGRESQL (Relational Data - Native SQL) ---
func connectPostgres() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Buka koneksi menggunakan driver 'postgres'
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Gagal membuka driver PostgreSQL:", err)
	}

	// Cek koneksi (Ping)
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Gagal koneksi ke PostgreSQL:", err)
	}

	log.Println("✅ Terhubung ke PostgreSQL (Native SQL)")

	// MANUAL MIGRATION
	runManualMigration(db)

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

// Fungsi untuk menjalankan DDL (Data Definition Language)
func runManualMigration(db *sql.DB) {
	log.Println("🔄 Menjalankan Manual Migration (Reset Schema)...")

	// ⚠️ PERHATIAN: Baris DROP ini akan menghapus data lama untuk memperbaiki struktur tabel.
	// Jika aplikasi sudah production, JANGAN gunakan DROP.
	dropQueries := []string{
		`DROP TABLE IF EXISTS achievement_references CASCADE;`,
		`DROP TABLE IF EXISTS students CASCADE;`,
		`DROP TABLE IF EXISTS lecturers CASCADE;`,
		`DROP TABLE IF EXISTS users CASCADE;`,
		`DROP TABLE IF EXISTS role_permissions CASCADE;`,
		`DROP TABLE IF EXISTS permissions CASCADE;`,
		`DROP TABLE IF EXISTS roles CASCADE;`,
	}

	for _, query := range dropQueries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("⚠️ Gagal reset tabel: %s\nError: %v", query, err)
		}
	}

	queries := []string{
		// 1. Enable UUID extension (jika belum aktif)
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,

		// 2. Tabel roles
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// 3. Tabel permissions
		`CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			resource VARCHAR(50) NOT NULL,
			action VARCHAR(50) NOT NULL,
			description TEXT
		);`,

		// 4. Tabel role_permissions (Many-to-Many)
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		);`,

		// 5. Tabel users
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			role_id UUID NOT NULL REFERENCES roles(id),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// 6. Tabel lecturers (Dibuat sebelum students karena direferensi oleh students)
		`CREATE TABLE IF NOT EXISTS lecturers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL UNIQUE REFERENCES users(id),
			lecturer_id VARCHAR(20) UNIQUE NOT NULL, -- NIP
			department VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// 7. Tabel students
		// Perbaikan: student_id harus VARCHAR agar bisa menyimpan NIM "20210001"
		`CREATE TABLE IF NOT EXISTS students (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL UNIQUE REFERENCES users(id),
			student_id VARCHAR(20) UNIQUE NOT NULL, -- NIM (String)
			program_study VARCHAR(100),
			academic_year VARCHAR(10),
			advisor_id UUID REFERENCES lecturers(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// 8. Tabel achievement_references
		`CREATE TABLE IF NOT EXISTS achievement_references (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			student_id UUID NOT NULL REFERENCES students(id),
			mongo_achievement_id VARCHAR(24) NOT NULL,
			title VARCHAR(255) NOT NULL,
			status VARCHAR(20) DEFAULT 'draft',
			submitted_at TIMESTAMP,
			verified_at TIMESTAMP,
			verified_by UUID REFERENCES users(id),
			rejection_note TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("⚠️ Gagal menjalankan migrasi query: %s\nError: %v", query, err)
		}
	}
	log.Println("✅ Manual Migration Selesai")
}