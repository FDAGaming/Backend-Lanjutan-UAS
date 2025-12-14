# Testing FR-010: View All Achievements (Admin)

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan berbagai prestasi dari mahasiswa berbeda
3. Login sebagai admin untuk mendapat token

## Step 1: Login sebagai Admin

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@university.ac.id",
    "password": "admin123"
  }'
```

Save the JWT token untuk digunakan di request berikutnya.

## Step 2: View All Achievements (Basic)

```bash
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "Data retrieved",
  "data": [
    {
      "id": "achievement-ref-uuid-1",
      "studentId": "student-uuid-1",
      "mongoAchievementId": "mongodb-object-id-1",
      "title": "Juara 1 Lomba Programming Nasional",
      "status": "verified",
      "submittedAt": "2024-12-13T10:30:00Z",
      "verifiedAt": "2024-12-13T15:45:00Z",
      "verifiedBy": "dosen-uuid",
      "rejectionNote": "",
      "createdAt": "2024-12-13T09:00:00Z",
      "updatedAt": "2024-12-13T15:45:00Z",
      "student": {
        "studentId": "123456789",
        "user": {
          "fullName": "John Doe"
        }
      }
    },
    {
      "id": "achievement-ref-uuid-2",
      "studentId": "student-uuid-2",
      "mongoAchievementId": "mongodb-object-id-2",
      "title": "Publikasi Paper AI",
      "status": "submitted",
      "submittedAt": "2024-12-13T11:00:00Z",
      "verifiedAt": null,
      "verifiedBy": null,
      "rejectionNote": "",
      "createdAt": "2024-12-13T10:30:00Z",
      "updatedAt": "2024-12-13T11:00:00Z",
      "student": {
        "studentId": "987654321",
        "user": {
          "fullName": "Jane Smith"
        }
      }
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "totalData": 25,
    "totalPage": 3,
    "sortBy": "created_at",
    "order": "desc",
    "search": ""
  }
}
```

## Step 3: Apply Pagination

### Get specific page

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?page=2&limit=5" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Change page size

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?limit=20" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Step 4: Apply Filters dan Sorting

### Search by title

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?search=programming" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Search by status

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?search=verified" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Sort by title (ascending)

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?sortBy=title&order=asc" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Sort by status

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?sortBy=status&order=desc" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Combined filters

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?page=1&limit=5&search=lomba&sortBy=created_at&order=desc" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Step 5: Get Achievement Detail

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response with MongoDB details:
```json
{
  "code": 200,
  "status": "success",
  "message": "Detail retrieved",
  "data": {
    "meta": {
      "id": "achievement-ref-uuid",
      "title": "Juara 1 Lomba Programming Nasional",
      "status": "verified",
      "student": {
        "studentId": "123456789",
        "user": {
          "fullName": "John Doe"
        }
      }
    },
    "content": {
      "id": "mongodb-object-id",
      "achievementType": "competition",
      "title": "Juara 1 Lomba Programming Nasional",
      "description": "Meraih juara 1 dalam lomba programming tingkat nasional",
      "details": {
        "competitionName": "National Programming Contest 2024",
        "competitionLevel": "national",
        "rank": 1,
        "medalType": "gold"
      },
      "attachments": [
        {
          "fileName": "certificate.pdf",
          "fileUrl": "/uploads/1234567890_certificate.pdf",
          "fileType": "application/pdf",
          "uploadedAt": "2024-12-13T09:30:00Z"
        }
      ],
      "tags": ["programming", "algorithm", "competition"],
      "points": 150
    }
  }
}
```

## Validation Points

✅ **Get all achievement references**: Query PostgreSQL dengan JOIN
✅ **Fetch details dari MongoDB**: Complete data via hybrid approach
✅ **Apply filters dan sorting**: Search, sort, pagination support
✅ **Return dengan pagination**: Meta information included
✅ **Admin access**: Melihat semua prestasi tanpa filter
✅ **RBAC implementation**: Role-based data filtering
✅ **Soft delete filtering**: Exclude deleted achievements
✅ **Performance optimization**: Efficient JOIN queries

## Admin-Specific Features

### 1. View All Statuses

Admin dapat melihat prestasi dengan semua status:
- `draft` - Prestasi yang belum disubmit
- `submitted` - Prestasi yang menunggu verifikasi
- `verified` - Prestasi yang sudah diverifikasi
- `rejected` - Prestasi yang ditolak

### 2. View All Students

Admin dapat melihat prestasi dari semua mahasiswa, tidak terbatas pada:
- Mahasiswa tertentu (berbeda dengan role Mahasiswa)
- Mahasiswa bimbingan (berbeda dengan role Dosen Wali)

### 3. Complete Information

Admin mendapat informasi lengkap:
- Student details (name, NIM)
- Achievement metadata (status, timestamps)
- Verifier information (who verified/rejected)
- Rejection notes (if any)

## Comparison with Other Roles

### Admin vs Mahasiswa

```bash
# Admin - melihat semua
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Mahasiswa - hanya milik sendiri
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

### Admin vs Dosen Wali

```bash
# Admin - melihat semua
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Dosen Wali - hanya mahasiswa bimbingan
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer LECTURER_TOKEN"
```

## Advanced Filtering Examples

### Filter by achievement type (via search)

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?search=competition" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Filter by student name (via search)

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?search=john" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Recent achievements (sort by created_at desc)

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?sortBy=created_at&order=desc&limit=20" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Pending verifications (search submitted)

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?search=submitted&sortBy=created_at&order=asc" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

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

### 2. Invalid pagination parameters

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?page=0&limit=-1" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Should handle gracefully with default values.

### 3. Invalid sort parameters

```bash
curl -X GET "http://localhost:3000/api/v1/achievements?sortBy=invalid_field" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Should fallback to default sorting.

## Performance Testing

### Large dataset pagination

```bash
# Test with large page numbers
curl -X GET "http://localhost:3000/api/v1/achievements?page=100&limit=50" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Search performance

```bash
# Test search with various terms
curl -X GET "http://localhost:3000/api/v1/achievements?search=a" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Database Verification

### Check total count

```sql
SELECT COUNT(*) 
FROM achievement_references ar 
WHERE ar.status != 'deleted';
```

### Check search functionality

```sql
SELECT ar.id, ar.title, ar.status, u.full_name
FROM achievement_references ar
JOIN students s ON ar.student_id = s.id
JOIN users u ON s.user_id = u.id
WHERE ar.status != 'deleted'
AND (LOWER(ar.title) LIKE '%programming%' OR LOWER(ar.status) LIKE '%programming%')
ORDER BY ar.created_at DESC;
```

### Check sorting

```sql
SELECT ar.id, ar.title, ar.created_at
FROM achievement_references ar
WHERE ar.status != 'deleted'
ORDER BY ar.title ASC
LIMIT 10 OFFSET 0;
```

## Analytics Use Cases

Admin dapat menggunakan endpoint ini untuk:

1. **Monitor submission trends**: Sort by created_at
2. **Track verification status**: Filter by status
3. **Find specific achievements**: Search functionality
4. **Student performance overview**: View all students' achievements
5. **System usage statistics**: Total counts and pagination info

## Integration with Other Features

FR-010 mendukung workflow admin untuk:
- **FR-009**: Manage users - melihat prestasi per user
- **FR-011**: Statistics - data source untuk analytics
- **FR-007/008**: Verification workflow - melihat pending submissions
- **Reporting**: Export data untuk laporan