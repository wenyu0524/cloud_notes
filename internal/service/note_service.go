package service

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"errors"
)

type NoteService struct {
	repo         *repository.NoteRepository
	notebookRepo *repository.NotebookRepository
}

func NewNoteService(repo *repository.NoteRepository, nbRepo *repository.NotebookRepository) *NoteService {
	return &NoteService{repo: repo, notebookRepo: nbRepo}
}

var ErrNoteTitleExists = errors.New("笔记标题已存在")

func (s *NoteService) CreateNote(userID, notebookID uint, title, content string) error {
	// notebook_id没传则放入默认笔记本
	if notebookID == 0 {
		nb, err := s.notebookRepo.EnsureDefaultNotebook(userID)
		if err != nil {
			return err
		}
		notebookID = nb.ID
	} else {
		if _, err := s.notebookRepo.FindByIDAndUser(notebookID, userID); err != nil {
			return errors.New("无效的 notebook_id")
		}
	}

	// 同一notebook下title不可重复
	exists, err := s.repo.ExistsTitle(userID, notebookID, title)
	if err != nil {
		return err
	}
	if exists {
		return ErrNoteTitleExists
	}

	note := &model.Note{
		UserID:     userID,
		NotebookID: notebookID,
		Title:      title,
		Content:    content,
	}
	return s.repo.Create(note)
}

func (s *NoteService) ListNotes(
	userID uint,
	notebookID string,
	tag string,
) ([]model.Note, error) {
	return s.repo.List(userID, notebookID, tag)
}

func (s *NoteService) UpdateNote(id string, userID uint, title string, content string) error {
	note, err := s.repo.GetByID(id, userID)
	if err != nil {
		return err
	}

	// 只有当title真的要改时才查重
	if title != "" && title != note.Title {
		exists, err := s.repo.ExistsTitleExcludeID(userID, note.NotebookID, title, note.ID)
		if err != nil {
			return err
		}
		if exists {
			return ErrNoteTitleExists
		}
		note.Title = title
	}

	// content允许为空覆盖
	note.Content = content

	return s.repo.Update(note)
}

func (s *NoteService) SearchNotes(
	userID uint,
	query string,
	notebookID string,
	tag string,
) ([]model.Note, error) {
	return s.repo.SearchNotes(userID, query, notebookID, tag)
}

func (s *NoteService) DeleteNote(
	id string,
	userID uint,
) error {
	return s.repo.Delete(id, userID)
}
