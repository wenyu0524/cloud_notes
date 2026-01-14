package service

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(username, password string) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return errors.New("username/password cannot be empty")
	}

	// 验证是否用户名已存在
	u, err := repository.GetUserByUsername(username)
	if err == nil && u != nil {
		return errors.New("username already exists")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("database error")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	return repository.CreateUser(&user)
}

// 登录
func Login(username, password string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return "", errors.New("invalid username or password")
	}

	user, err := repository.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	return GenerateToken(user.ID)
}
