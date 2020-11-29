package main

import (
	"fmt"
	q "github.com/ericpauley/go-quantize/quantize"
	"github.com/fogleman/gg"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
)

func main() {
	frames := 2
	images := make([]*image.Paletted, frames)
	delay := make([]int, frames)
	quantizer := q.MedianCutQuantizer{}

	for x := 0; x < frames; x++ {
		const S = 1000
		dc := gg.NewContext(S, S)
		dc.DrawCircle(500, 500, 400)
		dc.SetRGB(0.5, float64(x)/float64(frames), float64(x)/float64(frames))
		dc.Fill()
		originalImage := dc.Image()
		qPalette := quantizer.Quantize(make([]color.Color, 0, 256), originalImage)
		palettedImage := image.NewPaletted(originalImage.Bounds(), qPalette)

		draw.Draw(palettedImage, originalImage.Bounds(), originalImage, image.Point{
			X: 0,
			Y: 0,
		}, draw.Src)
		images[x] = palettedImage
		delay[x] = 100
		fmt.Println("Drawing frame ", x)
	}

	output := gif.GIF{
		Image:           images,
		Delay:           delay,
		LoopCount:       0,
		BackgroundIndex: 0,
	}

	file, _ := os.Create("output.gif")
	fmt.Println("Encoding output")

	exception := gif.EncodeAll(file, &output)
	if exception != nil {
		fmt.Println(exception)
	}
	defer file.Close()
}
