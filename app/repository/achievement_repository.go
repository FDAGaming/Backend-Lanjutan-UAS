package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"uas/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	pgDB      *sql.DB
	mongoColl *mongo.Collection
}

func NewAchievementRepository(pg *sql.DB, mongoDB *mongo.Database) *AchievementRepository {
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
	query := `
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, title, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err = r.pgDB.QueryRow(
		query,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Title,
		ref.Status,
		ref.CreatedAt,
		ref.UpdatedAt,
	).Scan(&ref.ID)

	if err != nil {
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

	// Base Query
	// Kita join ke Student & User untuk mendapatkan info dasar (misal nama mahasiswa)
	baseQuery := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.title, ar.status, 
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
			u.full_name, s.student_id -- Info tambahan untuk frontend (Nama & NIM)
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id`

	// Dynamic Filtering
	var conditions []string
	var args []interface{}
	argId := 1

	// Filter by Student (RBAC)
	if studentID != "" {
		conditions = append(conditions, fmt.Sprintf("ar.student_id = $%d", argId))
		args = append(args, studentID)
		argId++
	}

	// Filter by Advisor (RBAC)
	if advisorID != "" {
		conditions = append(conditions, fmt.Sprintf("s.advisor_id = $%d", argId))
		args = append(args, advisorID)
		argId++
	}

	// Search Logic (Title OR Status)
	if param.Search != "" {
		searchLike := "%" + strings.ToLower(param.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(ar.title) LIKE $%d OR LOWER(ar.status) LIKE $%d)", argId, argId))
		args = append(args, searchLike)
		// Karena kita pakai placeholder yang sama ($N) dua kali di query string, lib/pq mungkin butuh trik,
		// tapi standar sql driver biasanya mapping by index.
		// Untuk aman di PostgreSQL native driver ($1, $2), kita perlu append argumen yang sama jika indexnya beda.
		// Namun di sini kita pakai trik index yang sama ($3... $3) jika driver mendukung, 
		// atau kita anggap argumennya cuma satu searchLike.
		// Koreksi: Untuk Postgres ($1, $2), kita tidak bisa reuse index untuk value berbeda.
		// Tapi karena valuenya SAMA, kita reuse indexnya.
	}

	// Construct WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Count Total (Untuk Pagination)
	countQuery := `
		SELECT COUNT(*) 
		FROM achievement_references ar 
		JOIN students s ON ar.student_id = s.id 
	` + whereClause

	if err := r.pgDB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2. Sorting
	orderBy := "ar.created_at" // Default
	if param.SortBy != "" {
		// Whitelist column names to prevent SQL Injection
		validSorts := map[string]string{
			"title":      "ar.title",
			"status":     "ar.status",
			"created_at": "ar.created_at",
		}
		if val, ok := validSorts[param.SortBy]; ok {
			orderBy = val
		}
	}

	orderDir := "DESC"
	if strings.ToUpper(param.Order) == "ASC" {
		orderDir = "ASC"
	}

	// 3. Pagination
	limit := param.Limit
	offset := (param.Page - 1) * param.Limit

	// Final Query
	finalQuery := fmt.Sprintf(
		"%s %s ORDER BY %s %s LIMIT %d OFFSET %d",
		baseQuery, whereClause, orderBy, orderDir, limit, offset,
	)

	rows, err := r.pgDB.Query(finalQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var ar model.AchievementReference
		ar.Student = &model.Student{User: &model.User{}} // Init nested struct

		// Variable helper untuk scanning field nullable
		var subAt, verAt sql.NullTime
		var verBy sql.NullString
		var rejNote sql.NullString

		err := rows.Scan(
			&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Title, &ar.Status,
			&subAt, &verAt, &verBy, &rejNote, &ar.CreatedAt, &ar.UpdatedAt,
			&ar.Student.User.FullName, &ar.Student.StudentID, // Nama & NIM
		)
		if err != nil {
			return nil, 0, err
		}

		// Map Nullable Fields
		if subAt.Valid {
			ar.SubmittedAt = &subAt.Time
		}
		if verAt.Valid {
			ar.VerifiedAt = &verAt.Time
		}
		if verBy.Valid {
			str := verBy.String
			ar.VerifiedBy = &str
		}
		ar.RejectionNote = rejNote.String

		achievements = append(achievements, ar)
	}

	return achievements, total, nil
}

