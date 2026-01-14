package repository

import (
	"cloud_notes/internal/model"

	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// 检查是否存在同名
func (r *NoteRepository) ExistsTitle(userID uint, notebookID uint, title string) (bool, error) {
	var cnt int64
	err := r.db.Model(&model.Note{}).
		Where("user_id = ? AND notebook_id = ? AND title = ?", userID, notebookID, title).
		Count(&cnt).Error
	return cnt > 0, err
}

// 创建
func (r *NoteRepository) Create(note *model.Note) error {
	return r.db.Create(note).Error
}

// 查询
func (r *NoteRepository) List(
	userID uint,
	notebookID string,
	tag string,
) ([]model.Note, error) {

	db := r.db.Model(&model.Note{}).
		Where("notes.user_id = ?", userID)

	if notebookID != "" {
		db = db.Where("notes.notebook_id = ?", notebookID)
	}

	if tag != "" {
		db = db.
			Joins("JOIN note_tags ON note_tags.note_id = notes.id").
			Joins("JOIN tags ON tags.id = note_tags.tag_id").
			Where("tags.name = ?", tag)
	}

	var notes []model.Note
	err := db.Order("notes.created_at desc").Find(&notes).Error
	return notes, err
}

// update/delete 前检查 note 归属
func (r *NoteRepository) GetByID(
	id string,
	userID uint,
) (*model.Note, error) {

	var note model.Note
	err := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&note).Error

	if err != nil {
		return nil, err
	}
	return &note, nil
}

// 是否存在同名
func (r *NoteRepository) ExistsTitleExcludeID(userID uint, notebookID uint, title string, excludeID uint) (bool, error) {
	var cnt int64
	err := r.db.Model(&model.Note{}).
		Where("user_id = ? AND notebook_id = ? AND title = ? AND id <> ?", userID, notebookID, title, excludeID).
		Count(&cnt).Error
	return cnt > 0, err
}

// 更新
func (r *NoteRepository) Update(note *model.Note) error {
	return r.db.Save(note).Error
}

// 删除
func (r *NoteRepository) Delete(id string, userID uint) error {
	res := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Note{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 删除某个notebook下的所有note_tags（通过notes过滤，防越权）
func (r *NoteRepository) DeleteNoteTagsByNotebook(userID uint, notebookID uint) error {
	return r.db.
		Table("note_tags").
		Where("note_id IN (?)",
			r.db.Model(&model.Note{}).
				Select("id").
				Where("user_id = ? AND notebook_id = ?", userID, notebookID),
		).
		Delete(nil).Error
}

// 删除某个notebook下的所有notes（防越权）
func (r *NoteRepository) DeleteByNotebook(userID uint, notebookID uint) error {
	return r.db.
		Where("user_id = ? AND notebook_id = ?", userID, notebookID).
		Delete(&model.Note{}).Error
}
