# Testing FR-009: Manage Users

## Prerequisites
1. Server running: `go run main.go`
2. Database setup dengan admin user
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

## Step 2: Create Users

### Create Mahasiswa User

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "username": "john_doe",
    "email": "john.doe@student.university.ac.id",
    "password": "student123",
    "fullName": "John Doe",
    "roleId": "mahasiswa-role-uuid"
  }'
```

### Create Dosen User

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "username": "jane_smith",
    "email": "jane.smith@lecturer.university.ac.id",
    "password": "dosen123",
    "fullName": "Dr. Jane Smith",
    "roleId": "dosen-wali-role-uuid"
  }'
```

Expected response:
```json
{
  "code": 201,
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": "user-uuid",
    "username": "john_doe",
    "email": "john.doe@student.university.ac.id",
    "fullName": "John Doe",
    "roleId": "mahasiswa-role-uuid",
    "isActive": true,
    "createdAt": "2024-12-13T10:00:00Z",
    "updatedAt": "2024-12-13T10:00:00Z"
  }
}
```

## Step 3: List All Users

```bash
curl -X GET http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Step 4: Get User Detail

```bash
curl -X GET http://localhost:3000/api/v1/users/{USER_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Step 5: Update User

```bash
curl -X PUT http://localhost:3000/api/v1/users/{USER_ID} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "fullName": "John Doe Updated",
    "username": "john_doe_updated",
    "email": "john.doe.updated@student.university.ac.id",
    "isActive": true
  }'
```

## Step 6: Assign Role

```bash
curl -X PUT http://localhost:3000/api/v1/users/{USER_ID}/role \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "roleId": "new-role-uuid"
  }'
```

## Step 7: Set Student Profile

```bash
curl -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "student-user-uuid",
    "studentId": "123456789",
    "programStudy": "Teknik Informatika",
    "academicYear": "2024"
  }'
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "Student profile updated"
}
```

## Step 8: Set Lecturer Profile

```bash
curl -X POST http://localhost:3000/api/v1/lecturers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "lecturer-user-uuid",
    "lecturerId": "198501012010121001",
    "department": "Teknik Informatika"
  }'
```

## Step 9: Set Advisor untuk Mahasiswa

```bash
curl -X PUT http://localhost:3000/api/v1/students/{STUDENT_UUID}/advisor \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "advisorId": "lecturer-uuid"
  }'
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "Advisor assigned successfully"
}
```

## Step 10: List Students and Lecturers

### List All Students

```bash
curl -X GET http://localhost:3000/api/v1/students \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### List All Lecturers

```bash
curl -X GET http://localhost:3000/api/v1/lecturers \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Get Student Detail

```bash
curl -X GET http://localhost:3000/api/v1/students/{USER_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Step 11: Delete User

```bash
curl -X DELETE http://localhost:3000/api/v1/users/{USER_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 200,
  "status": "success",
  "message": "User deleted successfully"
}
```

## Validation Points

✅ **Create/update/delete user**: Full CRUD operations
✅ **Assign role**: Role assignment functionality
✅ **Set student/lecturer profile**: Profile management
✅ **Set advisor untuk mahasiswa**: Advisor assignment
✅ **Authentication**: JWT token required
✅ **Authorization**: Admin-only access (`user:manage` permission)
✅ **Password security**: Bcrypt hashing
✅ **Self-deletion prevention**: Cannot delete own account
✅ **Data validation**: Input validation and error handling
✅ **Cascade operations**: Profile deletion when user deleted

## Error Cases to Test

### 1. Access without authentication

```bash
curl -X GET http://localhost:3000/api/v1/users
```

Expected response:
```json
{
  "code": 401,
  "status": "error",
  "message": "Missing authorization header"
}
```

### 2. Access as non-admin

```bash
# Login sebagai mahasiswa/dosen, coba akses user management
curl -X GET http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer NON_ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 403,
  "status": "error",
  "message": "Access denied. Missing permission: user:manage"
}
```

### 3. Create user with duplicate email

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "username": "duplicate",
    "email": "existing@university.ac.id",
    "password": "password123",
    "fullName": "Duplicate User",
    "roleId": "role-uuid"
  }'
```

Expected response:
```json
{
  "code": 500,
  "status": "error",
  "message": "Email already exists"
}
```

### 4. Update non-existent user

```bash
curl -X PUT http://localhost:3000/api/v1/users/invalid-uuid \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "fullName": "Updated Name"
  }'
```

Expected response:
```json
{
  "code": 404,
  "status": "error",
  "message": "User not found"
}
```

### 5. Delete own account

```bash
# Coba delete user ID yang sama dengan admin yang login
curl -X DELETE http://localhost:3000/api/v1/users/{ADMIN_USER_ID} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Cannot delete yourself"
}
```

### 6. Invalid input data

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "username": "",
    "email": "invalid-email",
    "password": "",
    "fullName": "",
    "roleId": ""
  }'
```

Expected response:
```json
{
  "code": 400,
  "status": "error",
  "message": "Invalid input data"
}
```

## Database Verification

### Check user creation

```sql
SELECT id, username, email, full_name, role_id, is_active, created_at 
FROM users 
WHERE email = 'john.doe@student.university.ac.id';
```

### Check student profile

```sql
SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id
FROM students s
JOIN users u ON s.user_id = u.id
WHERE u.email = 'john.doe@student.university.ac.id';
```

### Check lecturer profile

```sql
SELECT l.id, l.user_id, l.lecturer_id, l.department
FROM lecturers l
JOIN users u ON l.user_id = u.id
WHERE u.email = 'jane.smith@lecturer.university.ac.id';
```

### Check advisor assignment

```sql
SELECT 
    s.student_id as nim,
    u.full_name as student_name,
    l.lecturer_id as nip,
    lu.full_name as advisor_name
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
WHERE s.advisor_id IS NOT NULL;
```

## Complete Workflow Example

### 1. Create complete user setup

```bash
# 1. Create user
USER_RESPONSE=$(curl -s -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "username": "alice_johnson",
    "email": "alice.johnson@student.university.ac.id",
    "password": "student123",
    "fullName": "Alice Johnson",
    "roleId": "'$STUDENT_ROLE_ID'"
  }')

USER_ID=$(echo $USER_RESPONSE | jq -r '.data.id')

# 2. Set student profile
curl -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "userId": "'$USER_ID'",
    "studentId": "987654321",
    "programStudy": "Sistem Informasi",
    "academicYear": "2024"
  }'

# 3. Assign advisor
STUDENT_RESPONSE=$(curl -s -X GET http://localhost:3000/api/v1/students/$USER_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN")

STUDENT_UUID=$(echo $STUDENT_RESPONSE | jq -r '.data.id')

curl -X PUT http://localhost:3000/api/v1/students/$STUDENT_UUID/advisor \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "advisorId": "'$LECTURER_UUID'"
  }'
```

## Role Management

### Get available roles (if endpoint exists)

```bash
curl -X GET http://localhost:3000/api/v1/roles \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### Common role UUIDs (from database seeding)

- Admin: `admin-role-uuid`
- Mahasiswa: `mahasiswa-role-uuid`  
- Dosen Wali: `dosen-wali-role-uuid`

## Best Practices

1. **Always create user first**, then set profile
2. **Assign appropriate role** before setting profile
3. **Set advisor after** both student and lecturer profiles exist
4. **Use proper validation** for NIM/NIP format
5. **Handle cascade deletion** when removing users
6. **Maintain referential integrity** in advisor assignments