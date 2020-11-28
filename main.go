package main

import (
	"github.com/fogleman/gg"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
)

func main() {
	images := make([]*image.Paletted, 2)
	delay := make([]int, 2)

	for x := 0; x < 2; x++ {
		const S = 1024
		dc := gg.NewContext(S, S)
		dc.SetRGBA(float64(x*100), 0, 0, 1)
		originalImage := dc.Image()
		palettedImage := image.NewPaletted(originalImage.Bounds(), palette.Plan9)
		draw.Draw(palettedImage, originalImage.Bounds(), originalImage, originalImage.Bounds().Min, draw.Over)
		images[x] = palettedImage
		delay[x] = 1
	}

	output := gif.GIF{
		Image:     images,
		Delay:     delay,
		LoopCount: -1,
	}

	file, _ := os.Create("output.gif")

	gif.EncodeAll(file, &output)

}
