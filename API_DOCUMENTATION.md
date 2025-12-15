# UAS Achievement Management API Documentation

## Overview

API ini adalah sistem manajemen prestasi mahasiswa yang mendukung:
- Autentikasi dan otorisasi berbasis role (Admin, Dosen, Mahasiswa)
- Manajemen prestasi mahasiswa dengan berbagai kategori
- Workflow verifikasi prestasi oleh dosen wali
- Laporan dan statistik prestasi
- Upload dan manajemen file lampiran

## Base URL

- **Development**: `http://localhost:3000/api/v1`
- **Production**: `https://api.example.com/api/v1`

## Authentication

API menggunakan JWT (JSON Web Token) untuk autentikasi. Token harus disertakan dalam header `Authorization` dengan format:

```
Authorization: Bearer <your-jwt-token>
```

### Mendapatkan Token

1. **Login**: `POST /auth/login`
2. **Refresh Token**: `POST /auth/refresh`

## Roles & Permissions

### 1. Admin
- Mengelola semua pengguna
- Melihat semua prestasi
- Mengatur role pengguna
- Mengatur dosen wali mahasiswa

### 2. Dosen Wali
- Melihat prestasi mahasiswa bimbingan
- Memverifikasi/menolak prestasi
- Melihat laporan mahasiswa bimbingan

### 3. Mahasiswa
- Membuat dan mengelola prestasi sendiri
- Upload lampiran prestasi
- Submit prestasi untuk verifikasi
- Melihat riwayat status prestasi

## Endpoint Categories

### 1. Authentication (`/auth`)
- `POST /login` - Login pengguna
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout pengguna
- `GET /profile` - Mendapatkan profil pengguna

### 2. Users Management (`/users`) - Admin Only
- `GET /` - Daftar semua pengguna
- `POST /` - Membuat pengguna baru
- `GET /{id}` - Detail pengguna
- `PUT /{id}` - Update pengguna
- `DELETE /{id}` - Hapus pengguna
- `PUT /{id}/role` - Update role pengguna

### 3. Achievements (`/achievements`)
- `GET /` - Daftar prestasi (filtered by role)
- `POST /` - Buat prestasi baru (Mahasiswa)
- `GET /{id}` - Detail prestasi
- `PUT /{id}` - Update prestasi (Mahasiswa pemilik)
- `DELETE /{id}` - Hapus prestasi (Mahasiswa pemilik)
- `POST /{id}/submit` - Submit untuk verifikasi
- `POST /{id}/verify` - Verifikasi prestasi (Dosen Wali)
- `POST /{id}/reject` - Tolak prestasi (Dosen Wali)
- `GET /{id}/history` - Riwayat status prestasi
- `POST /{id}/attachments` - Upload lampiran

### 4. Students (`/students`)
- `GET /` - Daftar mahasiswa
- `GET /{id}` - Detail mahasiswa
- `GET /{id}/achievements` - Prestasi mahasiswa
- `PUT /{id}/advisor` - Update dosen wali (Admin)

### 5. Lecturers (`/lecturers`)
- `GET /` - Daftar dosen
- `GET /{id}/advisees` - Prestasi mahasiswa bimbingan

### 6. Reports (`/reports`)
- `GET /statistics` - Statistik umum
- `GET /student/{id}` - Statistik mahasiswa

## Achievement Types

### 1. Competition (Kompetisi)
```json
{
  "achievementType": "competition",
  "details": {
    "competitionName": "Lomba Programming Nasional 2023",
    "competitionLevel": "national",
    "rank": 1,
    "medalType": "gold",
    "eventDate": "2023-06-15T00:00:00Z",
    "location": "Jakarta, Indonesia",
    "organizer": "Kementerian Pendidikan"
  }
}
```

### 2. Publication (Publikasi)
```json
{
  "achievementType": "publication",
  "details": {
    "publicationType": "journal",
    "publicationTitle": "Machine Learning in Healthcare",
    "authors": ["John Doe", "Jane Smith"],
    "publisher": "IEEE",
    "issn": "1234-5678"
  }
}
```

