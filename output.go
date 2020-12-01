package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/auyer/steganography"
	q "github.com/ericpauley/go-quantize/quantize"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
)

func OutputImage(input []image.Image, metadata *entity.Metadata) (string, string) {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	f, _ := os.Create("output.png")
	defer f.Close()
	var format string
	stegMessage, exception := json.Marshal(metadata)
	if exception != nil {
		stegMessage = []byte("OCELOTBOT")
		log.Println("Failed to marshal metadata: ", exception)
	}
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
		format = "png"
		buf := new(bytes.Buffer)
		exception = steganography.Encode(buf, input[0], stegMessage)

		if exception != nil {
			log.Println("Unable to encode message: ", exception)
			_ = png.Encode(encoder, input[0])
		} else {
			_, _ = buf.WriteTo(encoder)
		}
		// _ = png.Encode(encoder, input[0])
	}
	return buf.String(), format
}
