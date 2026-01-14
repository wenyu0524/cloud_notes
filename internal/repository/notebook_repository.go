package repository

import (
	"cloud_notes/internal/model"
	"errors"

	"gorm.io/gorm"
)

type NotebookRepository struct {
	db *gorm.DB
}

func NewNotebookRepository(db *gorm.DB) *NotebookRepository {
	return &NotebookRepository{db: db}
}

func (r *NotebookRepository) Create(nb *model.Notebook) error {
	return r.db.Create(nb).Error
}

func (r *NotebookRepository) FindByUser(userID uint) ([]model.Notebook, error) {
	var notebooks []model.Notebook
	err := r.db.Where("user_id = ?", userID).Find(&notebooks).Error
	return notebooks, err
}

func (r *NotebookRepository) FindByID(id uint) (*model.Notebook, error) {
	var nb model.Notebook
	err := r.db.First(&nb, id).Error
	return &nb, err
}

func (r *NotebookRepository) FindByUserAndName(userID uint, name string) (*model.Notebook, error) {
	var nb model.Notebook
	err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&nb).Error
	return &nb, err
}

// 防越权，防止别人猜到我的ID，从而将别人note绑定到我的notebook
func (r *NotebookRepository) FindByIDAndUser(id uint, userID uint) (*model.Notebook, error) {
	var nb model.Notebook
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&nb).Error
	return &nb, err
}

// 确保默认笔记本存在
func (r *NotebookRepository) EnsureDefaultNotebook(userID uint) (*model.Notebook, error) {
	var nb model.Notebook
	err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&nb).Error
	if err == nil {
		return &nb, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	nb = model.Notebook{
		UserID:    userID,
		Name:      "默认笔记本",
		IsDefault: true,
	}
	if err := r.db.Create(&nb).Error; err != nil {
		return nil, err
	}
	return &nb, nil
}

// 查重
func (r *NotebookRepository) FindByUserAndNameExcludeID(userID uint, name string, excludeID uint) (*model.Notebook, error) {
	var nb model.Notebook
	err := r.db.
		Where("user_id = ? AND name = ? AND id <> ?", userID, name, excludeID).
		First(&nb).Error
	return &nb, err
}

func (r *NotebookRepository) Update(nb *model.Notebook) error {
	return r.db.Save(nb).Error
}

func (r *NotebookRepository) Delete(id uint) error {
	return r.db.Delete(&model.Notebook{}, id).Error
}
