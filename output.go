package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	q "github.com/ericpauley/go-quantize/quantize"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
)

func OutputImage(input []image.Image) (string, string) {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	f, _ := os.Create("output.png")
	var format string
	if len(input) > 1 {
		images := make([]*image.Paletted, len(input))
		delay := make([]int, len(input))
		quantizer := q.MedianCutQuantizer{}
		for frame, img := range input {
			qPalette := quantizer.Quantize(make([]color.Color, 0, 256), img)
			palettedImage := image.NewPaletted(img.Bounds(), qPalette)

			draw.Draw(palettedImage, img.Bounds(), img, image.Point{
				X: 0,
				Y: 0,
			}, draw.Src)
			images[frame] = palettedImage
			delay[frame] = 100
		}
		output := gif.GIF{
			Image:           images,
			Delay:           delay,
			LoopCount:       0,
			BackgroundIndex: 0,
		}
		_ = gif.EncodeAll(encoder, &output)
		format = "gif"
	} else if len(input) == 1 {
		fmt.Println("Output")
		_ = png.Encode(f, input[0])
	}
	return buf.String(), format
}
