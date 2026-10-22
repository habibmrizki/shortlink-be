package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	slugCachePrefix      = "shortlink:slug:"
	tokenBlacklistPrefix = "shortlink:blacklist:"
)

func SlugCacheKey(slug string) string {
	return slugCachePrefix + slug
}

func BlacklistKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return tokenBlacklistPrefix + hex.EncodeToString(sum[:])
}

// SetSlugCache caches the original URL for a given slug.
func SetSlugCache(ctx context.Context, rdb *redis.Client, slug, originalURL string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	return rdb.Set(ctx, SlugCacheKey(slug), originalURL, ttl).Err()
}

// GetSlugCache returns the cached original URL, or an empty string if missing.
func GetSlugCache(ctx context.Context, rdb *redis.Client, slug string) (string, error) {
	if rdb == nil {
		return "", redis.Nil
	}
	return rdb.Get(ctx, SlugCacheKey(slug)).Result()
}

// InvalidateSlugCache deletes the slug mapping from Redis cache.
func InvalidateSlugCache(ctx context.Context, rdb *redis.Client, slug string) error {
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, SlugCacheKey(slug)).Err()
}

// AddToBlacklist blacklists a JWT token until its expiration time.
func AddToBlacklist(ctx context.Context, rdb *redis.Client, token string, expiresAt time.Time) error {
	if rdb == nil || token == "" {
		return nil
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return rdb.Set(ctx, BlacklistKey(token), "1", ttl).Err()
}

// IsBlacklisted checks if a token has been logged out/blacklisted.
func IsBlacklisted(ctx context.Context, rdb *redis.Client, token string) bool {
	if rdb == nil || token == "" {
		return false
	}
	val, err := rdb.Get(ctx, BlacklistKey(token)).Result()
	if err != nil {
		return false
	}
	return val == "1"
}
