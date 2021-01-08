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
func OutputImage(input []image.Image, delay []int, metadata *entity.Metadata, frameDisposal bool) (string, string, int) {
	buf := new(bytes.Buffer)
	var format string
	stegMessage, exception := json.Marshal(metadata)
	if exception != nil {
		stegMessage = []byte("OCELOTBOT")
		log.Println("Failed to marshal metadata: ", exception)
	}
	if len(input) > 1 {
		images := make([]*image.Paletted, len(input))
		disposal := make([]byte, len(input))

		var wg sync.WaitGroup
		for frame, img := range input {
			wg.Add(1)
			go quantizeWorker(frame, img, &wg, images)
			if frameDisposal {
				disposal[frame] = gif.DisposalBackground
			} else {
				disposal[frame] = gif.DisposalNone
			}

		}

		wg.Wait()

		firstFrame := images[0]
		bounds := firstFrame.Bounds()
		config := image.Config{
			ColorModel: firstFrame.ColorModel(),
			Width:      bounds.Max.X,
			Height:     bounds.Max.Y,
		}

		output := gif.GIF{
			Image:           images,
			Delay:           delay,
			Disposal:        disposal,
			LoopCount:       0,
			BackgroundIndex: 0,
			Config:          config,
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
			_ = png.Encode(buf, input[0])
		} else {
			_, _ = stegoBuf.WriteTo(buf)
		}
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), format, buf.Len() / 1000000 //Megabytes
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
