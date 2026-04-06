package middleware

import (
	"bytes"
	"cloud_notes/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimit 限制登录接口请求速率，按 IP + username 维度限流。
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"msg": "读取请求失败"})
				c.Abort()
				return
			}
		}

		var payload struct {
			Username string `json:"username"`
		}
		_ = json.Unmarshal(bodyBytes, &payload)
		key := strings.TrimSpace(payload.Username)
		if key == "" {
			key = c.ClientIP()
		} else {
			key = fmt.Sprintf("%s:%s", c.ClientIP(), key)
		}

		rateKey := fmt.Sprintf("rate:login:%s", key)
		ctx := context.Background()
		count, err := config.RedisClient.Incr(ctx, rateKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "限流服务异常"})
			c.Abort()
			return
		}

		if count == 1 {
			_ = config.RedisClient.Expire(ctx, rateKey, time.Minute).Err()
		}

		const limit = 10
		if count > limit {
			c.JSON(http.StatusTooManyRequests, gin.H{"msg": "登录请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}