### 3. Organization (Organisasi)
```json
{
  "achievementType": "organization",
  "details": {
    "organizationName": "Himpunan Mahasiswa Teknik Informatika",
    "position": "Ketua",
    "period": {
      "start": "2023-01-01T00:00:00Z",
      "end": "2023-12-31T00:00:00Z"
    }
  }
}
```

### 4. Certification (Sertifikasi)
```json
{
  "achievementType": "certification",
  "details": {
    "certificationName": "AWS Certified Solutions Architect",
    "issuedBy": "Amazon Web Services",
    "certificationNumber": "AWS-CSA-2023-001",
    "validUntil": "2026-01-01T00:00:00Z"
  }
}
```

## Response Format

Semua response menggunakan format standar:

```json
{
  "code": 200,
  "status": "success",
  "message": "Operation successful",
  "data": {
    // Response data
  },
  "meta": {
    // Pagination info (jika ada)
    "page": 1,
    "limit": 10,
    "totalData": 100,
    "totalPage": 10,
    "sortBy": "createdAt",
    "order": "desc",
    "search": ""
  }
}
```

## Error Handling

### Error Response Format
```json
{
  "code": 400,
  "status": "error",
  "message": "Error description"
}
```

### Common HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Pagination

Semua endpoint yang mengembalikan list data mendukung pagination dengan parameter:

- `page` - Nomor halaman (default: 1)
- `limit` - Jumlah item per halaman (default: 10, max: 100)
- `search` - Kata kunci pencarian
- `sortBy` - Field untuk sorting (default: createdAt)
- `order` - Urutan sorting: asc/desc (default: desc)

Contoh:
```
GET /api/v1/achievements?page=1&limit=20&search=programming&sortBy=title&order=asc
```

## File Upload

### Upload Lampiran Prestasi
- **Endpoint**: `POST /achievements/{id}/attachments`
- **Content-Type**: `multipart/form-data`
- **Field**: `files` (array of files)
- **Supported formats**: PDF, DOC, DOCX, JPG, JPEG, PNG
- **Max file size**: 10MB per file
- **Max files**: 5 files per upload

### File Access
Uploaded files dapat diakses melalui:
```
GET /uploads/achievements/{filename}
```

## Rate Limiting

- **Authentication endpoints**: 5 requests per minute per IP
- **General endpoints**: 100 requests per minute per user
- **File upload**: 10 requests per minute per user

## Viewing Documentation

### 1. Swagger UI (Interactive)
Buka file `swagger-ui.html` di browser untuk melihat dokumentasi interaktif dengan fitur "Try it out".

### 2. Swagger YAML
File `swagger.yaml` berisi spesifikasi OpenAPI 3.0 lengkap yang dapat diimport ke tools lain seperti Postman, Insomnia, atau API testing tools lainnya.

### 3. Import ke Postman
1. Buka Postman
2. Click "Import"
3. Pilih file `swagger.yaml`
4. Collection akan otomatis terbuat dengan semua endpoint

## Development Setup

### Prerequisites
- Go 1.19+
- PostgreSQL
- MongoDB
- JWT Secret Key

### Environment Variables
```env
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=uas_achievement
MONGO_URI=mongodb://localhost:27017/uas_achievement
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRE=24h
REFRESH_TOKEN_EXPIRE=168h
```

### Running the Server
```bash
go mod tidy
go run main.go
```

Server akan berjalan di `http://localhost:3000`

## Testing

### Unit Tests
```bash
go test ./...
```

### API Testing
Gunakan collection Postman yang telah diimport dari `swagger.yaml` atau gunakan Swagger UI untuk testing interaktif.

## Support

Untuk pertanyaan atau dukungan teknis, hubungi:
- Email: support@example.com
- Documentation: Lihat file `swagger-ui.html`
- API Spec: `swagger.yaml`