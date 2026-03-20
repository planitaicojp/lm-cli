package model

// RichMenuBounds defines the bounds of a rich menu area.
type RichMenuBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RichMenuAction defines the action for a rich menu area.
type RichMenuAction struct {
	Type  string `json:"type"`
	Data  string `json:"data,omitempty"`
	Text  string `json:"text,omitempty"`
	Label string `json:"label,omitempty"`
	URI   string `json:"uri,omitempty"`
}

// RichMenuArea defines an area in a rich menu.
type RichMenuArea struct {
	Bounds RichMenuBounds `json:"bounds"`
	Action RichMenuAction `json:"action"`
}

// RichMenuSize defines the size of a rich menu.
type RichMenuSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RichMenu is the full rich menu definition.
type RichMenu struct {
	RichMenuID  string         `json:"richMenuId,omitempty"  yaml:"richMenuId,omitempty"`
	Size        RichMenuSize   `json:"size"`
	Selected    bool           `json:"selected"`
	Name        string         `json:"name"`
	ChatBarText string         `json:"chatBarText"`
	Areas       []RichMenuArea `json:"areas"`
}

// RichMenuRow is a flat representation for table output.
type RichMenuRow struct {
	RichMenuID  string `json:"rich_menu_id"`
	Name        string `json:"name"`
	ChatBarText string `json:"chat_bar_text"`
	Selected    bool   `json:"selected"`
}

// RichMenuListResponse is the response from GET /v2/bot/richmenu/list.
type RichMenuListResponse struct {
	Richmenus []RichMenu `json:"richmenus"`
}

// RichMenuIDResponse is the response from POST /v2/bot/richmenu (create).
type RichMenuIDResponse struct {
	RichMenuID string `json:"richMenuId"`
}

// RichMenuAlias is a rich menu alias.
type RichMenuAlias struct {
	RichMenuAliasID string `json:"richMenuAliasId"`
	RichMenuID      string `json:"richMenuId"`
}

// RichMenuAliasRow is a flat representation for table output.
type RichMenuAliasRow struct {
	RichMenuAliasID string `json:"alias_id"`
	RichMenuID      string `json:"rich_menu_id"`
}

// RichMenuAliasListResponse is the response from GET /v2/bot/richmenu/alias/list.
type RichMenuAliasListResponse struct {
	Aliases []RichMenuAlias `json:"aliases"`
}
