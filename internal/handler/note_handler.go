package handler

import (
	"cloud_notes/internal/model"
	"cloud_notes/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NoteHandler struct {
	service *service.NoteService
}

func NewNoteHandler(s *service.NoteService) *NoteHandler {
	return &NoteHandler{service: s}
}

// 创建笔记
type CreateNoteReq struct {
	NotebookID uint   `json:"notebook_id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
}

func (h *NoteHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if err := h.service.CreateNote(userID, req.NotebookID, req.Title, req.Content); err != nil {
		if errors.Is(err, service.ErrNoteTitleExists) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		if err.Error() == "invalid notebook_id" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "笔记已创建"})
}

// 查询笔记
func (h *NoteHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	notebookID := c.Query("notebook_id")
	tag := c.Query("tag")
	query := c.Query("q") // 新增搜索参数

	var notes []model.Note
	var err error

	if query != "" {
		// 全文搜索
		notes, err = h.service.SearchNotes(userID, query, notebookID, tag)
	} else {
		// 普通列表
		notes, err = h.service.ListNotes(userID, notebookID, tag)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": notes})
}

// 更新笔记
type UpdateNoteReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *NoteHandler) Update(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var req UpdateNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if err := h.service.UpdateNote(
		id,
		userID,
		req.Title,
		req.Content,
	); err != nil {
		if errors.Is(err, service.ErrNoteTitleExists) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"msg": "未找到笔记"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "笔记已更新"})
}

// 删除笔记
func (h *NoteHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	if err := h.service.DeleteNote(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"msg": "未找到笔记"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "笔记已删除"})
}
