package helper

import (
	"github.com/fogleman/gg"
	"image"
)

func ForEveryPixel(img image.Image, iterator func(x int, y int, ctx *gg.Context, r uint32, b uint32, g uint32, a uint32)) image.Image {
	ctx := gg.NewContextForImage(img)
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dx(); y++ {
			r, b, g, a := img.At(x, y).RGBA()
			iterator(x, y, ctx, r, g, b, a)
		}
	}
	return ctx.Image()
}
