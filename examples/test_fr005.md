# Testing FR-005: Hapus Prestasi (Soft Delete)

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan mahasiswa
3. Achievement dengan status 'draft' sudah ada (dari FR-003)
4. Login sebagai mahasiswa untuk mendapat token

## Step 1: Login sebagai Mahasiswa

```bash
# Login untuk mendapat token
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@university.ac.id",
    "password": "student123"
  }'
```

## Step 2: Create Draft Achievement (jika belum ada)

```bash
curl -X POST http://localhost:3000/api/v1/achievements \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "achievementType": "competition",
    "title": "Test Achievement for Delete",
    "description": "Achievement yang akan dihapus untuk testing",
    "details": {
      "competitionName": "Test Competition",
      "competitionLevel": "local",
      "rank": 1
    },
    "tags": ["test"],
    "points": 50
  }'
```

**Save the `referenceId` from response untuk step selanjutnya.**

## Step 3: Verify Achievement Exists

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Should return achievement with `"status": "draft"`.

## Step 4: Delete Achievement (Soft Delete)

```bash
curl -X DELETE http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Step 5: Verify Response

Expected response:

```json
{
  "code": 200,
  "status": "success",
  "message": "Prestasi berhasil dihapus",
  "data": {
    "id": "achievement-uuid",
    "status": "deleted"
  }
}
```

## Step 6: Check Server Logs

Server akan menampilkan soft delete log:

```
[SOFT DELETE] Achievement deleted:
  - Student: John Doe (123456789)
  - Achievement: Test Achievement for Delete
  - Time: 2024-12-13 15:45:30
```

## Step 7: Verify Soft Delete

### 7.1 Achievement tidak muncul di list

```bash
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Achievement yang dihapus tidak akan muncul dalam list.

### 7.2 Achievement tidak bisa diakses via detail

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{DELETED_ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Achievement not found"
}
```

## Step 8: Database Verification

### PostgreSQL - Status berubah ke 'deleted'

```sql
SELECT id, title, status, updated_at 
FROM achievement_references 
WHERE id = 'your-achievement-id';

-- Should show:
-- status: 'deleted'
-- updated_at: timestamp when deleted
```

### MongoDB - Soft delete flags ditambahkan

```javascript
db.achievements.findOne({"_id": ObjectId("your-mongo-id")})

// Should show:
// {
//   ...existing fields...,
//   "isDeleted": true,
//   "deletedAt": ISODate("2024-12-13T15:45:30.000Z"),
//   "updatedAt": ISODate("2024-12-13T15:45:30.000Z")
// }
```

## Validation Points

✅ **Precondition check**: Hanya prestasi dengan status 'draft' yang bisa dihapus
✅ **Authentication**: JWT token required
✅ **Authorization**: `achievement:delete` permission required
✅ **Student validation**: User harus punya profil mahasiswa
✅ **Ownership check**: Hanya pemilik prestasi yang bisa hapus
✅ **Soft delete MongoDB**: Data tidak dihapus permanen, ditambah flag `isDeleted`
✅ **Update reference PostgreSQL**: Status berubah ke 'deleted'
✅ **Exclude from queries**: Data yang soft deleted tidak muncul di list/detail
✅ **Return success message**: Response dengan status deleted

## Error Cases to Test

### 1. Delete non-draft achievement

```bash
# Submit achievement dulu, lalu coba hapus
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/submit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Kemudian coba hapus
curl -X DELETE http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Only draft achievements can be deleted"
}
```

### 2. Delete achievement milik orang lain

```bash
# Login sebagai mahasiswa A, coba hapus achievement mahasiswa B
curl -X DELETE http://localhost:3000/api/v1/achievements/{OTHER_STUDENT_ACHIEVEMENT_ID} \
  -H "Authorization: Bearer STUDENT_A_TOKEN"
```

Expected response:
```json
{
  "code": 403,
  "status": "error",
  "message": "Unauthorized: You do not own this achievement"
}
```

### 3. Delete non-existent achievement

```bash
curl -X DELETE http://localhost:3000/api/v1/achievements/invalid-uuid \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Achievement not found"
}
```

### 4. Delete without authentication

```bash
curl -X DELETE http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 5. Delete as non-student user

```bash
# Login sebagai admin/dosen, coba hapus
curl -X DELETE http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Student profile not found"
}
```

### 6. Delete already deleted achievement

```bash
# Hapus achievement yang sudah dihapus sebelumnya
curl -X DELETE http://localhost:3000/api/v1/achievements/{DELETED_ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Achievement not found"
}
```

## Recovery Testing (Optional)

Untuk testing recovery dari soft delete, bisa buat endpoint admin untuk restore:

```sql
-- Manual restore via database
UPDATE achievement_references 
SET status = 'draft', updated_at = NOW() 
WHERE id = 'your-achievement-id';
```

```javascript
// MongoDB restore
db.achievements.updateOne(
  {"_id": ObjectId("your-mongo-id")},
  {
    "$unset": {"isDeleted": "", "deletedAt": ""},
    "$set": {"updatedAt": new Date()}
  }
)
```

## Benefits of Soft Delete

1. **Data Recovery**: Data bisa dipulihkan jika dihapus tidak sengaja
2. **Audit Trail**: Tetap ada jejak data yang pernah ada
3. **Referential Integrity**: Tidak merusak relasi dengan data lain
4. **Analytics**: Data historis tetap bisa dianalisis
5. **Compliance**: Memenuhi requirement audit dan compliance