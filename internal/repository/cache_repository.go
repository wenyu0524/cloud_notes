package repository

import (
	"cloud_notes/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func cacheKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func CacheKeyNotes(userID uint, notebookID, tag string) string {
	if strings.TrimSpace(notebookID) == "" {
		notebookID = "all"
	}
	if strings.TrimSpace(tag) == "" {
		tag = "all"
	}
	return cacheKey("cache", "notes", fmt.Sprintf("user:%d", userID), fmt.Sprintf("notebook:%s", notebookID), fmt.Sprintf("tag:%s", tag))
}

func CacheKeyNotebooks(userID uint) string {
	return cacheKey("cache", "notebooks", fmt.Sprintf("user:%d", userID))
}

func CacheKeyTags(userID uint) string {
	return cacheKey("cache", "tags", fmt.Sprintf("user:%d", userID))
}

func SetCache(key string, value interface{}, ttl time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return config.RedisClient.Set(ctx, key, data, ttl).Err()
}

func GetCache(key string, dest interface{}) (bool, error) {
	ctx := context.Background()
	data, err := config.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

func DeleteCacheByPattern(pattern string) error {
	ctx := context.Background()
	iter := config.RedisClient.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return config.RedisClient.Del(ctx, keys...).Err()
}
