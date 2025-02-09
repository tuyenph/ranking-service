package repository

import (
	"context"
	"log/slog"
	"ranking-service/config"
	"time"

	"github.com/go-redis/redis/v8"
)

const redisKey = "video_ranking"

var ctx = context.Background()

type RedisDB struct {
	redisClient *redis.Client
}

func NewRedisDB(conf config.RedisConfig) (*RedisDB, error) {
	start := time.Now()
	defer func() {
		slog.Info("Redis connection time", "time", time.Since(start).String())
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr: conf.Host,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisDB{redisClient}, nil
}

// UpdateVideoScore increments the score of a video in the Redis sorted set.
func (r *RedisDB) UpdateVideoScore(videoID string, delta float64) error {
	_, err := r.redisClient.ZIncrBy(ctx, redisKey, delta, videoID).Result()
	return err
}

// GetTopVideos retrieves the top videos based on their score.
func (r *RedisDB) GetTopVideos(limit int) ([]string, error) {
	return r.redisClient.ZRevRange(ctx, redisKey, 0, int64(limit-1)).Result()
}
