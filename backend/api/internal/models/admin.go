package models

type AdminVideo struct {
	Video
	UserEmail string `json:"user_email"`
}

type AdminStats struct {
	TotalUsers     int `json:"total_users"`
	TotalVideos    int `json:"total_videos"`
	VideosByStatus struct {
		Pending    int `json:"pending"`
		Processing int `json:"processing"`
		Ready      int `json:"ready"`
		Failed     int `json:"failed"`
	} `json:"videos_by_status"`
}
