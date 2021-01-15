package filter

import (
	"fmt"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
	"image/color"
	"sync"
)

type Rainbow struct{}

func (r Rainbow) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {

	if len(*images) == 1 {
		for i := 0; i < 10; i++ {
			*images = append(*images, (*images)[0])
			*delays = append(*delays, 10)
		}
	}

	totalFrames := len(*images)
	outputImages := make([]*image.Image, totalFrames)
	for i, frame := range *images {
		outputImages[i] = processFrame(frame, i, totalFrames)
		helper.WriteDebugPNG(*outputImages[i], fmt.Sprintf("rainbow.%d", i))
	}

	*images = outputImages
	return
}

func processFrame(frame *image.Image, frameNum int, totalFrames int) *image.Image {
	palettedFrame, ok := (*frame).(*image.Paletted)

	if ok {
		return processPalettedFrame(palettedFrame, frameNum, totalFrames)
	}

	rgbaFrame, ok := (*frame).(*image.RGBA)

	if ok {
		return processRGBAFrame(rgbaFrame, frameNum, totalFrames)
	}

	nrgbaFrame, ok := (*frame).(*image.NRGBA)

	if ok {
		return processNRGBAFrame(nrgbaFrame, frameNum, totalFrames)
	}

	yCbCrFrame, ok := (*frame).(*image.YCbCr)

	if ok {
		return processYCbCrFrame(yCbCrFrame, frameNum, totalFrames)
	}

	fmt.Printf("Unknown image type: %T\n", *frame)

	return frame
}

func processNRGBAFrame(frame *image.NRGBA, frameNum int, totalFrames int) *image.Image {
	newImage := &image.NRGBA{
		Rect:   frame.Rect,
		Stride: frame.Stride,
	}
	r, g, b := hslToRgb(float64(frameNum)/float64(totalFrames), 1, 0.5)
	newImage.Pix = processPixArray(frame.Pix, r, g, b)
	castImage := image.Image(newImage)
	return &castImage
}

func processRGBAFrame(frame *image.RGBA, frameNum int, totalFrames int) *image.Image {
	newImage := &image.RGBA{
		Rect:   frame.Rect,
		Stride: frame.Stride,
	}
	r, g, b := hslToRgb(float64(frameNum)/float64(totalFrames), 1, 0.5)
	newImage.Pix = processPixArray(frame.Pix, r, g, b)
	castImage := image.Image(newImage)
	return &castImage
}

// Processes arrays of RGBA pixels
func processPixArray(pix []uint8, r uint8, g uint8, b uint8) []uint8 {
	rgb := []uint8{r, g, b}
	newPix := make([]uint8, len(pix))
	for i, pixel := range pix {
		value := (i + 1) % 4
		if value == 0 {
			newPix[i] = pix[i]
			continue
		}
		newPix[i] = blendColours(pixel, rgb[value-1])
	}
	return newPix
}

func processPalettedFrame(frame *image.Paletted, frameNum int, totalFrames int) *image.Image {
	r, g, b := hslToRgb(float64(frameNum)/float64(totalFrames), 1, 0.5)
	for i, colour := range frame.Palette {
		rgbaColour, ok := colour.(color.RGBA)
		if !ok {
			fmt.Printf("Unsupported palette type: %T\n", colour)
			continue
		}
		frame.Palette[i] = color.RGBA{
			R: blendColours(rgbaColour.R, r),
			G: blendColours(rgbaColour.G, g),
			B: blendColours(rgbaColour.B, b),
			A: rgbaColour.A,
		}
	}
	castImage := image.Image(frame)
	return &castImage
}

// Fuck this format
func processYCbCrFrame(frame *image.YCbCr, frameNum int, totalFrames int) *image.Image {
	r, g, b := hslToRgb(float64(frameNum)/float64(totalFrames), 1, 0.5)
	y, cb, cr := color.RGBToYCbCr(r, g, b)
	newImage := image.YCbCr{
		Y:              make([]uint8, len(frame.Y)),
		Cb:             make([]uint8, len(frame.Cb)),
		Cr:             make([]uint8, len(frame.Cr)),
		YStride:        frame.YStride,
		CStride:        frame.CStride,
		SubsampleRatio: frame.SubsampleRatio,
		Rect:           frame.Rect,
	}

	var wg sync.WaitGroup

	wg.Add(3)
	go processYCbCrArray(frame.Y, newImage.Y, y, &wg)
	go processYCbCrArray(frame.Cb, newImage.Cb, cb, &wg)
	go processYCbCrArray(frame.Cr, newImage.Cr, cr, &wg)

	wg.Wait()
	castFrame := image.Image(&newImage)
	return &castFrame
}

func processYCbCrArray(arr []uint8, out []uint8, blend uint8, wg *sync.WaitGroup) {
	defer wg.Done()
	for i, val := range arr {
		out[i] = blendColours(val, blend)
	}
}

func blendColours(colour1 uint8, colour2 uint8) uint8 {
	output := colour1/2 + colour2/3
	if output > 255 {
		return 255
	}
	return output
}

func hslToRgb(h float64, s float64, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		return uint8(l * 255), uint8(l * 255), uint8(l * 255)
	}

	var v1, v2 float64
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r := hueToRGB(v1, v2, h+(1.0/3.0))
	g := hueToRGB(v1, v2, h)
	b := hueToRGB(v1, v2, h-(1.0/3.0))

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
