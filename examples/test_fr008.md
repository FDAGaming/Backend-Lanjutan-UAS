# Testing FR-008: Reject Prestasi

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
    "title": "Lomba Programming Regional",
    "description": "Mengikuti lomba programming tingkat regional",
    "details": {
      "competitionName": "Regional Programming Contest 2024",
      "competitionLevel": "regional",
      "rank": 3,
      "medalType": "bronze"
    },
    "tags": ["programming", "algorithm"],
    "points": 50
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

## Step 3: Review Achievement Detail

```bash
# Get achievement detail untuk review
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

## Step 4: Reject Prestasi dengan Catatan

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{
    "note": "Bukti sertifikat tidak jelas dan tidak ada dokumentasi foto kegiatan. Mohon lengkapi dengan bukti yang lebih valid seperti sertifikat asli dan foto dokumentasi saat kompetisi."
  }'
```

## Step 5: Verify Response

Expected response:

```json
{
  "code": 200,
  "status": "success",
  "message": "Prestasi berhasil ditolak",
  "data": {
    "id": "achievement-uuid",
    "status": "rejected",
    "rejectionNote": "Bukti sertifikat tidak jelas dan tidak ada dokumentasi foto kegiatan. Mohon lengkapi dengan bukti yang lebih valid seperti sertifikat asli dan foto dokumentasi saat kompetisi.",
    "rejectedBy": "dosen-user-uuid",
    "rejectedAt": "2024-12-13T16:30:00Z",
    "achievement": {
      "title": "Lomba Programming Regional",
      "type": "competition",
      "student": "John Doe",
      "studentId": "123456789"
    }
  }
}
```

## Step 6: Check Server Logs

Server akan menampilkan rejection log:

```
[REJECTION] Achievement rejected:
  - Student: John Doe (123456789)
  - Achievement: Lomba Programming Regional
  - Rejection reason: Bukti sertifikat tidak jelas dan tidak ada dokumentasi foto kegiatan. Mohon lengkapi dengan bukti yang lebih valid seperti sertifikat asli dan foto dokumentasi saat kompetisi.
  - Rejected by: Dr. Jane Smith
  - Time: 2024-12-13 16:30:00
```

## Step 7: Verify Status Change

```bash
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

Response should show:
- `"status": "rejected"`
- `"rejectionNote": "Bukti sertifikat tidak jelas..."`
- `"verifiedBy": dosen-user-uuid` (as rejectedBy)
- `"verifiedAt": timestamp` (as rejectedAt)

## Step 8: Student Check Rejected Achievement

```bash
# Login sebagai mahasiswa untuk melihat rejection
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@university.ac.id",
    "password": "student123"
  }'

# Check achievement status
curl -X GET http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID} \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

Student akan melihat rejection note dan bisa memperbaiki achievement.

## Validation Points

✅ **Precondition check**: Hanya prestasi dengan status 'submitted' yang bisa direject
✅ **Authentication**: JWT token required
✅ **Authorization**: `achievement:verify` permission required
✅ **Lecturer validation**: User harus punya profil dosen wali
✅ **Ownership check**: Hanya dosen wali yang bisa reject prestasi mahasiswa bimbingannya
✅ **Input validation**: Rejection note required dan tidak boleh kosong
✅ **Status update**: Status berubah dari 'submitted' ke 'rejected'
✅ **Save rejection_note**: Catatan penolakan disimpan di database
✅ **Create notification**: Log notification untuk mahasiswa
✅ **Return updated status**: Response dengan data lengkap rejection

## Error Cases to Test

### 1. Reject without note

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Rejection note is required"
}
```

### 2. Reject with empty note

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"note": ""}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Rejection note is required"
}
```

### 3. Reject with too long note

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"note": "Very long note that exceeds 1000 characters..."}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Rejection note too long (max 1000 characters)"
}
```

### 4. Reject non-submitted achievement

```bash
# Coba reject achievement dengan status 'draft' atau 'verified'
curl -X POST http://localhost:3000/api/v1/achievements/{DRAFT_ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"note": "Test rejection"}'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Only submitted achievements can be rejected"
}
```

### 5. Reject achievement bukan mahasiswa bimbingan

```bash
# Login sebagai dosen A, coba reject prestasi mahasiswa bimbingan dosen B
curl -X POST http://localhost:3000/api/v1/achievements/{OTHER_ADVISEE_ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_A_TOKEN" \
  -d '{"note": "Test rejection"}'
```

Expected response:
```json
{
  "code": 403,
  "status": "error",
  "message": "Unauthorized: You can only reject achievements of your advisees"
}
```

### 6. Reject non-existent achievement

```bash
curl -X POST http://localhost:3000/api/v1/achievements/invalid-uuid/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer DOSEN_TOKEN" \
  -d '{"note": "Test rejection"}'
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Achievement not found"
}
```

### 7. Reject without authentication

```bash
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -d '{"note": "Test rejection"}'
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 8. Reject as non-lecturer

```bash
# Login sebagai mahasiswa, coba reject
curl -X POST http://localhost:3000/api/v1/achievements/{ACHIEVEMENT_ID}/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer STUDENT_TOKEN" \
  -d '{"note": "Test rejection"}'
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Lecturer profile not found"
}
```

## Database Verification

### PostgreSQL - Status dan rejection note updated

```sql
SELECT 
    ar.id, ar.title, ar.status, 
    ar.verified_by, ar.verified_at, ar.rejection_note, ar.updated_at
FROM achievement_references ar 
WHERE ar.id = 'your-achievement-id';

-- Should show:
-- status: 'rejected'
-- verified_by: dosen-user-uuid (as rejectedBy)
-- verified_at: timestamp (as rejectedAt)
-- rejection_note: "Bukti sertifikat tidak jelas..."
-- updated_at: timestamp
```

## Common Rejection Reasons

### Competition Achievements
- "Sertifikat tidak mencantumkan nama lengkap yang sesuai dengan data mahasiswa"
- "Bukti dokumentasi kegiatan tidak lengkap, mohon sertakan foto saat kompetisi"
- "Level kompetisi tidak sesuai dengan yang diklaim, mohon klarifikasi"
- "Ranking yang dicantumkan tidak terlihat jelas di sertifikat"

### Publication Achievements
- "Link publikasi tidak dapat diakses atau tidak valid"
- "Nama penulis tidak sesuai dengan data mahasiswa"
- "Jurnal tidak terindeks atau tidak kredibel"
- "Bukti publikasi tidak lengkap, mohon sertakan screenshot halaman artikel"

### Organization Achievements
- "Surat keterangan organisasi tidak resmi atau tidak bermaterai"
- "Periode kepengurusan tidak jelas atau tidak sesuai"
- "Posisi dalam organisasi tidak tercantum dengan jelas"
- "Organisasi tidak diakui atau tidak kredibel"

### Certification Achievements
- "Sertifikat tidak dari lembaga yang diakui"
- "Masa berlaku sertifikat sudah habis"
- "Nama pada sertifikat tidak sesuai dengan data mahasiswa"
- "Sertifikat tidak mencantumkan nomor registrasi yang valid"

## Student Response Workflow

Setelah prestasi direject, mahasiswa bisa:
1. **Melihat rejection note** via GET achievement detail
2. **Memperbaiki bukti** dan upload attachment baru
3. **Update achievement** dengan informasi yang benar
4. **Submit ulang** untuk verifikasi (FR-004)

## Next Steps

Setelah FR-008 berhasil:
- Mahasiswa bisa memperbaiki prestasi yang direject
- Dosen bisa memberikan feedback konstruktif
- System tracking untuk improvement cycle
- Analytics untuk common rejection patterns