package model

import "time"

type Tag struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;uniqueIndex:uk_user_tag_name"`
	Name      string `gorm:"size:50;not null;uniqueIndex:uk_user_tag_name"`
	CreatedAt time.Time
}
