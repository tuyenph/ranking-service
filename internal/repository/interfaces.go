package repository

import "ranking-service/models"

// RedisRepository defines the methods required from a Redis implementation.
type RedisRepository interface {
	UpdateVideoScore(videoID string, delta float64) error
	GetTopVideos(limit int) ([]string, error)
}

// PostgresRepository defines the methods required from a PostgreSQL implementation.
type PostgresRepository interface {
	UpdateVideoScoreInPostgres(videoID, userID string, delta float64) error
	GetUserTopVideosFromDB(userID string, limit int) ([]models.Video, error)
}
