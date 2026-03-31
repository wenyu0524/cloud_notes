package main

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/router"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env 未加载:", err)
	}

	config.InitDB()
	config.InitJWT()

	r := gin.Default() // 初始化Gin

	router.SetupRouter(r) //注册路由

	r.Run(":8080")
}
