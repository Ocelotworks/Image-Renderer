package stage

import (
	"github.com/fogleman/gg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"golang.org/x/image/draw"
	"image"
	"log"
	"sync"
	"time"
)

var (
	frameDiffDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "frame_diff_duration",
		Help:      "Duration taken to diff completed frames",
	})
)

// The fully intact previous frame, used to determine what has changed in the next frame
var unmaskedPrevious image.Image

func GIFOptimise(outputCtx *gg.Context, frameNum int, wg *sync.WaitGroup) {
	// The image must be copied here, to avoid using the mask as the unmasked previous frame
	maskCopy := image.NewRGBA(outputCtx.Image().Bounds())
	draw.Copy(maskCopy, image.Point{X: 0, Y: 0}, outputCtx.Image(), outputCtx.Image().Bounds(), draw.Src, nil)
	if frameNum > 0 {
		wg.Add(1)
		go diffMask(outputCtx, unmaskedPrevious, wg, frameNum)
	}
	unmaskedPrevious = maskCopy
}

// Erases pixels on `context` that are different to those on `image2`
func diffMask(context *gg.Context, image2 image.Image, wg *sync.WaitGroup, num int) {
	defer wg.Done()
	diffMaskStart := time.Now()
	log.Printf("Diff for frame %d has finished", num)
	image1 := context.Image()
	dx := image1.Bounds().Dx()
	dy := image1.Bounds().Dy()
	context.SetRGBA255(0, 0, 0, 0)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			// This could probably be faster
			if helper.ColoursEqual(image1.At(x, y), image2.At(x, y)) {
				context.SetPixel(x, y)
			}
		}
	}

	frameDiffDuration.Observe(float64(time.Since(diffMaskStart).Milliseconds()))
}
