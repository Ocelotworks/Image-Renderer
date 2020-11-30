package entity

type ImageRequest struct {
	ImageComponents []*ImageComponent `json:"components"`
	Width           int               `json:"width"`
	Height          int               `json:"height"`
}
