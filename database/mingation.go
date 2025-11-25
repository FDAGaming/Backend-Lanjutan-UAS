package database

import (
	"context"
	"database/sql"
	// "fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// SeedDummyData mengisi data awal untuk testing
func SeedDummyData(pg *sql.DB, mongoDB *mongo.Database) {
	log.Println("🌱 Memulai Seeding Data Dummy...")

	ctx := context.Background()

	// 1. SEED ROLES
	// ==========================================
	roles := map[string]string{
		"Admin":      "Administrator Sistem",
		"Mahasiswa":  "Pengguna Mahasiswa",
		"Dosen Wali": "Dosen Verifikator",
	}

	roleIDs := make(map[string]string)

	for name, desc := range roles {
		var id string
		// Cek apakah role sudah ada
		err := pg.QueryRow("SELECT id FROM roles WHERE name = $1", name).Scan(&id)
		if err == sql.ErrNoRows {
			// Jika belum ada, Insert
			err = pg.QueryRow(
				"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
				name, desc,
			).Scan(&id)
			if err != nil {
				log.Fatalf("❌ Gagal seed role %s: %v", name, err)
			}
			log.Printf("✅ Role Created: %s", name)
		}
		roleIDs[name] = id
	}

	// 2. SEED USERS (Admin, Dosen, Mahasiswa)
	// ==========================================
	passwordHash, _ := hashPassword("123456") // Password default

	// -- User Admin --
	// PERBAIKAN: Gunakan '_' karena kita tidak butuh ID admin untuk proses selanjutnya
	_ = seedUser(pg, "admin", "admin@univ.ac.id", passwordHash, "Super Admin", roleIDs["Admin"])
	
	// -- User Dosen --
	dosenUserID := seedUser(pg, "dosen1", "dosen1@univ.ac.id", passwordHash, "Dr. Budi Santoso", roleIDs["Dosen Wali"])
	
	// -- User Mahasiswa --
	mhsUserID := seedUser(pg, "mhs1", "mhs1@univ.ac.id", passwordHash, "Andi Pratama", roleIDs["Mahasiswa"])

	// 3. SEED PROFILES (Lecturer & Student)
	// ==========================================
	
	// -- Profile Lecturer (Dosen) --
	var lecturerID string
	err := pg.QueryRow("SELECT id FROM lecturers WHERE user_id = $1", dosenUserID).Scan(&lecturerID)
	if err == sql.ErrNoRows {
		err = pg.QueryRow(`
			INSERT INTO lecturers (user_id, lecturer_id, department) 
			VALUES ($1, $2, $3) RETURNING id`,
			dosenUserID, "19850101201001", "Teknik Informatika",
		).Scan(&lecturerID)
		if err != nil {
			log.Fatalf("❌ Gagal seed lecturer: %v", err)
		}
		log.Println("✅ Profile Lecturer Created")
	}

	// -- Profile Student (Mahasiswa) --
	// Link Advisor ke Dosen di atas
	var studentID string
	err = pg.QueryRow("SELECT id FROM students WHERE user_id = $1", mhsUserID).Scan(&studentID)
	if err == sql.ErrNoRows {
		err = pg.QueryRow(`
			INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id) 
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			mhsUserID, "20210001", "D4 Teknik Informatika", "2021", lecturerID,
		).Scan(&studentID)
		if err != nil {
			log.Fatalf("❌ Gagal seed student: %v", err)
		}
		log.Println("✅ Profile Student Created")
	}

	// 4. SEED ACHIEVEMENTS (Hybrid: Mongo + Postgres)
	// ==========================================
	
	// Cek apakah mahasiswa ini sudah punya prestasi (cek di Postgres)
	var count int
	pg.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE student_id = $1", studentID).Scan(&count)

	if count == 0 {
		log.Println("Creating Dummy Achievement...")
		
		// A. Insert ke MongoDB
		achCollection := mongoDB.Collection("achievements")
		
		mongoDoc := bson.M{
			"studentId":       studentID, // UUID string dari Postgres
			"achievementType": "competition",
			"title":           "Juara 1 Hackathon Nasional",
			"description":     "Menang lomba coding di Jakarta",
			"createdAt":       time.Now(),
			"updatedAt":       time.Now(),
			"points":          100,
			"tags":            []string{"coding", "java", "winner"},
			"details": bson.M{
				"competitionName":  "Gemastik 2025",
				"competitionLevel": "national",
				"rank":             1,
				"location":         "Jakarta",
				"eventDate":        time.Now(),
			},
			"attachments": []bson.M{
				{
					"fileName":   "sertifikat.pdf",
					"fileUrl":    "http://localhost:3000/uploads/dummy.pdf",
					"fileType":   "application/pdf",
					"uploadedAt": time.Now(),
				},
			},
		}

		res, err := achCollection.InsertOne(ctx, mongoDoc)
		if err != nil {
			log.Fatalf("❌ Gagal seed Mongo achievement: %v", err)
		}
		
		mongoID := res.InsertedID.(primitive.ObjectID).Hex()
		log.Printf("✅ Mongo Achievement Inserted ID: %s", mongoID)

		// B. Insert ke Postgres (Reference)
		_, err = pg.Exec(`
			INSERT INTO achievement_references 
			(student_id, mongo_achievement_id, title, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			studentID, mongoID, "Juara 1 Hackathon Nasional", "submitted", time.Now(), time.Now(),
		)
		if err != nil {
			log.Fatalf("❌ Gagal seed Postgres reference: %v", err)
		}
		log.Println("✅ Postgres Reference Created")
	}

	log.Println("🎉 Seeding Selesai!")
}

// Helper: Seed User
func seedUser(db *sql.DB, username, email, passwordHash, fullName, roleID string) string {
	var id string
	// Cek exist
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err == sql.ErrNoRows {
		// Insert
		err = db.QueryRow(`
			INSERT INTO users (username, email, password_hash, full_name, role_id)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			username, email, passwordHash, fullName, roleID,
		).Scan(&id)
		if err != nil {
			log.Fatalf("❌ Gagal seed user %s: %v", username, err)
		}
		log.Printf("✅ User Created: %s (%s)", username, email)
	}
	return id
}

// Helper: Hash Password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}