package model

type NoteTag struct {
	NoteID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}
