package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ranking-service/internal/repository"
	"ranking-service/models"
)

type RankingHandler struct {
	postgres repository.PostgresRepository
	redis    repository.RedisRepository
}

func NewRankingHandler(postgres repository.PostgresRepository, redis repository.RedisRepository) *RankingHandler {
	return &RankingHandler{postgres, redis}
}

// UpdateVideoScoreHandler updates a video's score based on an interaction.
//
//	@Summary		Update video score based on interaction
//	@Description	Update a video's score by processing interactions (views, likes, etc.). The payload must include userID.
//	@Tags			Videos
//	@Accept			json
//	@Produce		json
//	@Param			video_id	path		string						true	"Video ID"
//	@Param			interaction	body		models.InteractionRequest	true	"Interaction payload"
//	@Success		200			{object}	map[string]interface{}
//	@Router			/videos/{video_id}/interaction [post]
func (h *RankingHandler) UpdateVideoScoreHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		var req models.InteractionRequest
		// Must use ShouldBindJSON before ShouldBindUri
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URI path for VideoID"})
			return
		}

		fmt.Printf("req: %#v\n", req)

		// Ensure userID is provided.
		if req.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userID in payload"})
			return
		}

		// Determine score delta based on interaction type.
		var delta float64
		switch req.Type {
		case "view":
			delta = 0.1
		case "like":
			delta = 1.0
		case "comment":
			delta = 1.5
		case "share":
			delta = 2.0
		case "watch_time":
			delta = req.Weight // Weight provided dynamically.
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown interaction type"})
			return
		}

		// Update the score in Redis.
		if err := h.redis.UpdateVideoScore(req.VideoID, delta); err != nil {
			slog.Error("UpdateVideoScoreHandler: Failed to update score in Redis", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update score in Redis", "err_details": err.Error()})
			return
		}

		// Update (or create) the video record in PostgreSQL using GORM.
		if err := h.postgres.UpdateVideoScoreInPostgres(req.VideoID, req.UserID, delta); err != nil {
			slog.Error("UpdateVideoScoreHandler: Failed to update score in PostgreSQL", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update score in PostgreSQL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"videoID": req.VideoID,
			"delta":   delta,
			"status":  "updated",
		})
	}
}

// GetGlobalTopVideosHandler retrieves the top-ranked videos globally using Redis.
//
//	@Summary		Retrieve global top videos
//	@Description	Get the top ranked videos globally.
//	@Tags			Videos
//	@Accept			json
//	@Produce		json
//	@Param			limit	query	int	false	"Number of videos to retrieve"
//	@Success		200		{array}	string
//	@Router			/videos/top [get]
func (h *RankingHandler) GetGlobalTopVideosHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		limit := 10 // default value
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		videos, err := h.redis.GetTopVideos(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching top videos"})
			return
		}
		c.JSON(http.StatusOK, videos)
	}
}

// GetUserTopVideosHandler retrieves the top videos for a given user from PostgreSQL.
//
//	@Summary		Retrieve personalized top videos for a user
//	@Description	Get the top ranked videos for a specific user.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Param			limit	query		int		false	"Number of videos to retrieve"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/users/{userID}/videos/top [get]
func (h *RankingHandler) GetUserTopVideosHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		limit := 10 // default value
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		// Retrieve top videos for the user from PostgreSQL.
		videos, err := h.postgres.GetUserTopVideosFromDB(userID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching personalized videos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"videos": videos,
		})
	}
}
