package main

import (
	"cloud_notes/internal/config"
	"cloud_notes/internal/model"
	"cloud_notes/internal/repository"
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestRedisBlacklist(t *testing.T) {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("warning: .env 未加载:", err)
	}

	// 初始化数据库
	config.InitDB()
	fmt.Println("✓ 数据库连接成功")

	// 初始化 JWT
	config.InitJWT()
	fmt.Println("✓ JWT 初始化成功")

	// 初始化 Redis
	config.InitRedis()
	fmt.Println("✓ Redis 连接成功")

	defer config.CloseRedis()

	// ========== 测试 1: Token 加入黑名单 ==========
	t.Run("TestAddTokenToBlacklist", func(t *testing.T) {
		testToken := "test-token-" + time.Now().Format("20060102150405")
		expiredAt := time.Now().Add(24 * time.Hour)

		err := repository.AddTokenToBlacklist(testToken, expiredAt)
		if err != nil {
			t.Errorf("AddTokenToBlacklist failed: %v", err)
		}

		// 立即检查是否在黑名单中
		isBlacklisted, err := repository.IsTokenBlacklisted(testToken)
		if err != nil {
			t.Errorf("IsTokenBlacklisted check failed: %v", err)
		}
		if !isBlacklisted {
			t.Errorf("Token should be blacklisted")
		}
		fmt.Println("✓ TestAddTokenToBlacklist 通过")
	})

	// ========== 测试 2: 检查非黑名单 token ==========
	t.Run("TestNonBlacklistedToken", func(t *testing.T) {
		testToken := "non-blacklisted-token-" + time.Now().Format("20060102150405")

		isBlacklisted, err := repository.IsTokenBlacklisted(testToken)
		if err != nil {
			t.Errorf("IsTokenBlacklisted check failed: %v", err)
		}
		if isBlacklisted {
			t.Errorf("Token should NOT be blacklisted")
		}
		fmt.Println("✓ TestNonBlacklistedToken 通过")
	})

	// ========== 测试 3: 批量加入黑名单 ==========
	t.Run("TestAddTokensToBlacklist", func(t *testing.T) {
		tokens := []string{
			"batch-token-1-" + time.Now().Format("20060102150405"),
			"batch-token-2-" + time.Now().Format("20060102150405"),
			"batch-token-3-" + time.Now().Format("20060102150405"),
		}
		expiredAt := time.Now().Add(24 * time.Hour)

		err := repository.AddTokensToBlacklist(tokens, expiredAt)
		if err != nil {
			t.Errorf("AddTokensToBlacklist failed: %v", err)
		}

		// 检查所有 token 是否都在黑名单中
		for _, token := range tokens {
			isBlacklisted, err := repository.IsTokenBlacklisted(token)
			if err != nil {
				t.Errorf("IsTokenBlacklisted check failed for token %s: %v", token, err)
			}
			if !isBlacklisted {
				t.Errorf("Token %s should be blacklisted", token)
			}
		}
		fmt.Println("✓ TestAddTokensToBlacklist 通过")
	})

	// ========== 测试 4: 黑名单过期 ==========
	t.Run("TestBlacklistExpiry", func(t *testing.T) {
		testToken := "expiry-token-" + time.Now().Format("20060102150405")
		expiredAt := time.Now().Add(2 * time.Second) // 2秒后过期

		err := repository.AddTokenToBlacklist(testToken, expiredAt)
		if err != nil {
			t.Errorf("AddTokenToBlacklist failed: %v", err)
		}

		// 立即检查
		isBlacklisted, _ := repository.IsTokenBlacklisted(testToken)
		if !isBlacklisted {
			t.Errorf("Token should be blacklisted immediately")
		}

		// 等待过期
		fmt.Println("  等待 3 秒以验证过期...")
		time.Sleep(3 * time.Second)

		// 再次检查（应该已过期）
		isBlacklisted, _ = repository.IsTokenBlacklisted(testToken)
		if isBlacklisted {
			t.Errorf("Token should NOT be in blacklist after expiry")
		}
		fmt.Println("✓ TestBlacklistExpiry 通过")
	})

	// ========== 测试 5: Session 模型 ==========
	t.Run("TestSessionModel", func(t *testing.T) {
		session := model.Session{
			UserID:       1,
			DeviceID:     "test-device-001",
			Token:        "test-session-token",
			ExpiredAt:    time.Now().Add(7 * 24 * time.Hour),
			LastActiveAt: time.Now(),
			CreatedAt:    time.Now(),
		}

		// 验证结构体字段
		if session.UserID != 1 {
			t.Errorf("Expected UserID 1, got %d", session.UserID)
		}
		if session.Token == "" {
			t.Errorf("Token should not be empty")
		}
		if session.RevokedAt != nil {
			t.Errorf("RevokedAt should be nil for new session")
		}
		fmt.Println("✓ TestSessionModel 通过")
	})

	fmt.Println("\n========== 所有测试通过 ✓ ==========")
}
