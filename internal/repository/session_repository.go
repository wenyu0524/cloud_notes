package repository

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/model"
	"time"
)

// 创建或更新session（同一user_id + device_id覆盖）
func CreateOrUpdateSession(session *model.Session) error {
	var existing model.Session
	err := config.DB.Where("user_id = ? AND device_id = ?", session.UserID, session.DeviceID).First(&existing).Error
	if err == nil {
		// 更新现有session
		existing.Token = session.Token
		existing.ExpiredAt = session.ExpiredAt
		existing.LastActiveAt = session.LastActiveAt
		return config.DB.Save(&existing).Error
	}
	// 创建新session
	return config.DB.Create(session).Error
}

// 根据user_id获取活跃session数量（未过期）
func GetActiveSessionCount(userID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&model.Session{}).Where("user_id = ? AND expired_at > ?", userID, time.Now()).Count(&count).Error
	return count, err
}

// 根据user_id获取所有活跃session，按last_active_at排序
func GetActiveSessionsByUserID(userID uint) ([]model.Session, error) {
	var sessions []model.Session
	err := config.DB.Where("user_id = ? AND expired_at > ?", userID, time.Now()).Order("last_active_at ASC").Find(&sessions).Error
	return sessions, err
}

// 删除session
func DeleteSession(userID uint, deviceID string) error {
	return config.DB.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.Session{}).Error
}

// 删除用户所有session
func DeleteAllSessionsByUserID(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&model.Session{}).Error
}

// 根据user_id和device_id获取session
func GetSessionByUserAndDevice(userID uint, deviceID string) (*model.Session, error) {
	var session model.Session
	err := config.DB.Where("user_id = ? AND device_id = ? AND expired_at > ?", userID, deviceID, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// 更新session的last_active_at
func UpdateSessionLastActive(userID uint, deviceID string) error {
	return config.DB.Model(&model.Session{}).Where("user_id = ? AND device_id = ?", userID, deviceID).Update("last_active_at", time.Now()).Error
}
