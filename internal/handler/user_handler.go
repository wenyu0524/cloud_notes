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

type LoginReq struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	DeviceID string `json:"device_id" binding:"required,min=1,max=128"`
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
		c.JSON(http.StatusBadRequest, gin.H{"msg": "用户名和密码不能为空"})
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
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效参数"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.Username == "" || req.Password == "" || req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "username/password/device_id 不能为空"})
		return
	}

	token, err := service.Login(req.Username, req.Password, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效的 username 或 password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// 单设备登出
func Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "未认证"})
		return
	}

	deviceID, exists := c.Get("device_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "缺少 device_id"})
		return
	}

	err := service.Logout(userID.(uint), deviceID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "登出成功"})
}

// 全局登出
func LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "未认证"})
		return
	}

	err := service.LogoutAll(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "全局登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "全局登出成功"})
}
