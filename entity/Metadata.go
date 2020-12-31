package entity

// Metadata is metadata about an image request
type Metadata struct {
	ServerID  string `json:"s,omitempty"`
	UserID    string `json:"u,omitempty"`
	ChannelID string `json:"c,omitempty"`
	MessageID string `json:"m,omitempty"`
}
