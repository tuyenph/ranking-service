package repository

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ranking-service/config"
	"ranking-service/models"
)

type PostgresDB struct {
	db *gorm.DB
}

func NewPostgresDB(conf config.PostgresConfig) (*PostgresDB, error) {
	start := time.Now()
	defer func() {
		slog.Info("Postgres connection time", "time", time.Since(start).String())
	}()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.DB)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Video{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %v", err)
	}

	return &PostgresDB{db}, nil
}

// UpdateVideoScoreInPostgres upserts a video record in PostgreSQL using GORM.
// If the video record does not exist, it creates one; otherwise, it updates the score.
func (p *PostgresDB) UpdateVideoScoreInPostgres(videoID, userID string, delta float64) error {
	var video models.Video
	result := p.db.First(&video, "video_id = ?", videoID)
	if result.Error != nil {
		// If the video is not found, create a new record.
		if result.Error == gorm.ErrRecordNotFound {
			video = models.Video{
				VideoID: videoID,
				UserID:  userID,
				Score:   delta,
			}
			return p.db.Create(&video).Error
		}
		return result.Error
	}
	// Update the existing video's score.
	video.Score += delta
	return p.db.Save(&video).Error
}

// GetUserTopVideosFromDB retrieves the top videos for a given user from PostgreSQL.
func (p *PostgresDB) GetUserTopVideosFromDB(userID string, limit int) ([]models.Video, error) {
	var videos []models.Video
	err := p.db.Where("user_id = ?", userID).Order("score desc").Limit(limit).Find(&videos).Error
	return videos, err
}
