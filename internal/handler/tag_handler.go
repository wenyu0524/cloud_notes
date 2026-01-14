package handler

import (
	"cloud_notes/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagHandler struct {
	service *service.TagService
}

func NewTagHandler(s *service.TagService) *TagHandler {
	return &TagHandler{service: s}
}

func (h *TagHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.service.Create(userID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (h *TagHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	tags, err := h.service.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) BindNoteTags(c *gin.Context) {
	userID := c.GetUint("user_id")

	noteID, err := strconv.Atoi(c.Param("id"))
	if err != nil || noteID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var req struct {
		TagIDs []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BindNoteTags(userID, uint(noteID), req.TagIDs); err != nil {
		if err.Error() == "no permission" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tags updated"})
}

func (h *TagHandler) GetNotesByTag(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID, _ := strconv.Atoi(c.Param("id"))

	notes, err := h.service.GetNotesByTag(userID, uint(tagID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func (h *TagHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID, _ := strconv.Atoi(c.Param("id"))

	err := h.service.Delete(userID, uint(tagID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}
		if err.Error() == "no permission" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
