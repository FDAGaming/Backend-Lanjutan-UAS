# Sistem Pelaporan Prestasi Mahasiswa (Backend API)

> **Proyek Ujian Akhir Semester (UAS)**
> Mata Kuliah: Pemrograman Backend Lanjut (Praktikum)
> D4 Teknik Informatika - Universitas Airlangga

---

## 👤 Informasi Mahasiswa

| Atribut | Detail |
| :--- | :--- |
| **Nama** | **Falih Dwi Anggara** |
| **NIM** | **434231062** |
| **Kelas** | **C3** |
| **Mata Kuliah** | Pemrograman Backend Lanjutan |

---

## 📖 Deskripsi Proyek

Sistem ini adalah layanan **REST API** yang dirancang untuk mengelola pelaporan prestasi mahasiswa secara digital. Sistem ini memungkinkan mahasiswa untuk melaporkan prestasi dengan atribut yang dinamis, dosen wali untuk melakukan verifikasi, dan admin untuk mengelola seluruh pengguna dan data referensi[cite: 6].

Aplikasi ini mengimplementasikan **Role-Based Access Control (RBAC)** untuk membedakan hak akses antara **Admin**, **Mahasiswa**, dan **Dosen Wali**[cite: 9].

## 🏗️ Arsitektur & Teknologi

Sistem ini menggunakan pendekatan **Hybrid Database** untuk menangani struktur data yang berbeda:

1.  **PostgreSQL (Relasional):** Digunakan untuk manajemen *User*, *Role*, *Permissions* (RBAC), serta data referensi relasional antar entitas (Mahasiswa & Dosen)[cite: 34].
2.  **MongoDB (NoSQL):** Digunakan untuk menyimpan data detail prestasi yang bersifat **dinamis** dan fleksibel tergantung jenis prestasinya (Kompetisi, Organisasi, Publikasi, dll)[cite: 106].

**Tech Stack:**
* **Language:** Go (Golang) 
* **Database:** PostgreSQL & MongoDB
* **Authentication:** JWT (JSON Web Token) [cite: 16]
* **Architecture:** RESTful API

---

## ✨ Fitur Utama

### 1. Autentikasi & Otorisasi (RBAC)
* Login, Refresh Token, dan Logout.
* Middleware untuk memvalidasi permission berdasarkan role (Admin, Mahasiswa, Dosen Wali)[cite: 169].

### 2. Manajemen Prestasi (Mahasiswa)
* **Input Dinamis:** Mendukung berbagai tipe prestasi seperti Akademik, Kompetisi, Organisasi, Publikasi, dan Sertifikasi[cite: 111].
* **Workflow:** Prestasi dimulai dari status `draft`, kemudian di-`submit` untuk verifikasi[cite: 96].
* **Upload Bukti:** Mendukung lampiran file bukti prestasi[cite: 147].

### 3. Verifikasi (Dosen Wali)
* Melihat daftar prestasi mahasiswa bimbingan.
* Melakukan **Approval** (Verified) atau **Rejection** (dengan catatan penolakan)[cite: 212, 222].

### 4. Manajemen User (Admin)
* CRUD User, assign Role, dan mapping data Mahasiswa ke Dosen Wali[cite: 235].

---

## 📂 Struktur Database

### PostgreSQL Schema
Menangani data inti dan relasi:
* `users`, `roles`, `permissions`, `role_permissions`
* `students`, `lecturers`
* `achievement_references` (Menyimpan status dan link ke MongoDB)[cite: 92].

### MongoDB Collection
Menangani detail prestasi (`achievements`):
* Menyimpan field dinamis seperti `rank`, `medalType` (untuk kompetisi), atau `publicationType`, `issn` (untuk publikasi) dalam satu dokumen JSON[cite: 114].

---

## 🚀 Instalasi & Menjalankan Aplikasi

1.  **Clone Repository**
    ```bash
    git clone [https://github.com/username/project-name.git](https://github.com/username/project-name.git)
    cd project-name
    ```

2.  **Konfigurasi Environment**
    Buat file `.env` berdasarkan contoh:
    ```env
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=yourpassword
    DB_NAME=prestasi_db
    
    MONGO_URI=mongodb://localhost:27017
    MONGO_DB_NAME=prestasi_logs
    
    JWT_SECRET=rahasia_super_aman
    ```

3.  **Jalankan Aplikasi**
    ```bash
    go run main.go
    ```

---

## 🔗 Dokumentasi API

Berikut adalah ringkasan endpoint utama yang tersedia:

| Method | Endpoint | Deskripsi | Akses |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/login` | Masuk ke sistem | Public |
| `GET` | `/api/v1/achievements` | List prestasi (Filter by role) | All |
| `POST` | `/api/v1/achievements` | Tambah prestasi baru | Mahasiswa |
| `POST` | `/api/v1/achievements/:id/submit` | Ajukan verifikasi | Mahasiswa |
| `POST` | `/api/v1/achievements/:id/verify` | Setujui prestasi | Dosen Wali |
| `POST` | `/api/v1/achievements/:id/reject` | Tolak prestasi | Dosen Wali |
| `GET` | `/api/v1/reports/statistics` | Statistik prestasi | All |

---

## 🧪 Testing

Strategi testing mencakup **Unit Testing** untuk fungsi individual dan **Mocking** untuk dependensi eksternal (Database/Service)[cite: 295].

---
*Dibuat untuk memenuhi tugas UAS Pemrograman Backend Lanjut.*