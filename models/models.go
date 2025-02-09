package models

// Video represents a video record in the database.
type Video struct {
	VideoID string `gorm:"primaryKey"`
	UserID  string `gorm:"index"` // Index this field to optimize queries by user_id.
	Score   float64
}

// InteractionRequest represents the payload for updating video score.
// Note: UserID here represents the owner of the video.
type InteractionRequest struct {
	VideoID string  `uri:"video_id" json:"video_id" validate:"required"`
	Type    string  `json:"type" validate:"required"`    // e.g., view, like, comment, share, watch_time
	Weight  float64 `json:"weight" validate:"required"`  // Used for interactions like watch_time.
	UserID  string  `json:"user_id" validate:"required"` // Required: Owner of the video.
}
