# Testing FR-011: Achievement Statistics

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan berbagai prestasi dari mahasiswa berbeda
3. Data prestasi dengan berbagai tipe, level, dan periode

## Actor-Based Testing

### 1. Admin Statistics (All Data)

#### Login sebagai Admin

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@university.ac.id",
    "password": "admin123"
  }'
```

#### Get Overall Statistics

```bash
curl -X GET http://localhost:3000/api/v1/reports/statistics \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "Overall Statistics",
  "data": {
    "totalPerType": {
      "competition": 15,
      "publication": 8,
      "organization": 12,
      "certification": 5
    },
    "totalPerLevel": {
      "international": 3,
      "national": 8,
      "regional": 4,
      "local": 2,
      "unknown": 1
    },
    "totalPerPeriod": {
      "2024-07": 5,
      "2024-08": 8,
      "2024-09": 12,
      "2024-10": 7,
      "2024-11": 6,
      "2024-12": 2
    },
    "topStudents": [
      {
        "name": "John Doe",
        "programStudy": "Teknik Informatika",
        "totalPoints": 450
      },
      {
        "name": "Jane Smith",
        "programStudy": "Sistem Informasi",
        "totalPoints": 380
      },
      {
        "name": "Bob Johnson",
        "programStudy": "Teknik Informatika",
        "totalPoints": 320
      }
    ],
    "summary": {
      "totalAchievements": 40,
      "totalVerified": 25,
      "totalPending": 8,
      "totalRejected": 7,
      "totalPoints": 1850
    }
  }
}
```

### 2. Dosen Wali Statistics (Advisee Data)

#### Login sebagai Dosen Wali

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dosen@university.ac.id",
    "password": "dosen123"
  }'
```

#### Get Advisee Statistics

```bash
curl -X GET http://localhost:3000/api/v1/reports/statistics \
  -H "Authorization: Bearer DOSEN_TOKEN"
```

Expected response (filtered untuk mahasiswa bimbingan):
```json
{
  "code": 200,
  "status": "success",
  "message": "Advisee Statistics",
  "data": {
    "totalPerType": {
      "competition": 6,
      "publication": 2,
      "organization": 4,
      "certification": 1
    },
    "totalPerLevel": {
      "national": 3,
      "regional": 2,
      "local": 1
    },
    "totalPerPeriod": {
      "2024-09": 4,
      "2024-10": 3,
      "2024-11": 2,
      "2024-12": 1
    },
    "topStudents": [
      {
        "name": "Alice Johnson",
        "programStudy": "Teknik Informatika",
        "totalPoints": 280
      },
      {
        "name": "Charlie Brown",
        "programStudy": "Teknik Informatika",
        "totalPoints": 150
      }
    ],
    "summary": {
      "totalAchievements": 13,
      "totalVerified": 8,
      "totalPending": 3,
      "totalRejected": 2,
      "totalPoints": 650
    }
  }
}
```

### 3. Mahasiswa Statistics (Own Data)

#### Login sebagai Mahasiswa

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@university.ac.id",
    "password": "student123"
  }'
```

#### Get Own Statistics

```bash
curl -X GET http://localhost:3000/api/v1/reports/statistics \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

Expected response (hanya data mahasiswa sendiri):
```json
{
  "code": 200,
  "status": "success",
  "message": "Your Statistics",
  "data": {
    "totalPerType": {
      "competition": 3,
      "publication": 1,
      "organization": 2
    },
    "totalPerLevel": {
      "national": 2,
      "regional": 1
    },
    "totalPerPeriod": {
      "2024-09": 2,
      "2024-10": 1,
      "2024-11": 2,
      "2024-12": 1
    },
    "topStudents": [],
    "summary": {
      "totalAchievements": 6,
      "totalVerified": 4,
      "totalPending": 1,
      "totalRejected": 1,
      "totalPoints": 280
    }
  }
}
```

## Student-Specific Statistics

### Get Specific Student Statistics (Admin/Dosen)

