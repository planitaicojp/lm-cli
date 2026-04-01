package model

// BotInfo is the response from GET /v2/bot/info.
type BotInfo struct {
	UserID          string `json:"userId"          yaml:"userId"`
	BasicID         string `json:"basicId"         yaml:"basicId"`
	DisplayName     string `json:"displayName"     yaml:"displayName"`
	PictureURL      string `json:"pictureUrl"      yaml:"pictureUrl"`
	ChatMode        string `json:"chatMode"        yaml:"chatMode"`
	MarkAsReadMode  string `json:"markAsReadMode"  yaml:"markAsReadMode"`
}

// BotInfoRow is a flat representation for table output.
type BotInfoRow struct {
	UserID      string `json:"user_id"`
	BasicID     string `json:"basic_id"`
	DisplayName string `json:"display_name"`
	ChatMode    string `json:"chat_mode"`
}

// QuotaInfo is the response from GET /v2/bot/message/quota.
type QuotaInfo struct {
	Type  string `json:"type"  yaml:"type"`
	Value int    `json:"value" yaml:"value"`
}

// QuotaRow is a flat representation for table output.
type QuotaRow struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// ConsumptionInfo is the response from GET /v2/bot/message/quota/consumption.
type ConsumptionInfo struct {
	TotalUsage int `json:"totalUsage" yaml:"totalUsage"`
}

// ConsumptionRow is a flat representation for table output.
type ConsumptionRow struct {
	TotalUsage int `json:"total_usage"`
}

// StatusInfo is the JSON/YAML output for lm status.
type StatusInfo struct {
	API            string `json:"api"              yaml:"api"`
	BotID          string `json:"bot_id"           yaml:"bot_id"`
	DisplayName    string `json:"display_name"     yaml:"display_name"`
	TokenType      string `json:"token_type"       yaml:"token_type"`
	TokenExpiresAt string `json:"token_expires_at" yaml:"token_expires_at"`
}

// StatusRow is a flat representation for table output.
type StatusRow struct {
	API   string `json:"api"`
	Bot   string `json:"bot"`
	Token string `json:"token"`
}

// BotUsageRow combines quota and consumption for table output.
type BotUsageRow struct {
	Type      string  `json:"type"`
	Limit     int     `json:"limit"`
	Used      int     `json:"used"`
	Remaining int     `json:"remaining"`
	UsagePct  float64 `json:"usage_pct"`
}
