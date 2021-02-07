package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/auyer/steganography"
	q "github.com/ericpauley/go-quantize/quantize"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"sync"
	"time"
)

var (
	pngEncode = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "output_png_encode",
		Help:      "Duration taken to encode a PNG",
	})
	gifEncode = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "output_gif_encode",
		Help:      "Duration taken to encode a GIF",
	})
	compress = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "output_compress",
		Help:      "Duration taken to compress the image",
	})
)

// OutputImage outputs an image as a byte array and file extension combination
func OutputImage(input []image.Image, delay []int, metadata *entity.Metadata, frameDisposal bool, compression bool) (string, string, int, error) {
	buf := new(bytes.Buffer)
	var format string
	stegMessage, exception := json.Marshal(metadata)
	if exception != nil {
		stegMessage = []byte("OCELOTBOT")
		sentry.CaptureException(exception)
		log.Println("Failed to marshal metadata: ", exception)
	}
	if len(input) > 1 {
		gifEncodeStart := time.Now()
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

		exception = gif.EncodeAll(buf, &output)
		if exception != nil {
			return "", "", 0, exception
		}
		format = "gif"
		gifEncode.Observe(float64(time.Since(gifEncodeStart).Milliseconds()))
	} else if len(input) == 1 {
		format = "png"
		pngEncodeStart := time.Now()
		stegoBuf := new(bytes.Buffer)
		exception = steganography.Encode(stegoBuf, input[0], stegMessage)

		if exception != nil {
			sentry.CaptureException(exception)
			log.Println("Unable to encode message: ", exception)
			exception = png.Encode(buf, input[0])
		} else {

			_, exception := stegoBuf.WriteTo(buf)
			if exception != nil {
				sentry.CaptureException(exception)
				log.Println("Unable to write steg message: ", exception)
				exception = png.Encode(buf, input[0])
			}
		}

		pngEncode.Observe(float64(time.Since(pngEncodeStart).Milliseconds()))
		if exception != nil {
			return "", "", 0, exception
		}
	}

	originalLength := buf.Len() / 1000000

	if compression {
		compressionStart := time.Now()
		var compressedBuf bytes.Buffer
		gz := gzip.NewWriter(&compressedBuf)
		gz.Name = "output." + format
		_, exception = gz.Write(buf.Bytes())

		if exception == nil && gz.Close() == nil {
			compress.Observe(float64(time.Since(compressionStart).Milliseconds()))
			return base64.StdEncoding.EncodeToString(compressedBuf.Bytes()), "gzip/" + format, originalLength, nil
		}
		fmt.Println("failed to compress: ", exception)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), format, originalLength, nil //Megabytes

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
