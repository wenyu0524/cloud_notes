package repository

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/model"
)

// 保存对象到数据库
func CreateUser(user *model.User) error {
	return config.DB.Create(user).Error
}

// 在数据库查询用户
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := config.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
