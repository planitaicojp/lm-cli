package model

// GroupSummary is the response from GET /v2/bot/group/{groupId}/summary.
type GroupSummary struct {
	GroupID     string `json:"groupId"     yaml:"groupId"`
	GroupName   string `json:"groupName"   yaml:"groupName"`
	PictureURL  string `json:"pictureUrl"  yaml:"pictureUrl"`
}

// GroupSummaryRow is a flat representation for table output.
type GroupSummaryRow struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// GroupMembersResponse is the response from GET /v2/bot/group/{groupId}/members/ids.
type GroupMembersResponse struct {
	MemberIDs []string `json:"memberIds"`
	Next      string   `json:"next"`
}
