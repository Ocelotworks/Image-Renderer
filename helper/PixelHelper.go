package helper

import (
	"image"
)

func ForEveryPixel(img image.Image, iterator func(x int, y int, r uint8, b uint8, g uint8, a uint8)) image.Image {
	palettedImage, ok := img.(*image.Paletted)
	if ok {
		return forEveryPalettedPixel(palettedImage, iterator)
	}

	return img

	//dx := img.Bounds().Dx()
	//dy := img.Bounds().Dy()
	//for x := 0; x < dx; x++ {
	//	for y := 0; y < dy; y++ {
	//		i := rgbaImage.PixOffset(x, y)
	//		pixel := rgbaImage.Pix[i : i+4 : i+4]
	//		iterator(x, y, pixel[0], pixel[1], pixel[2], pixel[3])
	//	}
	//}
	//return img
}

func forEveryPalettedPixel(img *image.Paletted, iterator func(x int, y int, r uint8, b uint8, g uint8, a uint8)) image.Image {

	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			i := img.PixOffset(x, y)
			pixel := img.Pix[i : i+4 : i+4]
			iterator(x, y, pixel[0], pixel[1], pixel[2], pixel[3])
		}
	}
	return img
}
