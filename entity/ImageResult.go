package entity

// ImageResult is a resulting rendered image
type ImageResult struct {
	Data      string `json:"data"`
	Extension string `json:"extension"`
	Size      int    `json:"size"`
}
