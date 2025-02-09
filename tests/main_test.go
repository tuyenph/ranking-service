// internal/handlers/handlers_test.go
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"ranking-service/internal/handlers"
	"ranking-service/models"
)

// FakeRedis simulates the Redis repository.
type FakeRedis struct {
	UpdateError   error
	TopVideosList []string
	GetError      error
}

func (f *FakeRedis) UpdateVideoScore(videoID string, delta float64) error {
	return f.UpdateError
}

func (f *FakeRedis) GetTopVideos(limit int) ([]string, error) {
	return f.TopVideosList, f.GetError
}

// FakePostgres simulates the PostgreSQL repository.
type FakePostgres struct {
	UpdateError error
	Videos      []models.Video
	GetError    error
}

func (f *FakePostgres) UpdateVideoScoreInPostgres(videoID, userID string, delta float64) error {
	return f.UpdateError
}

func (f *FakePostgres) GetUserTopVideosFromDB(userID string, limit int) ([]models.Video, error) {
	return f.Videos, f.GetError
}

// --- Unit Test Cases ---

func TestUpdateVideoScoreHandler_Success(t *testing.T) {
	// Set up fake repositories with no errors.
	fakeRedis := &FakeRedis{}
	fakePostgres := &FakePostgres{}

	// Create handler instance with fake repos.
	handler := handlers.NewRankingHandler(fakePostgres, fakeRedis)

	// Set up Gin router with the UpdateVideoScore endpoint.
	router := gin.Default()
	router.POST("/videos/:video_id/interaction", handler.UpdateVideoScoreHandler())

	// Create a valid request payload.
	reqPayload := models.InteractionRequest{
		VideoID: "test-video",
		Type:    "like", // expected delta: 1.0
		Weight:  0,
		UserID:  "user123",
	}
	bodyBytes, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/videos/test-video/interaction", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert HTTP response.
	assert.Equal(t, http.StatusOK, w.Code)

	// Decode response.
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "test-video", resp["videoID"])
	assert.Equal(t, float64(1.0), resp["delta"]) // like interaction returns delta 1.0
	assert.Equal(t, "updated", resp["status"])
}

func TestUpdateVideoScoreHandler_MissingUserID(t *testing.T) {
	// Set up fake repositories.
	fakeRedis := &FakeRedis{}
	fakePostgres := &FakePostgres{}

	handler := handlers.NewRankingHandler(fakePostgres, fakeRedis)
	router := gin.Default()
	router.POST("/videos/:video_id/interaction", handler.UpdateVideoScoreHandler())

	// Create payload missing UserID.
	reqPayload := models.InteractionRequest{
		VideoID: "test-video",
		Type:    "like",
		Weight:  0,
		// UserID is missing
	}
	bodyBytes, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/videos/test-video/interaction", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Expect HTTP 400 for missing userID.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGlobalTopVideosHandler(t *testing.T) {
	// Set up fake Redis to return a list of videos.
	fakeRedis := &FakeRedis{
		TopVideosList: []string{"video1", "video2"},
	}
	// Postgres not used in this endpoint.
	fakePostgres := &FakePostgres{}

	handler := handlers.NewRankingHandler(fakePostgres, fakeRedis)
	router := gin.Default()
	router.GET("/videos/top", handler.GetGlobalTopVideosHandler())

	req, _ := http.NewRequest("GET", "/videos/top?limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp))
	assert.Equal(t, "video1", resp[0])
	assert.Equal(t, "video2", resp[1])
}

func TestGetUserTopVideosHandler(t *testing.T) {
	// Set up fake Postgres to return a slice of Video models.
	fakePostgres := &FakePostgres{
		Videos: []models.Video{
			{VideoID: "video1", UserID: "user123", Score: 10},
			{VideoID: "video2", UserID: "user123", Score: 5},
		},
	}
	// Redis not used in this endpoint.
	fakeRedis := &FakeRedis{}

	handler := handlers.NewRankingHandler(fakePostgres, fakeRedis)
	router := gin.Default()
	router.GET("/users/:userID/videos/top", handler.GetUserTopVideosHandler())

	req, _ := http.NewRequest("GET", "/users/user123/videos/top?limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "user123", resp["userID"])

	// Assert that the "videos" field is a list of videos.
	videos, ok := resp["videos"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(videos))
}
