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
		r, g, b := hslToRgb(float64(i)/float64(totalFrames), 1, 0.5)
		rgb := []uint8{r, g, b}
		y, cb, cr := color.RGBToYCbCr(r, g, b)
		yCbCr := []uint8{y, cb, cr}
		outputImages[i] = helper.ProcessFrame(frame, func(subPixel uint8, index int) uint8 {
			return blendColours(subPixel, rgb[(index+1)%4-1])
		}, func(palette color.Color, index int) color.Color {
			rgbaColour, ok := palette.(color.RGBA)
			if !ok {
				fmt.Printf("Unsupported palette type: %T\n", palette)
				return palette
			}
			return color.RGBA{
				R: blendColours(rgbaColour.R, r),
				G: blendColours(rgbaColour.G, g),
				B: blendColours(rgbaColour.B, b),
				A: rgbaColour.A,
			}
		}, func(arr []uint8, out []uint8, wg *sync.WaitGroup, index int) {
			defer wg.Done()
			for i, val := range arr {
				out[i] = blendColours(val, yCbCr[index])
			}
		})
		helper.WriteDebugPNG(*outputImages[i], fmt.Sprintf("rainbow.%d", i))
	}

	*images = outputImages
	return
}

func blendColours(colour1 uint8, colour2 uint8) uint8 {
	output := colour1/2 + colour2/3
	if output > 255 {
		return 255
	}
	return output
}

func hslToRgb(h float64, s float64, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		return uint8(l * 255), uint8(l * 255), uint8(l * 255)
	}

	var v1, v2 float64
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r := hueToRGB(v1, v2, h+(1.0/3.0))
	g := hueToRGB(v1, v2, h)
	b := hueToRGB(v1, v2, h-(1.0/3.0))

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
