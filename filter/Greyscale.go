package filter

import (
	"fmt"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
	"image/color"
	"sync"
)

type Greyscale struct{}

func (r Greyscale) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {
	totalFrames := len(*images)
	outputImages := make([]*image.Image, totalFrames)
	for i, frame := range *images {
		// Average out each subpixel here and reset it when it reaches 3 count, simples
		colourTotal := 0
		colourCount := 0
		outputImages[i] = helper.ProcessFrame(frame, func(subPixel uint8, index int, newPix []uint8) uint8 {
			colourCount++
			colourTotal += int(subPixel)
			if colourCount < 3 {
				return 0
			}
			colour := colourTotal / colourCount
			colourCount = 0
			colourTotal = 0
			newPix[index-1] = uint8(colour)
			newPix[index-2] = uint8(colour)
			return uint8(colour)
		}, func(palette color.Color, index int) color.Color {
			rgbaColour, ok := palette.(color.RGBA)
			if !ok {
				return palette
			}
			colour := uint8((int(rgbaColour.R) + int(rgbaColour.G) + int(rgbaColour.B)) / 3)
			return color.RGBA{
				R: colour,
				G: colour,
				B: colour,
				A: rgbaColour.A,
			}
		}, func(arr []uint8, out []uint8, wg *sync.WaitGroup, index int) {
			defer wg.Done()
			if index == 0 {
				copy(out, arr)
				return
			}
			for i := range out {
				out[i] = 128
			}
		})
		helper.WriteDebugPNG(*outputImages[i], fmt.Sprintf("greyscale.%d", i))
	}

	*images = outputImages
	return
}
