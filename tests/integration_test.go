package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"ranking-service/models"
)

// TestIntegration_UpdateVideoScoreAndRetrieve performs an end-to-end test of updating a video score
// and then retrieving it from both global and per-user endpoints.
func TestIntegration_UpdateVideoScoreAndRetrieve(t *testing.T) {
	baseURL := "http://localhost:8080"

	// Step 1: Update video score for a video.
	videoID := "integration-video-" + strconv.FormatInt(time.Now().Unix(), 10)
	updateURL := baseURL + "/videos/" + videoID + "/interaction"

	interaction := models.InteractionRequest{
		Type:   "like",
		Weight: 0,
		UserID: "integration-user",
	}
	bodyBytes, _ := json.Marshal(interaction)
	resp, err := http.Post(updateURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 on update, got %v", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 2: Get global top videos.
	getGlobalURL := baseURL + "/videos/top?limit=5"
	resp, err = http.Get(getGlobalURL)
	if err != nil {
		t.Fatalf("Failed to send GET request for global videos: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 on global GET, got %v", resp.StatusCode)
	}
	globalBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var globalVideos []string
	if err := json.Unmarshal(globalBody, &globalVideos); err != nil {
		t.Fatalf("Failed to parse global videos JSON: %v", err)
	}
	// Check that our video appears in the global top list.
	foundGlobal := false
	for _, v := range globalVideos {
		if v == videoID {
			foundGlobal = true
			break
		}
	}
	if !foundGlobal {
		t.Errorf("Expected video %s to appear in global top videos", videoID)
	}

	// Step 3: Get user top videos.
	getUserURL := baseURL + "/users/integration-user/videos/top?limit=5"
	resp, err = http.Get(getUserURL)
	if err != nil {
		t.Fatalf("Failed to send GET request for user videos: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 on user GET, got %v", resp.StatusCode)
	}
	userBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var userResp map[string]interface{}
	if err := json.Unmarshal(userBody, &userResp); err != nil {
		t.Fatalf("Failed to parse user top videos JSON: %v", err)
	}
	// Verify that the response contains the correct userID.
	if userID, ok := userResp["userID"].(string); !ok || userID != "integration-user" {
		t.Errorf("Expected userID 'integration-user', got %v", userResp["userID"])
	}
	// Check that there is at least one video returned.
	videos, ok := userResp["videos"].([]interface{})
	if !ok || len(videos) == 0 {
		t.Errorf("Expected at least one video in user top videos")
	}
}
