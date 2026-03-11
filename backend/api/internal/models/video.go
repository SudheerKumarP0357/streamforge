package models

import "time"

type VideoStatus string

const (
	StatusPending    VideoStatus = "pending"
	StatusProcessing VideoStatus = "processing"
	StatusReady      VideoStatus = "ready"
	StatusFailed     VideoStatus = "failed"
)

type HLSRendition struct {
	ID          string    `json:"id"`
	VideoID     string    `json:"video_id"`
	Resolution  string    `json:"resolution"`
	PlaylistURL string    `json:"playlist_url"`
	Bandwidth   int       `json:"bandwidth"`
	CreatedAt   time.Time `json:"created_at"`
}

type Video struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Title           string         `json:"title"`
	Description     *string        `json:"description"`
	Status          VideoStatus    `json:"status"`
	RawBlobURL      *string        `json:"raw_blob_url,omitempty"`
	DurationSeconds *int           `json:"duration_seconds,omitempty"`
	FileSizeBytes   *int64         `json:"file_size_bytes,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Renditions      []HLSRendition `json:"renditions,omitempty"`
}

type VideoListResponse struct {
	Videos []Video `json:"videos"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
}
