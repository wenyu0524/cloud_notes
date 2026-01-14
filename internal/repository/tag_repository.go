package repository

import (
	"cloud_notes/internal/model"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepository) FindByUser(userID uint) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("user_id = ?", userID).Find(&tags).Error
	return tags, err
}

func (r *TagRepository) FindByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.First(&tag, id).Error
	return &tag, err
}

func (r *TagRepository) BindNoteTags(noteID uint, tagIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", noteID).Delete(&model.NoteTag{}).Error; err != nil {
			return err
		}

		for _, tid := range tagIDs {
			if err := tx.Create(&model.NoteTag{
				NoteID: noteID,
				TagID:  tid,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TagRepository) FindByUserAndName(userID uint, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
	return &tag, err
}

func (r *NoteRepository) FindByIDAndUser(noteID uint, userID uint) (*model.Note, error) {
	var note model.Note
	err := r.db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error
	return &note, err
}

func (r *TagRepository) FindByUserAndNameExcludeID(userID uint, name string, excludeID uint) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("user_id = ? AND name = ? AND id <> ?", userID, name, excludeID).First(&tag).Error
	return &tag, err
}

func (r *TagRepository) FindNotesByTag(tagID uint) ([]model.Note, error) {
	var notes []model.Note

	err := r.db.
		Joins("JOIN note_tags ON note_tags.note_id = notes.id").
		Where("note_tags.tag_id = ?", tagID).
		Find(&notes).Error

	return notes, err
}

func (r *TagRepository) DeleteNoteTagsByTagID(tagID uint) error {
	return r.db.Where("tag_id = ?", tagID).Delete(&model.NoteTag{}).Error
}

func (r *TagRepository) Delete(tagID uint) error {
	res := r.db.Delete(&model.Tag{}, tagID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
