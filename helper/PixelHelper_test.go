package helper

import (
	"image"
	"image/png"
	"os"
	"testing"
)

func Benchmark_PixelHelperNoop(b *testing.B) {
	file, _ := os.Open("../res/hospital.png")

	img, _ := png.Decode(file)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ForEveryPixel(img, func(x int, y int, r uint8, b uint8, g uint8, a uint8) {
			// noop
		})
	}
}

func Benchmark_PixelHelperSetPixel(b *testing.B) {
	file, _ := os.Open("../res/hospital.png")

	img, _ := png.Decode(file)

	rgbaImage := img.(*image.RGBA)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ForEveryPixel(img, func(x int, y int, r uint8, b uint8, g uint8, a uint8) {
			i := rgbaImage.PixOffset(x, y)
			s := rgbaImage.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
			s[0] = 0
			s[1] = 0
			s[2] = 0
			s[3] = 0

		})
	}
}
