package helper

import (
	"fmt"
	"image"
	"image/color"
	"sync"
)

func ProcessFrame(frame *image.Image,
	rgbCallback func(subPixel uint8, index int) uint8,
	palettedCallback func(palette color.Color, index int) color.Color,
	ycbcrCallback func(arr []uint8, out []uint8, wg *sync.WaitGroup, index int)) *image.Image {
	switch (*frame).(type) {
	case *image.Paletted:
		return ProcessPalettedFrame((*frame).(*image.Paletted), palettedCallback)
	case *image.RGBA:
		return ProcessRGBAFrame((*frame).(*image.RGBA), rgbCallback)
	case *image.NRGBA:
		return ProcessNRGBAFrame((*frame).(*image.NRGBA), rgbCallback)
	case *image.YCbCr:
		return ProcessYCbCrFrame((*frame).(*image.YCbCr), ycbcrCallback)
	default:
		fmt.Printf("Unknown image type: %T\n", *frame)
		return frame
	}
}

func ProcessNRGBAFrame(frame *image.NRGBA, callback func(subPixel uint8, index int) uint8) *image.Image {
	newImage := &image.NRGBA{
		Rect:   frame.Rect,
		Stride: frame.Stride,
	}
	newImage.Pix = ProcessPixArray(frame.Pix, callback)
	castImage := image.Image(newImage)
	return &castImage
}

func ProcessRGBAFrame(frame *image.RGBA, callback func(subPixel uint8, index int) uint8) *image.Image {
	newImage := &image.RGBA{
		Rect:   frame.Rect,
		Stride: frame.Stride,
	}
	newImage.Pix = ProcessPixArray(frame.Pix, callback)
	castImage := image.Image(newImage)
	return &castImage
}

// Processes arrays of RGBA pixels
func ProcessPixArray(pix []uint8, callback func(subPixel uint8, index int) uint8) []uint8 {
	newPix := make([]uint8, len(pix))
	for i, subPixel := range pix {
		value := (i + 1) % 4
		if value == 0 {
			newPix[i] = pix[i]
			continue
		}
		newPix[i] = callback(subPixel, i)
	}
	return newPix
}

func ProcessPalettedFrame(frame *image.Paletted, callback func(palette color.Color, index int) color.Color) *image.Image {
	for i, colour := range frame.Palette {
		frame.Palette[i] = callback(colour, i)
	}
	castImage := image.Image(frame)
	return &castImage
}

// Fuck this format
func ProcessYCbCrFrame(frame *image.YCbCr, callback func(arr []uint8, out []uint8, wg *sync.WaitGroup, index int)) *image.Image {
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
	go callback(frame.Y, newImage.Y, &wg, 0)
	go callback(frame.Cb, newImage.Cb, &wg, 1)
	go callback(frame.Cr, newImage.Cr, &wg, 2)

	wg.Wait()
	castFrame := image.Image(&newImage)
	return &castFrame
}

// Returns if two colours are equal
func ColoursEqual(colour1 color.Color, colour2 color.Color) bool {
	r1, b1, g1, a1 := colour1.RGBA()
	r2, b2, g2, a2 := colour2.RGBA()
	return r1 == r2 && b1 == b2 && g1 == g2 && a1 == a2
}
