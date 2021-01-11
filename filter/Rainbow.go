package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
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

	var wg sync.WaitGroup
	totalFrames := len(*images)
	outputImages := make([]*image.Image, totalFrames)
	for i, frame := range *images {
		wg.Add(1)
		go processFrame(frame, i, totalFrames, outputImages, &wg)

	}
	wg.Wait()
	*images = outputImages
	return
}

func processFrame(frame *image.Image, frameNum int, totalFrames int, outputImages []*image.Image, wg *sync.WaitGroup) {
	defer wg.Done()
	r, g, b := hslToRgb(float64(frameNum)/float64(totalFrames), 1, 0.5)
	updatedFrame := helper.ForEveryPixel(*frame, func(x int, y int, ctx *gg.Context, r1 uint32, b1 uint32, g1 uint32, a1 uint32) {
		if a1 > 0 {
			ctx.SetRGBA255(blendColours(r1, r), blendColours(g1, g), blendColours(b1, b), to8Bit(a1))
			ctx.SetPixel(x, y)
		}
	})
	outputImages[frameNum] = &updatedFrame
}

func blendColours(colour1 uint32, colour2 int) int {
	output := (to8Bit(colour1) + (colour2 / 3)) / 2
	if output > 255 {
		return 255
	}
	return output
}

func to8Bit(input uint32) int {
	// :sadcat:
	return int(uint8(input))
}

func hslToRgb(h float64, s float64, l float64) (int, int, int) {
	if s == 0 {
		// it's gray
		return int(l * 255), int(l * 255), int(l * 255)
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

	return int(r * 255), int(g * 255), int(b * 255)
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
