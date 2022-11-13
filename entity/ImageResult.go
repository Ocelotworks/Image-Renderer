package entity

// ImageResult is a resulting rendered image
type ImageResult struct {
	Data      string `json:"data,omitempty"`
	Path      string `json:"path,omitempty"`
	Extension string `json:"extension,omitempty"`
	Size      int    `json:"size,omitempty"`
	Error     string `json:"err,omitempty"`
	Version   int    `json:"version"`
}
