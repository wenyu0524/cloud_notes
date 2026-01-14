package config

import (
	"log"
	"os"
)

// 初始化JWT配置
var JWTSecret []byte

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("No JWT_SECRET")
	}
	JWTSecret = []byte(secret)
}
