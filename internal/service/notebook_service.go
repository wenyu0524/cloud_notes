package service

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type NotebookService struct {
	db       *gorm.DB
	repo     *repository.NotebookRepository
	noteRepo *repository.NoteRepository
}

func NewNotebookService(db *gorm.DB, repo *repository.NotebookRepository, noteRepo *repository.NoteRepository) *NotebookService {
	return &NotebookService{db: db, repo: repo, noteRepo: noteRepo}
}

// 创建笔记本
var ErrNotebookNameExists = errors.New("笔记本名称已存在")

func (s *NotebookService) Create(userID uint, name string) (*model.Notebook, error) {
	// 1) 预检查：同一用户下 name 唯一
	if _, err := s.repo.FindByUserAndName(userID, name); err == nil {
		return nil, ErrNotebookNameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2) 创建
	nb := &model.Notebook{
		UserID: userID,
		Name:   name,
	}
	err := s.repo.Create(nb)
	if err != nil {
		return nil, err
	}
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notebooks:user:%d", userID))
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notes:user:%d:*", userID))
	return nb, nil
}

// 查询笔记本列表
func (s *NotebookService) List(userID uint) ([]model.Notebook, error) {
	cacheKey := repository.CacheKeyNotebooks(userID)
	var notebooks []model.Notebook
	hit, err := repository.GetCache(cacheKey, &notebooks)
	if err == nil && hit {
		return notebooks, nil
	}

	notebooks, err = s.repo.FindByUser(userID)
	if err != nil {
		return nil, err
	}
	_ = repository.SetCache(cacheKey, notebooks, 60*time.Second)
	return notebooks, nil
}

// 更新笔记本
func (s *NotebookService) Update(userID, notebookID uint, name string) error {
	nb, err := s.repo.FindByID(notebookID)
	if err != nil {
		return err
	}
	if nb.UserID != userID {
		return errors.New("无权限")
	}

	// 不允许把默认笔记本改名
	if nb.IsDefault {
		return errors.New("默认笔记本不能被重命名")
	}

	// 更新保证同一用户下name唯一
	if _, err := s.repo.FindByUserAndNameExcludeID(userID, name, notebookID); err == nil {
		return ErrNotebookNameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	nb.Name = name
	err = s.repo.Update(nb)
	if err != nil {
		return err
	}
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notebooks:user:%d", userID))
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notes:user:%d:*", userID))
	return nil
}

// 删除笔记本
func (s *NotebookService) Delete(userID, notebookID uint) error {
	// 1、权限检查
	nb, err := s.repo.FindByID(notebookID)
	if err != nil {
		return err
	}
	if nb.UserID != userID {
		return errors.New("无权限")
	}

	// 2、事务：先删子表，再删父表
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 用事务的tx创建一组repo（保证同一事务连接）
		noteRepo := repository.NewNoteRepository(tx)
		nbRepo := repository.NewNotebookRepository(tx)

		if err := noteRepo.DeleteNoteTagsByNotebook(userID, notebookID); err != nil {
			return err
		}
		if err := noteRepo.DeleteByNotebook(userID, notebookID); err != nil {
			return err
		}
		if err := nbRepo.Delete(notebookID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notebooks:user:%d", userID))
	_ = repository.DeleteCacheByPattern(fmt.Sprintf("cache:notes:user:%d:*", userID))
	return nil
}
