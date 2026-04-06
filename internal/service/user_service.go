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

	// 生成 token
	token, err := GenerateToken(user.ID, deviceID)
	if err != nil {
		return "", errors.New("生成 token 失败")
	}

	// 处理session（通过token）
	err = manageSession(user.ID, deviceID, token)
	if err != nil {
		return "", errors.New("session 管理失败")
	}

	return token, nil
}

// 管理session：检查活跃设备数量，删除最旧的设备 session，创建新的
func manageSession(userID uint, deviceID string, token string) error {
	currentSession, _ := repository.GetSessionByUserAndDevice(userID, deviceID)

	count, err := repository.GetActiveDeviceCountByRedis(userID)
	if err != nil {
		count, err = repository.GetActiveSessionCount(userID)
		if err != nil {
			return err
		}
	}

	if currentSession == nil && count >= 2 {
		oldestDevice, err := repository.GetOldestDeviceFromActiveSet(userID)
		if err != nil || oldestDevice == "" {
			sessions, err := repository.GetActiveSessionsByUserID(userID)
			if err != nil {
				return err
			}
			if len(sessions) > 0 {
				oldestDevice = sessions[0].DeviceID
			}
		}

		if oldestDevice != "" && oldestDevice != deviceID {
			oldSession, err := repository.GetSessionByUserAndDevice(userID, oldestDevice)
			if err == nil {
				err = repository.RevokeSession(userID, oldestDevice)
				if err != nil {
					return err
				}
				if oldSession.Token != "" && !oldSession.ExpiredAt.IsZero() {
					_ = repository.AddTokenToBlacklist(oldSession.Token, oldSession.ExpiredAt)
				}
			}
			_ = repository.RemoveDeviceFromActiveSet(userID, oldestDevice)
		}
	}

	// 创建或更新session
	session := model.Session{
		UserID:       userID,
		DeviceID:     deviceID,
		Token:        token,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour), // 7天过期
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}
	if err := repository.CreateOrUpdateSession(&session); err != nil {
		return err
	}

	// 同步 Redis 设备集
	_ = repository.AddDeviceToActiveSet(userID, deviceID, float64(time.Now().Unix()))
	return nil
}

// 单设备登出
func Logout(userID uint, deviceID string) error {
	// 先获取session信息以便获取token和过期时间
	session, err := repository.GetSessionByUserAndDevice(userID, deviceID)
	if err != nil {
		// session不存在或已过期/撤销，仍然返回成功
		_ = repository.RemoveDeviceFromActiveSet(userID, deviceID)
		return nil
	}

	// 撤销session
	err = repository.RevokeSession(userID, deviceID)
	if err != nil {
		return err
	}

	// 将token加入Redis黑名单
	if session.Token != "" && !session.ExpiredAt.IsZero() {
		_ = repository.AddTokenToBlacklist(session.Token, session.ExpiredAt)
	}

	// 同步 Redis 设备集
	_ = repository.RemoveDeviceFromActiveSet(userID, deviceID)

	return nil
}

// 全局登出
func LogoutAll(userID uint) error {
	// 先获取所有活跃session信息
	sessions, err := repository.GetActiveSessionsByUserID(userID)
	if err != nil {
		_ = repository.ClearActiveDeviceSet(userID)
		return repository.RevokeAllSessionsByUserID(userID) // 尽量撤销所有
	}

	// 撤销所有session
	err = repository.RevokeAllSessionsByUserID(userID)
	if err != nil {
		return err
	}

	// 批量将tokens加入黑名单
	tokens := make([]string, 0, len(sessions))
	var expiredAt time.Time
	for _, session := range sessions {
		if session.Token != "" {
			tokens = append(tokens, session.Token)
			if expiredAt.IsZero() || session.ExpiredAt.After(expiredAt) {
				expiredAt = session.ExpiredAt
			}
		}
	}

	if len(tokens) > 0 && !expiredAt.IsZero() {
		_ = repository.AddTokensToBlacklist(tokens, expiredAt)
	}

	// 清空 Redis 设备集
	_ = repository.ClearActiveDeviceSet(userID)

	return nil
}
