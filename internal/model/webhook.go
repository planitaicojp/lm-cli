package model

// WebhookInfo is the response from GET /v2/bot/channel/webhook/endpoint.
type WebhookInfo struct {
	WebhookEndpoint string `json:"webhookEndpoint" yaml:"webhookEndpoint"`
	Active          bool   `json:"active"          yaml:"active"`
}

// WebhookInfoRow is a flat representation for table output.
type WebhookInfoRow struct {
	WebhookEndpoint string `json:"webhook_endpoint"`
	Active          bool   `json:"active"`
}

// WebhookTestResponse is the response from POST /v2/bot/channel/webhook/test.
type WebhookTestResponse struct {
	Success        bool   `json:"success"        yaml:"success"`
	Timestamp      string `json:"timestamp"      yaml:"timestamp"`
	StatusCode     int    `json:"statusCode"     yaml:"statusCode"`
	Reason         string `json:"reason"         yaml:"reason"`
	Detail         string `json:"detail"         yaml:"detail"`
}

// WebhookTestRow is a flat representation for table output.
type WebhookTestRow struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Reason     string `json:"reason"`
}
