package model

import "time"

type Notebook struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;uniqueIndex:uk_user_notebook_name"`
	Name      string `gorm:"size:100;not null;uniqueIndex:uk_user_notebook_name"`
	IsDefault bool   `gorm:"not null;default:false;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
