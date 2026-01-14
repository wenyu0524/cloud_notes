package model

import "time"

type Note struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"not null;uniqueIndex:uk_user_notebook_title"`
	NotebookID uint   `gorm:"uniqueIndex:uk_user_notebook_title"`
	Title      string `gorm:"size:255;not null;uniqueIndex:uk_user_notebook_title"`
	Content    string `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
