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
	config.InitRedis()
	defer config.CloseRedis()

	r := gin.Default()

	router.SetupRouter(r)

	r.Run(":8080")
}
