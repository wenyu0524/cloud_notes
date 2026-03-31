package service

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"errors"

	"gorm.io/gorm"
)

type TagService struct {
	repo     *repository.TagRepository
	noteRepo *repository.NoteRepository
}

func NewTagService(repo *repository.TagRepository, noteRepo *repository.NoteRepository) *TagService {
	return &TagService{repo: repo, noteRepo: noteRepo}
}

var ErrTagNameExists = errors.New("标签名已存在")

func (s *TagService) Create(userID uint, name string) (*model.Tag, error) {
	if _, err := s.repo.FindByUserAndName(userID, name); err == nil {
		return nil, ErrTagNameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tag := &model.Tag{
		UserID: userID,
		Name:   name,
	}
	return tag, s.repo.Create(tag)
}

func (s *TagService) List(userID uint) ([]model.Tag, error) {
	return s.repo.FindByUser(userID)
}

func uniqUint(ids []uint) []uint {
	m := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := m[id]; ok {
			continue
		}
		m[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *TagService) BindNoteTags(userID, noteID uint, tagIDs []uint) error {
	// 1) note 必须存在且属于 user
	if _, err := s.noteRepo.FindByIDAndUser(noteID, userID); err != nil {
		return err
	}

	// 2) 去重（避免 (note_id, tag_id) 主键冲突导致事务回滚）
	tagIDs = uniqUint(tagIDs)

	// 3) tag 必须属于 user
	for _, tid := range tagIDs {
		tag, err := s.repo.FindByID(tid)
		if err != nil {
			return err
		}
		if tag.UserID != userID {
			return errors.New("无权限")
		}
	}

	// 4) 绑定（当前语义：覆盖绑定；tagIDs 为空表示清空）
	return s.repo.BindNoteTags(noteID, tagIDs)
}

func (s *TagService) GetNotesByTag(userID, tagID uint) ([]model.Note, error) {
	tag, err := s.repo.FindByID(tagID)
	if err != nil {
		return nil, err
	}
	if tag.UserID != userID {
		return nil, errors.New("无权限")
	}
	return s.repo.FindNotesByTag(tagID)
}

func (s *TagService) Delete(userID, tagID uint) error {
	tag, err := s.repo.FindByID(tagID)
	if err != nil {
		return err
	}
	if tag.UserID != userID {
		return errors.New("无权限")
	}

	// 先删关联表，再删 tag
	if err := s.repo.DeleteNoteTagsByTagID(tagID); err != nil {
		return err
	}
	return s.repo.Delete(tagID)
}
