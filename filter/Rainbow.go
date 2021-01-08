package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
)

type Rainbow struct{}

func (r Rainbow) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {

	if len(*images) == 1 {
		for i := 0; i < 10; i++ {
			*images = append(*images, (*images)[0])
			*delays = append(*delays, 10)
		}
	}

	outputImages := make([]*image.Image, len(*images))
	for i, frame := range *images {
		ctx := gg.NewContextForImage(*frame)
		r, g, b := hslToRgb(float64(i)/float64(len(*images)), 1, 0.5)
		for x := 0; x < (*frame).Bounds().Dx(); x++ {
			for y := 0; y < (*frame).Bounds().Dy(); y++ {
				r1, b1, g1, a1 := (*frame).At(x, y).RGBA()
				if a1 > 0 {
					ctx.SetRGBA255(blendColours(r1, r), blendColours(g1, g), blendColours(b1, b), to8Bit(a1))
					ctx.SetPixel(x, y)
				}
			}
		}
		updatedFrame := ctx.Image()
		outputImages[i] = &updatedFrame
	}
	*images = outputImages
	return
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
