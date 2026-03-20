package model

// UserProfile is the response from GET /v2/bot/profile/{userId}.
type UserProfile struct {
	UserID        string `json:"userId"        yaml:"userId"`
	DisplayName   string `json:"displayName"   yaml:"displayName"`
	PictureURL    string `json:"pictureUrl"    yaml:"pictureUrl"`
	StatusMessage string `json:"statusMessage" yaml:"statusMessage"`
	Language      string `json:"language"      yaml:"language"`
}

// UserProfileRow is a flat representation for table output.
type UserProfileRow struct {
	UserID        string `json:"user_id"`
	DisplayName   string `json:"display_name"`
	Language      string `json:"language"`
	StatusMessage string `json:"status_message"`
}

// FollowerIDsResponse is the response from GET /v2/bot/followers/ids.
type FollowerIDsResponse struct {
	UserIDs []string `json:"userIds"`
	Next    string   `json:"next"`
}
