package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Pastikan Anda sudah set JWT_SECRET di file .env nanti
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type JwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken membuat token baru yang valid selama 24 jam
func GenerateToken(userID string, role string) (string, error) {
	// Jika env kosong, pakai default (HANYA UTK DEVELOPMENT)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("rahasia_default_jangan_dipakai_production")
	}

	claims := JwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken memvalidasi token string dan mengembalikan object token
func ParseToken(tokenString string) (*jwt.Token, error) {
    // Jika env kosong, pakai default
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("rahasia_default_jangan_dipakai_production")
	}

	return jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validasi method signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
}