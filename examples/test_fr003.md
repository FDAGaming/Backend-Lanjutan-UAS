# Testing FR-003: Submit Prestasi

## Prerequisites
1. Server running: `go run main.go`
2. Database setup: `./scripts/setup-db.sh`
3. Login sebagai mahasiswa untuk mendapat token

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

**Note**: Jika belum ada user mahasiswa, buat dulu melalui admin atau langsung di database.

## Step 2: Submit Prestasi (Competition)

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
    "points": 100,
    "attachments": []
  }'
```

## Step 3: Submit Prestasi (Publication)

```bash
curl -X POST http://localhost:3000/api/v1/achievements \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "achievementType": "publication",
    "title": "Publikasi Paper AI",
    "description": "Publikasi paper tentang machine learning di jurnal internasional",
    "details": {
      "publicationType": "journal",
      "publicationTitle": "Advanced Machine Learning Techniques",
      "authors": ["John Doe", "Jane Smith"],
      "publisher": "IEEE",
      "issn": "1234-5678"
    },
    "tags": ["research", "AI", "publication"],
    "points": 150,
    "attachments": []
  }'
```

## Step 4: Upload Attachment

```bash
# Upload file attachment ke achievement yang sudah dibuat
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/attachments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/certificate.pdf"
```

## Step 5: Verify Response

Expected response untuk submit prestasi:

```json
{
  "code": 201,
  "status": "success",
  "message": "Prestasi berhasil disimpan sebagai draft",
  "data": {
    "referenceId": "uuid-postgres-id",
    "mongoId": "mongodb-object-id",
    "status": "draft",
    "points": 100
  }
}
```

## Step 6: Check Achievement List

```bash
curl -X GET http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Validation Points

✅ **Data tersimpan di MongoDB**: Detail prestasi dengan field dinamis
✅ **Reference tersimpan di PostgreSQL**: Metadata dan status
✅ **Status awal 'draft'**: Sesuai requirement
✅ **Permission check**: Hanya mahasiswa yang bisa submit
✅ **Student validation**: User harus punya profil mahasiswa
✅ **Hybrid transaction**: Rollback jika salah satu database gagal
✅ **File upload**: Support attachment dengan validasi
✅ **Dynamic fields**: Mendukung berbagai jenis prestasi

## Error Cases to Test

1. **Unauthorized access** (no token):
```bash
curl -X POST http://localhost:3000/api/v1/achievements \
  -H "Content-Type: application/json" \
  -d '{"title": "test"}'
# Expected: 401 Unauthorized
```

2. **Non-student user**:
```bash
# Login sebagai admin/dosen, lalu coba submit
# Expected: 404 "Student profile not found"
```

3. **Invalid file upload**:
```bash
# Upload file > 2MB atau tipe tidak didukung
# Expected: 400 Bad Request
```

4. **Invalid JSON**:
```bash
curl -X POST http://localhost:3000/api/v1/achievements \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d 'invalid json'
# Expected: 400 Invalid input
```