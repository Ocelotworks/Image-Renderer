package entity

// ImageRequest is a request to render an image
type ImageRequest struct {
	ImageComponents []*ImageComponent `json:"components"`
	Metadata        *Metadata         `json:"metadata"`
	Width           int               `json:"width"`
	Height          int               `json:"height"`
	Debug           bool              `json:"debug"`
	Version         int               `json:"version"`
	Compression     bool              `json:"compression"`
	MaxWidth        int               `json:"maxWidth"`
}
