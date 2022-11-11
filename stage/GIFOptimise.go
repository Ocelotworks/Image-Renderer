package stage

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"golang.org/x/image/draw"
	"image"
	"sync"
)

// The fully intact previous frame, used to determine what has changed in the next frame
var unmaskedPrevious image.Image

func GIFOptimise(outputCtx *gg.Context, frameNum int, wg *sync.WaitGroup) {
	// The image must be copied here, to avoid using the mask as the unmasked previous frame
	maskCopy := image.NewRGBA(outputCtx.Image().Bounds())
	draw.Copy(maskCopy, image.Point{X: 0, Y: 0}, outputCtx.Image(), outputCtx.Image().Bounds(), draw.Src, nil)
	if frameNum > 0 {
		wg.Add(1)
		go diffMaskRGBA(maskCopy, unmaskedPrevious.(*image.RGBA), wg)
		//go diffMask(outputCtx, unmaskedPrevious, wg, frameNum)
	}
	unmaskedPrevious = maskCopy
}

// Erases pixels on `context` that are different to those on `image2`
func diffMask(context *gg.Context, image2 image.Image, wg *sync.WaitGroup, num int) {
	if wg != nil {
		defer wg.Done()
	}
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
}

func diffMaskRGBA(image1 *image.RGBA, image2 *image.RGBA, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	imgLength := len(image1.Pix) - 3

	for i := 0; i < imgLength; i += 3 {
		if image1.Pix[i] != image2.Pix[i] || image1.Pix[i+1] != image2.Pix[i+1] || image1.Pix[i+2] != image2.Pix[i+2] {
			image1.Pix[i] = 0x00
			image1.Pix[i+1] = 0x00
			image1.Pix[i+2] = 0x00
		}
	}
}
