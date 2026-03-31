package config

import (
	"cloud_notes/internal/model"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("没有 MYSQL_DSN")
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("连接数据库出错：" + err.Error())
	}
	DB.AutoMigrate(
		&model.User{},
		&model.Note{},
		&model.Notebook{},
		&model.Tag{},
		&model.NoteTag{},
	)
}
