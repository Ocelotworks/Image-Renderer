package filter

import (
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
	"math"
)

type Hyper struct{}

func (r Hyper) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {

	outputDelays := make([]int, len(*delays))
	for i, delay := range *delays {
		// GIF rendering ignores delays of 1 and instead uses default. Wooo gifs
		if delay > 10 {
			outputDelays[i] = int(math.Min(float64(delay/2), float64(10)))
		}
	}
	*delays = outputDelays
}
