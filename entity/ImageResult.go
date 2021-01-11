package entity

// ImageResult is a resulting rendered image
type ImageResult struct {
	Data      string `json:"data,omitempty"`
	Extension string `json:"extension,omitempty"`
	Size      int    `json:"size,omitempty"`
	Error     string `json:"err,omitempty"`
}
