package middleware

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(config.JWTSecret) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "未设置 JWT_SECRET"})
			c.Abort()
			return
		}

		// 1、检查有无token
		autherHeader := c.GetHeader("Authorization")
		if autherHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "缺少 token"})
			c.Abort()
			return
		}

		// 2、检查Authorization头
		parts := strings.SplitN(autherHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 token 头"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 3、解析并验签
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return config.JWTSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 token"})
			c.Abort()
			return
		}

		// 4、检验claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 claims"})
			c.Abort()
			return
		}

		// 5、检验user_id 和 device_id
		v, exists := claims["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "claims 缺少 user_id"})
			c.Abort()
			return
		}
		userID, ok := v.(float64)
		if !ok || userID <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 user_id"})
			c.Abort()
			return
		}

		dv, exists := claims["device_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "claims 缺少 device_id"})
			c.Abort()
			return
		}
		deviceID, ok := dv.(string)
		if !ok || deviceID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "无效 device_id"})
			c.Abort()
			return
		}

		// 6、验证session
		_, err = repository.GetSessionByUserAndDevice(uint(userID), deviceID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "登录已失效"})
			c.Abort()
			return
		}

		// 7、检查活跃设备数量（防止超出限制）
		count, err := repository.GetActiveSessionCount(uint(userID))
		if err != nil || count > 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "超出设备限制"})
			c.Abort()
			return
		}

		// 8、更新last_active_at
		err = repository.UpdateSessionLastActive(uint(userID), deviceID)
		if err != nil {
			// 可选：记录错误，但不中断请求
		}

		c.Set("user_id", uint(userID))
		c.Set("device_id", deviceID)

		c.Next()
	}
}
