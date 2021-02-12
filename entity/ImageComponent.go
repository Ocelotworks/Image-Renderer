package entity

// ImageComponent describes an image component in a request
type ImageComponent struct {
	URL      string   `json:"url"`
	Local    bool     `json:"local"`
	Position Position `json:"pos"`

	Rotation   float64   `json:"rot"`
	Filters    []*Filter `json:"filter"`
	Background string    `json:"background"`
}
