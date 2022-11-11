package stage

import (
	"github.com/fogleman/gg"
	"image"
	"testing"
)

func BenchmarkDiffMask(b *testing.B) {

	blackWhiteImage := image.NewRGBA(image.Rect(0, 0, 800, 800))
	whiteBlackImage := image.NewRGBA(image.Rect(0, 0, 800, 800))

	for i := range blackWhiteImage.Pix {
		if i%2 == 0 {
			blackWhiteImage.Pix[i] = uint8(0xFF)
		} else {
			whiteBlackImage.Pix[i] = uint8(0xFF)
		}
	}

	for n := 0; n < b.N; n++ {
		blackWhiteContext := gg.NewContextForImage(blackWhiteImage)
		b.StartTimer()
		diffMask(blackWhiteContext, whiteBlackImage, nil, n)
		b.StopTimer()
	}
}

func BenchmarkDiffMaskRGBA(b *testing.B) {
	blackWhiteImage := image.NewRGBA(image.Rect(0, 0, 800, 800))
	whiteBlackImage := image.NewRGBA(image.Rect(0, 0, 800, 800))

	for i := range blackWhiteImage.Pix {
		if i%2 == 0 {
			blackWhiteImage.Pix[i] = uint8(0xFF)
		} else {
			whiteBlackImage.Pix[i] = uint8(0xFF)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		diffMaskRGBA(blackWhiteImage, whiteBlackImage, nil)
	}
}
