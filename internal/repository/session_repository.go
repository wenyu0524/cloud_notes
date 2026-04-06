package repository

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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

// 根据user_id获取活跃session数量（未过期且未撤销）
func GetActiveSessionCount(userID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&model.Session{}).
		Where("user_id = ? AND expired_at > ? AND revoked_at IS NULL", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// 根据user_id获取所有活跃session，按last_active_at排序
func GetActiveSessionsByUserID(userID uint) ([]model.Session, error) {
	var sessions []model.Session
	err := config.DB.Where("user_id = ? AND expired_at > ? AND revoked_at IS NULL", userID, time.Now()).
		Order("last_active_at ASC").
		Find(&sessions).Error
	return sessions, err
}

// 删除session（硬删除，兼容旧逻辑）
func DeleteSession(userID uint, deviceID string) error {
	return config.DB.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.Session{}).Error
}

// 删除用户所有session（硬删除）
func DeleteAllSessionsByUserID(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&model.Session{}).Error
}

// 撤销session（软删除，以支持黑名单逻辑）
func RevokeSession(userID uint, deviceID string) error {
	now := time.Now()
	return config.DB.Model(&model.Session{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Updates(map[string]interface{}{"revoked_at": now}).Error
}

// 撤销用户所有session（软删除）
func RevokeAllSessionsByUserID(userID uint) error {
	now := time.Now()
	return config.DB.Model(&model.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]interface{}{"revoked_at": now}).Error
}

// 根据user_id和device_id获取session（必须未撤销）
func GetSessionByUserAndDevice(userID uint, deviceID string) (*model.Session, error) {
	var session model.Session
	err := config.DB.Where("user_id = ? AND device_id = ? AND expired_at > ? AND revoked_at IS NULL", userID, deviceID, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func activeDeviceKey(userID uint) string {
	return fmt.Sprintf("user:%d:devices", userID)
}

// 更新session的last_active_at
func UpdateSessionLastActive(userID uint, deviceID string) error {
	return config.DB.Model(&model.Session{}).Where("user_id = ? AND device_id = ?", userID, deviceID).Update("last_active_at", time.Now()).Error
}

// AddDeviceToActiveSet 将 deviceID 添加到用户活跃设备 Sorted Set 中
func AddDeviceToActiveSet(userID uint, deviceID string, score float64) error {
	ctx := context.Background()
	key := activeDeviceKey(userID)
	return config.RedisClient.ZAdd(ctx, key, redis.Z{Score: score, Member: deviceID}).Err()
}

// RemoveDeviceFromActiveSet 从用户活跃设备 Sorted Set 中删除 deviceID
func RemoveDeviceFromActiveSet(userID uint, deviceID string) error {
	ctx := context.Background()
	key := activeDeviceKey(userID)
	return config.RedisClient.ZRem(ctx, key, deviceID).Err()
}

// GetActiveDeviceCountByRedis 返回用户活跃设备数量
func GetActiveDeviceCountByRedis(userID uint) (int64, error) {
	ctx := context.Background()
	key := activeDeviceKey(userID)
	return config.RedisClient.ZCard(ctx, key).Result()
}

// GetOldestDeviceFromActiveSet 返回最早活跃的设备ID
func GetOldestDeviceFromActiveSet(userID uint) (string, error) {
	ctx := context.Background()
	key := activeDeviceKey(userID)
	devices, err := config.RedisClient.ZRange(ctx, key, 0, 0).Result()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", nil
	}
	return devices[0], nil
}

// ClearActiveDeviceSet 清空用户活跃设备 Sorted Set
func ClearActiveDeviceSet(userID uint) error {
	ctx := context.Background()
	key := activeDeviceKey(userID)
	return config.RedisClient.Del(ctx, key).Err()
}

// 将token加入Redis黑名单（登出/撤销时调用）
func AddTokenToBlacklist(token string, expiredAt time.Time) error {
	ctx := context.Background()
	ttl := time.Until(expiredAt)
	if ttl <= 0 {
		ttl = 1 * time.Second // 最少1秒
	}
	return config.RedisClient.Set(ctx, "blacklist:"+token, "revoked", ttl).Err()
}

// 批量将tokens加入黑名单
func AddTokensToBlacklist(tokens []string, expiredAt time.Time) error {
	ctx := context.Background()
	ttl := time.Until(expiredAt)
	if ttl <= 0 {
		ttl = 1 * time.Second
	}
	pipe := config.RedisClient.Pipeline()
	for _, token := range tokens {
		pipe.Set(ctx, "blacklist:"+token, "revoked", ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// 检查token是否在黑名单中
func IsTokenBlacklisted(token string) (bool, error) {
	ctx := context.Background()
	exists, err := config.RedisClient.Exists(ctx, "blacklist:"+token).Result()
	return exists > 0, err
}
