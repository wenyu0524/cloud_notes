package middleware

import (
	"cloud_notes/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(config.JWTSecret) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "JWT_SECRET not configured"})
			c.Abort()
			return
		}

		// 1、检查有无token
		autherHeader := c.GetHeader("Authorization")
		if autherHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "no token"})
			c.Abort()
			return
		}

		// 2、检查Authorization头
		parts := strings.SplitN(autherHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid token format"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			c.Abort()
			return
		}

		// 4、检验claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid claims"})
			c.Abort()
			return
		}

		// 5、检验user_id
		v, exists := claims["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "missing user_id"})
			c.Abort()
			return
		}
		userID, ok := v.(float64)
		if !ok || userID <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid user_id"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))

		c.Next()
	}
}
