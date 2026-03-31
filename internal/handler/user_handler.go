package handler

import (
	"cloud_notes/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserReq struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// 注册
func Register(c *gin.Context) {
	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效参数"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "username/password 不能为空"})
		return
	}

	err := service.Register(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "注册成功"})
}

// 登录
func Login(c *gin.Context) {
	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效参数"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "username/password 不能为空"})
		return
	}

	token, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 username 或 password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
