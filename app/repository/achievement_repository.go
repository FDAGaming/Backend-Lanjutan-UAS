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
		mongoColl: mongoDB.Collection("achievements"),
	}
}

// --- CREATE (HYBRID TRANSACTION) ---
func (r *AchievementRepository) Create(ctx context.Context, content *model.Achievement, ref *model.AchievementReference) error {
	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now
	ref.CreatedAt = now
	ref.UpdatedAt = now

	// 1. Insert ke MongoDB
	res, err := r.mongoColl.InsertOne(ctx, content)
	if err != nil {
		return err
	}

	// Ambil ID dari Mongo
	oid, _ := res.InsertedID.(primitive.ObjectID)
	ref.MongoAchievementID = oid.Hex()

	// 2. Insert ke PostgreSQL
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

// --- FIND ALL (PAGINATION, SORT, SEARCH) ---
func (r *AchievementRepository) FindAll(param model.PaginationParam, studentID string, advisorID string) ([]model.AchievementReference, int64, error) {
	var achievements []model.AchievementReference
	var total int64

	// Base Query
	baseQuery := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.title, ar.status, 
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
			u.full_name, s.student_id
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id`

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
		argId++ // Increment argId for the next potential argument
	}

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

	// Note: We need to pass 'args' carefully. If search was used, args has the search term twice? 
	// No, in my implementation above I used the SAME arg index ($N) for both title and status check.
	// But 'args' slice only needs the value ONCE if I use the same index.
	// Wait, standard sql package placeholder $1, $2.. implies distinct arguments usually.
	// Actually, for Postgres $1 can be reused. So if I used the same index in the format string, I only append once.
	// Correct logic:
	// conditions: ... LIKE $3 OR ... LIKE $3
	// args: [..., searchTerm] -> correct.

	if err := r.pgDB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2. Sorting
	orderBy := "ar.created_at" // Default
	if param.SortBy != "" {
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
		ar.Student = &model.Student{User: &model.User{}} 

		var subAt, verAt sql.NullTime
		var verBy sql.NullString
		var rejNote sql.NullString

		err := rows.Scan(
			&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Title, &ar.Status,
			&subAt, &verAt, &verBy, &rejNote, &ar.CreatedAt, &ar.UpdatedAt,
			&ar.Student.User.FullName, &ar.Student.StudentID, 
		)
		if err != nil {
			return nil, 0, err
		}

		if subAt.Valid { ar.SubmittedAt = &subAt.Time }
		if verAt.Valid { ar.VerifiedAt = &verAt.Time }
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
			s.student_id, u.full_name,
			ver_u.full_name
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

	if subAt.Valid { ref.SubmittedAt = &subAt.Time }
	if verAt.Valid { ref.VerifiedAt = &verAt.Time }
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
		return &ref, nil, errors.New("detail data not found in mongo")
	}

	return &ref, &content, nil
}

// --- UPDATE STATUS (VERIFIKASI DOSEN & UPDATE POIN) ---
func (r *AchievementRepository) UpdateStatus(id string, status string, verifiedBy string, note string, points int) error {
	// 1. Update di PostgreSQL (Status, VerifiedBy, RejectionNote)
	query := `
		UPDATE achievement_references 
		SET status = $1, updated_at = $2, verified_by = $3, verified_at = $4, rejection_note = $5
		WHERE id = $6
		RETURNING mongo_achievement_id`
	
	now := time.Now()
	
	var verBy interface{} = nil
	if verifiedBy != "" {
		verBy = verifiedBy
	}

	var mongoID string
	// QueryRow digunakan karena kita mengharapkan RETURNING mongo_achievement_id
	err := r.pgDB.QueryRow(query, status, now, verBy, now, note, id).Scan(&mongoID)
	if err != nil {
		return err
	}

	// 2. Update di MongoDB (Hanya jika Verified, kita simpan Poinnya)
	if status == "verified" && mongoID != "" {
		objID, err := primitive.ObjectIDFromHex(mongoID)
		if err == nil {
			// Update field "points" di dokumen MongoDB
			_, err = r.mongoColl.UpdateOne(
				context.Background(), 
				bson.M{"_id": objID}, 
				bson.M{"$set": bson.M{"points": points}},
			)
			if err != nil {
				return errors.New("failed to update points in mongo: " + err.Error())
			}
		}
	}

	return nil
}

// --- DELETE ---
func (r *AchievementRepository) Delete(ctx context.Context, id string) error {
	var mongoID string
	var status string
	
	err := r.pgDB.QueryRow("SELECT mongo_achievement_id, status FROM achievement_references WHERE id = $1", id).Scan(&mongoID, &status)
	if err != nil {
		return err
	}

	if status != "draft" {
		return errors.New("cannot delete submitted or verified achievement")
	}

	_, err = r.pgDB.Exec("DELETE FROM achievement_references WHERE id = $1", id)
	if err != nil {
		return err
	}

	if mongoID != "" {
		objID, _ := primitive.ObjectIDFromHex(mongoID)
		_, err := r.mongoColl.DeleteOne(ctx, bson.M{"_id": objID})
		return err
	}
	
	return nil
}