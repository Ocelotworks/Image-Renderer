package entity

type Metadata struct {
	ServerID  string `json:"s,omitempty"`
	UserID    string `json:"u,omitempty"`
	ChannelID string `json:"c,omitempty"`
	MessageID string `json:"m,omitempty"`
}
