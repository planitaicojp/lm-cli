package model

// FollowerStats is the response from GET /v2/bot/insight/followers.
type FollowerStats struct {
	Status    string `json:"status"    yaml:"status"`
	Followers int    `json:"followers" yaml:"followers"`
	Targeted  int    `json:"targetedReaches" yaml:"targetedReaches"`
	Blocks    int    `json:"blocks"    yaml:"blocks"`
}

// FollowerStatsRow is a flat representation for table output.
type FollowerStatsRow struct {
	Status    string `json:"status"`
	Followers int    `json:"followers"`
	Targeted  int    `json:"targeted_reaches"`
	Blocks    int    `json:"blocks"`
}

// DeliveryStats is the response from GET /v2/bot/insight/message/delivery.
type DeliveryStats struct {
	Status    string `json:"status"    yaml:"status"`
	Broadcast int    `json:"broadcast" yaml:"broadcast"`
	Targeting int    `json:"targeting" yaml:"targeting"`
}

// DeliveryStatsRow is a flat representation for table output.
type DeliveryStatsRow struct {
	Status    string `json:"status"`
	Broadcast int    `json:"broadcast"`
	Targeting int    `json:"targeting"`
}
