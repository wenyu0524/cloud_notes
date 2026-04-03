package model

import "time"

type Session struct {
	ID           uint       `gorm:"primaryKey"`
	UserID       uint       `gorm:"not null;index"`
	DeviceID     string     `gorm:"size:128;not null"`
	Token        string     `gorm:"size:512"` // 可选，用于存储JWT token
	ExpiredAt    time.Time  `gorm:"not null"`
	LastActiveAt time.Time  `gorm:"not null"`
	RevokedAt    *time.Time `gorm:"index"` // 单点登出/强制踢出标记
	CreatedAt    time.Time
}
