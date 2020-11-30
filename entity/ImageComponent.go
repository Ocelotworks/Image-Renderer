package entity

type ImageComponent struct {
	Url      string `json:"url"`
	Local    bool   `json:"local"`
	Position struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"w"`
		Height int `json:"h"`
	} `json:"pos"`

	Rotation float64   `json:"rot"`
	Filters  []*Filter `json:"filter"`
}
