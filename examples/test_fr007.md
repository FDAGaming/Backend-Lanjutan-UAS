# Testing FR-007: Verify Prestasi

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan dosen wali dan mahasiswa bimbingan
3. Achievement dengan status 'submitted' sudah ada (dari FR-004)
4. Relasi advisor_id sudah diset di tabel students

## Step 1: Setup Data (if needed)

### Create submitted achievement

```bash
# Login sebagai mahasiswa
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@university.ac.id",
    "password": "student123"
  }'

# Create draft achievement
curl -X POST http://localhost:3000/api/v1/achievements \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer STUDENT_TOKEN" \
  -d '{
    "achievementType": "competition",
    "title": "Juara 1 Lomba Programming Nasional",
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

# Submit for verification
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/submit \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

## Step 2: Login sebagai Dosen Wali

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dosen@university.ac.id",
    "password": "dosen123"
  }'
```

## Step 3: Dosen Review Prestasi Detail

```bash
# Get achievement detail untuk review
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

Expected response untuk review:
```json
{
  "code": 200,
  "status": "success",
  "message": "Detail retrieved",
  "data": {
    "meta": {
      "id": "achievement-ref-uuid",
      "title": "Juara 1 Lomba Programming Nasional",
      "status": "submitted",
      "submittedAt": "2024-12-13T10:30:00Z",
      "student": {
        "studentId": "123456789",
        "user": {
          "fullName": "John Doe"
        }
      }
    },
    "content": {
      "achievementType": "competition",
      "title": "Juara 1 Lomba Programming Nasional",
      "description": "Meraih juara 1 dalam lomba programming tingkat nasional",
      "details": {
        "competitionName": "National Programming Contest 2024",
        "competitionLevel": "national",
        "rank": 1,
        "medalType": "gold"
      },
      "tags": ["programming", "algorithm", "competition"],
      "points": 100
    }
  }
}
```

## Step 4: Verify Prestasi

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{
    "points": 150
  }'
```

## Step 5: Verify Response

Expected response:

```json
{
  "code": 200,
  "status": "success",
  "message": "Prestasi berhasil diverifikasi",
  "data": {
    "id": "achievement-uuid",
    "status": "verified",
    "points": 150,
    "verifiedBy": "dosen-user-uuid",
    "verifiedAt": "2024-12-13T15:45:30Z",
    "achievement": {
      "title": "Juara 1 Lomba Programming Nasional",
      "type": "competition",
      "student": "John Doe",
      "studentId": "123456789"
    }
  }
}
```

## Step 6: Check Server Logs

Server akan menampilkan verification log:

```
[VERIFICATION] Achievement verified:
  - Student: John Doe (123456789)
  - Achievement: Juara 1 Lomba Programming Nasional
  - Points awarded: 150
  - Verified by: Dr. Jane Smith
  - Time: 2024-12-13 15:45:30
```

## Step 7: Verify Status Change

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

Response should show:
- `"status": "verified"`
- `"verifiedAt": timestamp`
- `"verifiedBy": dosen-user-uuid`

## Validation Points

✅ **Precondition check**: Hanya prestasi dengan status 'submitted' yang bisa diverifikasi
✅ **Authentication**: JWT token required
✅ **Authorization**: `achievement:verify` permission required
✅ **Lecturer validation**: User harus punya profil dosen wali
✅ **Ownership check**: Hanya dosen wali yang bisa verify prestasi mahasiswa bimbingannya
✅ **Achievement review**: Dosen bisa review detail prestasi sebelum verify
✅ **Status update**: Status berubah dari 'submitted' ke 'verified'
✅ **Metadata update**: Set verified_by dan verified_at
✅ **Points assignment**: Update points di MongoDB
✅ **Return updated status**: Response dengan data lengkap verification

## Error Cases to Test

### 1. Verify non-submitted achievement

```bash
# Coba verify achievement dengan status 'draft'
curl -X POST http://localhost:3000/api/v1/achievements/{DRAFT_ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Only submitted achievements can be verified"
}
```

### 2. Verify achievement bukan mahasiswa bimbingan

```bash
# Login sebagai dosen A, coba verify prestasi mahasiswa bimbingan dosen B
curl -X POST http://localhost:3000/api/v1/achievements/{OTHER_ADVISEE_ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_A_TOKEN" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 403,
  "status": "error",
  "message": "Unauthorized: You can only verify achievements of your advisees"
}
```

### 3. Verify with invalid points

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"points": 0}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Points must be greater than 0"
}
```

### 4. Verify non-existent achievement

```bash
curl -X POST http://localhost:3000/api/v1/achievements/invalid-uuid/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Achievement not found"
}
```

### 5. Verify without authentication

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 6. Verify as non-lecturer

```bash
# Login sebagai mahasiswa, coba verify
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer STUDENT_TOKEN" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Lecturer profile not found"
}
```

### 7. Verify already verified achievement

```bash
# Coba verify achievement yang sudah verified
curl -X POST http://localhost:3000/api/v1/achievements/{VERIFIED_ACHIEVEMENT_ID}/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"points": 100}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Only submitted achievements can be verified"
}
```

## Database Verification

### PostgreSQL - Status dan metadata updated

```sql
SELECT 
    ar.id, ar.title, ar.status, 
    ar.verified_by, ar.verified_at, ar.updated_at
FROM achievement_references ar 
WHERE ar.id = 'your-achievement-id';

-- Should show:
-- status: 'verified'
-- verified_by: dosen-user-uuid
-- verified_at: timestamp
-- updated_at: timestamp
```

### MongoDB - Points updated

```javascript
db.achievements.findOne({"_id": ObjectId("your-mongo-id")})

// Should show updated points:
// {
//   ...existing fields...,
//   "points": 150,
//   "updatedAt": ISODate("2024-12-13T15:45:30.000Z")
// }
```

## Workflow Testing

Test complete workflow:
1. **FR-003**: Student submit draft
2. **FR-004**: Student submit for verification  
3. **FR-006**: Dosen view prestasi mahasiswa bimbingan
4. **FR-007**: Dosen verify prestasi ✅
5. **FR-011**: Check statistics update

## Points System

Test different point values:
- Competition: 50-200 points
- Publication: 100-300 points  
- Organization: 25-100 points
- Certification: 30-150 points

Dosen wali bisa adjust points berdasarkan:
- Level kompetisi (local/national/international)
- Ranking yang diraih
- Impact dan prestige
- Kualitas achievem