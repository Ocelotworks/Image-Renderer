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
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
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
func OutputImage(input []image.Image, delay []int, frameDisposal bool, request *entity.ImageRequest) *entity.ImageResult {
	buf := new(bytes.Buffer)
	var format string
	stegMessage, exception := json.Marshal(request.Metadata)
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
		log.Println("Finished Quantizing")

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
			return &entity.ImageResult{Error: "image_encode"}
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
			return &entity.ImageResult{Error: "image_encode"}
		}
	}

	if request.Version >= 1 {
		hostname, _ := os.Hostname()
		fileName := fmt.Sprintf("%d.%s", time.Now().Unix(), format)
		exception := ioutil.WriteFile("output/"+fileName, buf.Bytes(), os.ModePerm)
		if exception != nil {
			return &entity.ImageResult{Error: "write_error"}
		}
		// Mind your FUCKING business
		//goland:noinspection HttpUrlsUsage
		return &entity.ImageResult{
			Path:      fmt.Sprintf("http://%s:2112/output/%s", hostname, fileName),
			Extension: format,
		}
	}

	originalLength := buf.Len() / 1000000

	if request.Compression {
		log.Println("Compressing output...")
		compressionStart := time.Now()
		var compressedBuf bytes.Buffer
		gz := gzip.NewWriter(&compressedBuf)
		gz.Name = "output." + format
		_, exception = gz.Write(buf.Bytes())

		if exception == nil && gz.Close() == nil {
			compress.Observe(float64(time.Since(compressionStart).Milliseconds()))
			log.Println("Finished Compressing")
			return &entity.ImageResult{
				Data:      base64.StdEncoding.EncodeToString(compressedBuf.Bytes()),
				Extension: "gzip/" + format,
				Size:      originalLength,
			}
		}
		fmt.Println("failed to compress: ", exception)
	}

	return &entity.ImageResult{
		Data:      base64.StdEncoding.EncodeToString(buf.Bytes()),
		Extension: format,
		Size:      originalLength,
	}
}

func quantizeWorker(frameNum int, img image.Image, wg *sync.WaitGroup, output []*image.Paletted) {
	defer wg.Done()

	rgbaImage := img.(*image.RGBA)

	log.Printf("Quantizing frame %d...", frameNum)

	// quantize the frame to a paletted image
	quantizer := q.MedianCutQuantizer{AddTransparent: true}
	qPalette := quantizer.Quantize(make([]color.Color, 0, 256), img)

	palettedImage := image.NewPaletted(img.Bounds(), qPalette)

	// Convert the RGBA image into a PNG
	for i := range rgbaImage.Pix {
		// Grab the first 4
		if i%4 != 0 {
			continue
		}
		palettedImage.Pix[i/4] = uint8(qPalette.Index(color.RGBA{
			R: rgbaImage.Pix[i],
			G: rgbaImage.Pix[i+1],
			B: rgbaImage.Pix[i+2],
			A: rgbaImage.Pix[i+3],
		}))
	}

	output[frameNum] = palettedImage
}
