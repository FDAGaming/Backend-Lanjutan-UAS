# API Routes Summary

## 5.1 Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/profile` - Get user profile (authenticated)

## 5.2 Users (Admin)
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/:id` - Get user detail
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `PUT /api/v1/users/:id/role` - Assign role to user

## 5.4 Achievements
- `GET /api/v1/achievements` - List achievements (filtered by role)
- `GET /api/v1/achievements/:id` - Get achievement detail
- `POST /api/v1/achievements` - Create achievement (Mahasiswa)
- `PUT /api/v1/achievements/:id` - Update achievement (Mahasiswa)
- `DELETE /api/v1/achievements/:id` - Delete achievement (Mahasiswa)
- `POST /api/v1/achievements/:id/submit` - Submit for verification
- `POST /api/v1/achievements/:id/verify` - Verify achievement (Dosen Wali)
- `POST /api/v1/achievements/:id/reject` - Reject achievement (Dosen Wali)
- `GET /api/v1/achievements/:id/history` - Get status history
- `POST /api/v1/achievements/:id/attachments` - Upload files

## 5.5 Students & Lecturers
- `GET /api/v1/students` - List all students
- `GET /api/v1/students/:id` - Get student detail
- `GET /api/v1/students/:id/achievements` - Get student achievements
- `PUT /api/v1/students/:id/advisor` - Set advisor for student
- `GET /api/v1/lecturers` - List all lecturers
- `GET /api/v1/lecturers/:id/advisees` - Get advisee achievements

## 5.8 Reports & Analytics
- `GET /api/v1/reports/statistics` - Get overall statistics
- `GET /api/v1/reports/student/:id` - Get student-specific statistics

## Authentication & Authorization

### Authentication Required
All endpoints except login require JWT token in Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Role-Based Permissions
- **Admin**: Full access to all endpoints
- **Mahasiswa**: Can create, update, delete own achievements
- **Dosen Wali**: Can verify/reject achievements of advisees

### Permission Requirements
- `user:manage` - Admin only (user management)
- `achievement:create` - Mahasiswa (create/submit achievements)
- `achievement:update` - Mahasiswa (update own achievements)
- `achievement:delete` - Mahasiswa (delete own achievements)
- `achievement:verify` - Dosen Wali (verify/reject achievements)

## HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Response Format
All responses follow this format:
```json
{
  "code": 200,
  "status": "success",
  "message": "Operation successful",
  "data": {...},
  "meta": {...} // For paginated responses
}
```

## Pagination Support
List endpoints support pagination:
- `?page=1` - Page number (default: 1)
- `?limit=10` - Items per page (default: 10)
- `?search=keyword` - Search filter
- `?sortBy=field` - Sort field
- `?order=asc|desc` - Sort order (default: desc)

## File Upload
Achievement attachments support:
- Max file size: 2MB
- Allowed types: PDF, DOC, DOCX, JPG, PNG
- Files stored in `/uploads` directory
- Accessible via `/uploads/{filename}`