// --- FIND DETAIL (HYBRID FETCH) ---

func (r *AchievementRepository) FindDetail(ctx context.Context, id string) (*model.AchievementReference, *model.Achievement, error) {
	// 1. Ambil data Metadata dari Postgres
	query := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.title, ar.status, 
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
			s.student_id, u.full_name, -- Info Student
			ver_u.full_name -- Info Verifier
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		LEFT JOIN users ver_u ON ar.verified_by = ver_u.id
		WHERE ar.id = $1`

	var ref model.AchievementReference
	ref.Student = &model.Student{User: &model.User{}}
	ref.Verifier = &model.User{}

	var subAt, verAt sql.NullTime
	var verBy sql.NullString
	var rejNote sql.NullString
	var verifierName sql.NullString

	err := r.pgDB.QueryRow(query, id).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Title, &ref.Status,
		&subAt, &verAt, &verBy, &rejNote, &ref.CreatedAt, &ref.UpdatedAt,
		&ref.Student.StudentID, &ref.Student.User.FullName,
		&verifierName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("achievement reference not found")
		}
		return nil, nil, err
	}

	// Map Nullable
	if subAt.Valid {
		ref.SubmittedAt = &subAt.Time
	}
	if verAt.Valid {
		ref.VerifiedAt = &verAt.Time
	}
	if verBy.Valid {
		str := verBy.String
		ref.VerifiedBy = &str
		ref.Verifier.FullName = verifierName.String
	}
	ref.RejectionNote = rejNote.String

	// 2. Ambil data Detail dari MongoDB
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
	// Build Query
	query := `
		UPDATE achievement_references 
		SET status = $1, updated_at = $2, verified_by = $3, verified_at = $4, rejection_note = $5
		WHERE id = $6`
	
	now := time.Now()
	
	// Handle nullable verifiedBy
	var verBy interface{} = nil
	if verifiedBy != "" {
		verBy = verifiedBy
	}

	// Note: Field 'points' tidak ada di Struct AchievementReference Postgres dalam kode model terakhir kita,
	// Jika SRS meminta simpan poin di SQL, pastikan kolom 'points' ada di tabel achievement_references.
	// Jika belum ada, query ini mungkin error. Asumsi kolom sudah dibuat di migration manual config/database.go
	// Jika kolom points belum ada, hapus baris update points. 
	// (Berdasarkan model AchievementReference di chat sebelumnya, field Points belum ditambahkan ke Struct SQL, hanya di Mongo).
	// Namun, jika mau update points, idealnya update juga di Mongo.
	
	// Kita update SQL standard dulu
	_, err := r.pgDB.Exec(query, status, now, verBy, now, note, id)
	return err
}

// --- DELETE (SOFT DELETE / HARD DELETE) ---
// Sesuai FR-005, mahasiswa bisa hapus draft
func (r *AchievementRepository) Delete(ctx context.Context, id string) error {
	// 1. Cari dulu datanya untuk dapatkan MongoID & Cek Status
	var mongoID string
	var status string
	
	err := r.pgDB.QueryRow("SELECT mongo_achievement_id, status FROM achievement_references WHERE id = $1", id).Scan(&mongoID, &status)
	if err != nil {
		return err
	}

	// Cek status: Hanya boleh hapus jika Draft
	if status != "draft" {
		return errors.New("cannot delete submitted or verified achievement")
	}

	// 2. Hapus dari Postgres
	_, err = r.pgDB.Exec("DELETE FROM achievement_references WHERE id = $1", id)
	if err != nil {
		return err
	}

	// 3. Hapus dari Mongo (Cleanup)
	if mongoID != "" {
		objID, _ := primitive.ObjectIDFromHex(mongoID)
		_, err := r.mongoColl.DeleteOne(ctx, bson.M{"_id": objID})
		return err
	}
	
	return nil
}