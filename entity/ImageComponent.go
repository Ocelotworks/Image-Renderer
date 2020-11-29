package entity

import "image"

type ImageComponent struct {
	Url      string      `json:"url"`
	Local    bool        `json:"local"`
	Position image.Point `json:"pos"`
	Rotation int         `json:"rot"`
	Filters  []*Filter   `json:"filter"`
}
