package filter

import (
	"fmt"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
	"image/color"
	"sync"
)

type Rainbow struct{}

func (r Rainbow) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {

	if len(*images) == 1 {
		for i := 0; i < 10; i++ {
			*images = append(*images, (*images)[0])
			*delays = append(*delays, 10)
		}
	}

	totalFrames := len(*images)
	outputImages := make([]*image.Image, totalFrames)
	for i, frame := range *images {
		// Get the RGB value of the rainbow for this frame by getting the percentage progress through the gif
		r, g, b := helper.HslToRgb(float64(i)/float64(totalFrames), 1, 0.5)
		rgb := []uint8{r, g, b}
		// Convert that colour to YCbCr in case this is a YCbCr image
		y, cb, cr := color.RGBToYCbCr(r, g, b)
		yCbCr := []uint8{y, cb, cr}
		outputImages[i] = helper.ProcessFrame(frame, func(subPixel uint8, index int, newPix []uint8) uint8 {
			// For an RGB frame, we just need to repeat over the pixels in the RGB array we made earlier
			return helper.BlendColours(subPixel, rgb[(index+1)%4-1])
		}, func(palette color.Color, index int) color.Color {
			// Palette colours just need to be all blended together with teh RGB
			rgbaColour, ok := palette.(color.RGBA)
			if !ok {
				fmt.Printf("Unsupported palette type: %T\n", palette)
				return palette
			}
			return color.RGBA{
				R: helper.BlendColours(rgbaColour.R, r),
				G: helper.BlendColours(rgbaColour.G, g),
				B: helper.BlendColours(rgbaColour.B, b),
				A: rgbaColour.A,
			}
		}, func(arr []uint8, out []uint8, wg *sync.WaitGroup, index int) {
			defer wg.Done()
			for i, val := range arr {
				out[i] = helper.BlendColours(val, yCbCr[index])
			}
		})
		helper.WriteDebugPNG(*outputImages[i], fmt.Sprintf("rainbow.%d", i))
	}

	*images = outputImages
	return
}
