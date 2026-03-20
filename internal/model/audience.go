package model

// AudienceGroup represents an audience group.
type AudienceGroup struct {
	AudienceGroupID int64  `json:"audienceGroupId" yaml:"audienceGroupId"`
	Type            string `json:"type"            yaml:"type"`
	Description     string `json:"description"     yaml:"description"`
	Status          string `json:"status"          yaml:"status"`
	AudienceCount   int    `json:"audienceCount"   yaml:"audienceCount"`
	Created         int64  `json:"created"         yaml:"created"`
}

// AudienceGroupRow is a flat representation for table output.
type AudienceGroupRow struct {
	AudienceGroupID int64  `json:"audience_group_id"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	AudienceCount   int    `json:"audience_count"`
}

// AudienceGroupsResponse is the response from GET /v2/bot/audienceGroup/list.
type AudienceGroupsResponse struct {
	AudienceGroups []AudienceGroup `json:"audienceGroups"`
	HasNextPage    bool            `json:"hasNextPage"`
	TotalCount     int             `json:"totalCount"`
}

// CreateAudienceGroupResponse is the response from POST /v2/bot/audienceGroup/upload.
type CreateAudienceGroupResponse struct {
	AudienceGroupID int64  `json:"audienceGroupId"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	Created         int64  `json:"created"`
}
