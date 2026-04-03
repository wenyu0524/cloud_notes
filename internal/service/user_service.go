package service

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(username, password string) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return errors.New("username/password 不能为空")
	}

	// 验证是否用户名已存在
	u, err := repository.GetUserByUsername(username)
	if err == nil && u != nil {
		return errors.New("username 已经存在")
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
func Login(username, password, deviceID string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	deviceID = strings.TrimSpace(deviceID)
	if username == "" || password == "" || deviceID == "" {
		return "", errors.New("无效的 username 或 password 或 device_id")
	}

	user, err := repository.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("无效的 username 或 password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("无效的 username 或 password")
	}

	// 处理session
	err = manageSession(user.ID, deviceID)
	if err != nil {
		return "", errors.New("session 管理失败")
	}

	return GenerateToken(user.ID, deviceID)
}

// 管理session：检查活跃数量，删除旧的，创建新的
func manageSession(userID uint, deviceID string) error {
	count, err := repository.GetActiveSessionCount(userID)
	if err != nil {
		return err
	}

	if count >= 2 {
		// 撤销最旧的session（软删除，支持审计）
		sessions, err := repository.GetActiveSessionsByUserID(userID)
		if err != nil {
			return err
		}
		if len(sessions) > 0 {
			err = repository.RevokeSession(sessions[0].UserID, sessions[0].DeviceID)
			if err != nil {
				return err
			}
		}
	}

	// 创建新session
	session := model.Session{
		UserID:       userID,
		DeviceID:     deviceID,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour), // 7天过期
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}
	return repository.CreateOrUpdateSession(&session)
}

// 单设备登出
func Logout(userID uint, deviceID string) error {
	return repository.RevokeSession(userID, deviceID)
}

// 全局登出
func LogoutAll(userID uint) error {
	return repository.RevokeAllSessionsByUserID(userID)
}
