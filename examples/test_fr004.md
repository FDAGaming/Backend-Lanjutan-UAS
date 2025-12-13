# Testing FR-004: Submit untuk Verifikasi

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan mahasiswa dan dosen wali
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
    "title": "Juara 1 Lomba Programming",
    "description": "Meraih juara 1 dalam lomba programming tingkat nasional",
    "details": {
      "competitionName": "National Programming Contest 2024",
      "competitionLevel": "national",
      "rank": 1,
      "medalType": "gold"
    },
    "tags": ["programming", "algorithm", "competition"],
    "points": 100
  }'
```

**Save the `referenceId` from response untuk step selanjutnya.**

## Step 3: Submit untuk Verifikasi

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/submit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Step 4: Verify Response

Expected response:

```json
{
  "code": 200,
  "status": "success",
  "message": "Prestasi berhasil diajukan untuk verifikasi",
  "data": {
    "id": "achievement-uuid",
    "status": "submitted"
  }
}
```

## Step 5: Check Server Logs

Server akan menampilkan notification log:

```
[NOTIFICATION] Prestasi baru untuk verifikasi:
  - Student: John Doe (123456789)
  - Achievement: Juara 1 Lomba Programming
  - Advisor ID: advisor-uuid
  - Status: submitted
  - Time: 2024-12-13 15:30:45
```

## Step 6: Verify Status Change

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response should show `"status": "submitted"` dan `submittedAt` timestamp.

## Validation Points

✅ **Precondition check**: Hanya prestasi dengan status 'draft' yang bisa disubmit
✅ **Authentication**: JWT token required
✅ **Authorization**: `achievement:create` permission required
✅ **Student validation**: User harus punya profil mahasiswa
✅ **Ownership check**: Hanya pemilik prestasi yang bisa submit
✅ **Status update**: Status berubah dari 'draft' ke 'submitted'
✅ **Timestamp update**: `submitted_at` dan `updated_at` diset
✅ **Notification**: Log notification untuk dosen wali
✅ **Response format**: Return updated status

## Error Cases to Test

### 1. Submit non-draft achievement

```bash
# Submit achievement yang sudah submitted/verified/rejected
curl -X POST http://localhost:3000/api/v1/achievements/{SUBMITTED_ACHIEVEMENT_ID}/submit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Only draft achievement can be submitted for verification"
}
```

### 2. Submit achievement milik orang lain

```bash
# Login sebagai mahasiswa A, coba submit achievement mahasiswa B
curl -X POST http://localhost:3000/api/v1/achievements/{OTHER_STUDENT_ACHIEVEMENT_ID}/submit \
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

### 3. Submit non-existent achievement

```bash
curl -X POST http://localhost:3000/api/v1/achievements/invalid-uuid/submit \
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

### 4. Submit without authentication

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/submit
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 5. Submit as non-student user

```bash
# Login sebagai admin/dosen, coba submit
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/submit \
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

## Database Verification

Setelah submit, check database:

```sql
-- PostgreSQL
SELECT id, title, status, submitted_at, updated_at 
FROM achievement_references 
WHERE id = 'your-achievement-id';

-- Should show:
-- status: 'submitted'
-- submitted_at: timestamp
-- updated_at: timestamp
```

## Next Steps

Setelah FR-004 berhasil, prestasi siap untuk:
- FR-007: Verify oleh dosen wali
- FR-008: Reject oleh dosen wali
- FR-006: View oleh dosen wali dalam daftar bimbingan