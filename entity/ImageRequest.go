package entity

type ImageRequest struct {
	ImageComponents []*ImageComponent `json:"components"`
	Metadata        *Metadata         `json:"metadata"`
	Width           int               `json:"width"`
	Height          int               `json:"height"`
}
