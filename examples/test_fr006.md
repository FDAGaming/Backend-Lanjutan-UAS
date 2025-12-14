# Testing FR-006: View Prestasi Mahasiswa Bimbingan

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan dosen wali dan mahasiswa bimbingan
3. Mahasiswa sudah memiliki prestasi (dari FR-003)
4. Relasi advisor_id sudah diset di tabel students

## Setup Data (Admin Task)

### Step 1: Create Lecturer Profile

```bash
# Login sebagai admin
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@university.ac.id",
    "password": "admin123"
  }'

# Create lecturer profile
curl -X POST http://localhost:3000/api/v1/lecturers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "dosen-user-uuid",
    "lecturerId": "198501012010121001",
    "department": "Teknik Informatika"
  }'
```

### Step 2: Assign Advisor to Student

```bash
# Set dosen wali untuk mahasiswa
curl -X PUT http://localhost:3000/api/v1/students/{STUDENT_UUID}/advisor \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "advisorId": "lecturer-uuid"
  }'
```

## Testing FR-006

### Step 3: Login sebagai Dosen Wali

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dosen@university.ac.id",
    "password": "dosen123"
  }'
```

### Step 4: View Prestasi Mahasiswa Bimbingan (Method 1)

```bash
# Menggunakan endpoint general dengan RBAC auto-filter
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

### Step 5: View Prestasi Mahasiswa Bimbingan (Method 2)

```bash
# Menggunakan dedicated endpoint
curl -X GET http://localhost:3000/api/v1/lecturers/{LECTURER_ID}/advisees \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

### Step 6: Test Pagination & Search

```bash
# Dengan pagination
curl -X GET "http://localhost:3000/api/v1/achievements?page=1&limit=5" \
  -H "Authorization: Bearer DOSEN_TOKEN"

# Dengan search
curl -X GET "http://localhost:3000/api/v1/achievements?search=programming" \
  -H "Authorization: Bearer DOSEN_TOKEN"

# Dengan sorting
curl -X GET "http://localhost:3000/api/v1/achievements?sortBy=title&order=asc" \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

## Expected Response

```json
{
  "code": 200,
  "status": "success",
  "message": "Data retrieved",
  "data": [
    {
      "id": "achievement-ref-uuid",
      "studentId": "student-uuid",
      "mongoAchievementId": "mongodb-object-id",
      "title": "Juara 1 Lomba Programming",
      "status": "submitted",
      "submittedAt": "2024-12-13T10:30:00Z",
      "verifiedAt": null,
      "verifiedBy": null,
      "rejectionNote": "",
      "createdAt": "2024-12-13T09:00:00Z",
      "updatedAt": "2024-12-13T10:30:00Z",
      "student": {
        "studentId": "123456789",
        "user": {
          "fullName": "John Doe"
        }
      }
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "totalData": 5,
    "totalPage": 1,
    "sortBy": "created_at",
    "order": "desc",
    "search": ""
  }
}
```

## Validation Points

✅ **Get list student IDs**: Query JOIN dengan `s.advisor_id = lecturer.id`
✅ **Get achievements references**: Filter berdasarkan advisor_id
✅ **Fetch detail dari MongoDB**: Data lengkap achievement + student info
✅ **Return list dengan pagination**: Support page, limit, search, sort
✅ **Authentication**: JWT token required
✅ **Lecturer validation**: Must have lecturer profile
✅ **RBAC filtering**: Hanya mahasiswa bimbingan sendiri
✅ **Soft delete filtering**: Exclude deleted achievements
✅ **Multiple endpoints**: General + dedicated endpoint

## Error Cases to Test

### 1. Access without authentication

```bash
curl -X GET http://localhost:3000/api/v1/achievements
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 2. Access as non-lecturer

```bash
# Login sebagai mahasiswa, coba akses endpoint dosen
curl -X GET http://localhost:3000/api/v1/lecturers/{LECTURER_ID}/advisees \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Lecturer profile not found"
}
```

### 3. Lecturer without advisees

```bash
# Dosen yang belum punya mahasiswa bimbingan
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer DOSEN_WITHOUT_ADVISEES_TOKEN"
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "Data retrieved",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "totalData": 0,
    "totalPage": 0,
    "sortBy": "created_at",
    "order": "desc",
    "search": ""
  }
}
```

## Database Verification

### Check advisor assignment

```sql
SELECT 
    s.id as student_id,
    s.student_id as nim,
    u.full_name as student_name,
    l.id as lecturer_id,
    l.lecturer_id as nip,
    lu.full_name as lecturer_name
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
WHERE s.advisor_id IS NOT NULL;
```

### Check achievements for specific advisor

```sql
SELECT 
    ar.title,
    ar.status,
    s.student_id as nim,
    u.full_name as student_name
FROM achievement_references ar
JOIN students s ON ar.student_id = s.id
JOIN users u ON s.user_id = u.id
WHERE s.advisor_id = 'lecturer-uuid'
AND ar.status != 'deleted'
ORDER BY ar.created_at DESC;
```

## RBAC Testing

### Test role-based filtering

1. **Admin**: Melihat semua prestasi
2. **Mahasiswa**: Hanya prestasi sendiri
3. **Dosen Wali**: Hanya prestasi mahasiswa bimbingan

```bash
# Login dengan role berbeda dan test endpoint yang sama
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer TOKEN_WITH_DIFFERENT_ROLES"
```

## Performance Considerations

1. **Efficient JOIN**: Single query untuk get data lengkap
2. **Index optimization**: Index pada `advisor_id` dan `student_id`
3. **Pagination**: Limit data per request
4. **Soft delete filtering**: Exclude deleted records

## Next Steps

Setelah FR-006 berhasil, dosen wali bisa:
- FR-007: Verify prestasi mahasiswa bimbingan
- FR-008: Reject prestasi mahasiswa bimbingan
- View detail prestasi untuk verifikasi