```bash
# Admin dapat melihat statistik mahasiswa mana saja
curl -X GET http://localhost:3000/api/v1/reports/student/{USER_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Dosen Wali dapat melihat statistik mahasiswa bimbingannya
curl -X GET http://localhost:3000/api/v1/reports/student/{ADVISEE_USER_ID} \
  -H "Authorization: Bearer DOSEN_TOKEN"

# Mahasiswa melihat statistik sendiri dengan "me"
curl -X GET http://localhost:3000/api/v1/reports/student/me \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

## Validation Points

✅ **Total prestasi per tipe**: Competition, publication, organization, certification
✅ **Total prestasi per periode**: Monthly breakdown (last 6-12 months)
✅ **Top mahasiswa berprestasi**: Top 5 by total points dengan nama dan program studi
✅ **Distribusi tingkat kompetisi**: International, national, regional, local
✅ **Role-based access**: 
  - Admin: Semua data
  - Dosen Wali: Data mahasiswa bimbingan
  - Mahasiswa: Data sendiri
✅ **Summary statistics**: Total counts by status dan total points
✅ **Hybrid data source**: PostgreSQL untuk counts, MongoDB untuk aggregations

## Output Analysis

### 1. Total Prestasi per Tipe
- **competition**: Lomba/kompetisi
- **publication**: Publikasi ilmiah
- **organization**: Kepengurusan organisasi
- **certification**: Sertifikasi

### 2. Total Prestasi per Periode
- Format: `YYYY-MM` (e.g., "2024-12")
- Last 6 months untuk overall stats
- Last 12 months untuk individual stats
- Sorted chronologically

### 3. Top Mahasiswa Berprestasi
- Top 5 berdasarkan total points
- Include nama lengkap dan program studi
- Only students dengan points > 0

### 4. Distribusi Tingkat Kompetisi
- **international**: Tingkat internasional
- **national**: Tingkat nasional
- **regional**: Tingkat regional/wilayah
- **local**: Tingkat lokal/universitas
- **unknown**: Tidak diketahui/tidak diisi

### 5. Summary Statistics
- **totalAchievements**: Total prestasi (exclude deleted)
- **totalVerified**: Prestasi yang sudah diverifikasi
- **totalPending**: Prestasi yang menunggu verifikasi (submitted)
- **totalRejected**: Prestasi yang ditolak
- **totalPoints**: Total poin dari semua prestasi verified

## Error Cases to Test

### 1. Access without authentication

```bash
curl -X GET http://localhost:3000/api/v1/reports/statistics
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 2. Student accessing other student's stats

```bash
curl -X GET http://localhost:3000/api/v1/reports/student/{OTHER_STUDENT_ID} \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

Expected response:
```json
{
  "code": 403,
  "status": "error",
  "message": "Forbidden"
}
```

### 3. Non-existent student

```bash
curl -X GET http://localhost:3000/api/v1/reports/student/invalid-uuid \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "Student not found"
}
```

## Database Verification

### Check aggregation results

```sql
-- Total per status
SELECT status, COUNT(*) 
FROM achievement_references 
WHERE status != 'deleted'
GROUP BY status;

-- Total per type (need to join with MongoDB data)
-- This would require application-level aggregation
```

```javascript
// MongoDB aggregations
// Total per type
db.achievements.aggregate([
  { $group: { _id: "$achievementType", count: { $sum: 1 } } }
])

// Total per level
db.achievements.aggregate([
  { $match: { achievementType: "competition" } },
  { $group: { _id: "$details.competitionLevel", count: { $sum: 1 } } }
])

// Total per period
db.achievements.aggregate([
  { $match: { createdAt: { $gte: new Date("2024-06-01") } } },
  { $group: { 
    _id: { 
      year: { $year: "$createdAt" }, 
      month: { $month: "$createdAt" } 
    }, 
    count: { $sum: 1 } 
  } },
  { $sort: { "_id": 1 } }
])

// Top students by points
db.achievements.aggregate([
  { $match: { points: { $gt: 0 } } },
  { $group: { _id: "$studentId", totalPoints: { $sum: "$points" } } },
  { $sort: { totalPoints: -1 } },
  { $limit: 5 }
])
```

## Performance Considerations

1. **MongoDB Aggregations**: Efficient untuk large datasets
2. **Hybrid Queries**: PostgreSQL untuk relational data, MongoDB untuk analytics
3. **Caching**: Consider caching results untuk frequently accessed stats
4. **Indexing**: Ensure proper indexes pada fields yang di-aggregate

## Use Cases

### Admin Dashboard
- Monitor overall system performance
- Track submission trends
- Identify top performing students
- Analyze achievement distribution

### Dosen Wali Dashboard
- Monitor advisee performance
- Track verification workload
- Compare advisee achievements

### Student Dashboard
- Personal achievement overview
- Progress tracking over time
- Goal setting based on statistics

## Integration with Other Features

Statistics mendukung:
- **Dashboard widgets**: Real-time metrics
- **Reporting**: Export untuk laporan
- **Analytics**: Trend analysis
- **Gamification**: Leaderboards dan achievements