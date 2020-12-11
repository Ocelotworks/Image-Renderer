package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"sync"

	"github.com/auyer/steganography"
	q "github.com/ericpauley/go-quantize/quantize"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
)

// OutputImage outputs an image as a byte array and file extension combination
func OutputImage(input []image.Image, metadata *entity.Metadata) (string, string) {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	var format string
	stegMessage, exception := json.Marshal(metadata)
	if exception != nil {
		stegMessage = []byte("OCELOTBOT")
		log.Println("Failed to marshal metadata: ", exception)
	}
	if len(input) > 1 {
		images := make([]*image.Paletted, len(input))
		delay := make([]int, len(input))
		disposal := make([]byte, len(input))

		var wg sync.WaitGroup
		for frame, img := range input {
			wg.Add(1)
			go quantizeWorker(frame, img, &wg, images)

			// TODO: probably dont hardcode this
			delay[frame] = 10
			disposal[frame] = gif.DisposalPrevious
		}

		wg.Wait()

		output := gif.GIF{
			Image:           images,
			Delay:           delay,
			LoopCount:       0,
			Disposal:        disposal,
			BackgroundIndex: 0,
		}
		_ = gif.EncodeAll(buf, &output)
		format = "gif"
	} else if len(input) == 1 {
		fmt.Println("Output")
		format = "png"
		stegoBuf := new(bytes.Buffer)
		exception = steganography.Encode(stegoBuf, input[0], stegMessage)

		if exception != nil {
			log.Println("Unable to encode message: ", exception)
			_ = png.Encode(encoder, input[0])
		} else {
			_, _ = stegoBuf.WriteTo(encoder)
		}
	}
	return buf.String(), format
}

func quantizeWorker(frameNum int, img image.Image, wg *sync.WaitGroup, output []*image.Paletted) {
	defer wg.Done()

	log.Printf("Quantizing frame %d...", frameNum)

	// quantize the frame to a paletted image
	quantizer := q.MedianCutQuantizer{AddTransparent: true}
	qPalette := quantizer.Quantize(make([]color.Color, 0, 256), img)
	palettedImage := image.NewPaletted(img.Bounds(), qPalette)
	draw.Draw(palettedImage, img.Bounds(), img, image.Point{X: 0, Y: 0}, draw.Src)

	output[frameNum] = palettedImage
}
