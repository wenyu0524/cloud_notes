package service

import (
	"cloud_notes/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 生成token
func GenerateToken(userID uint, deviceID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"device_id": deviceID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}
