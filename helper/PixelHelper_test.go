package helper

import (
	"image/color"
	"testing"
)

func BenchmarkColoursEqualTrue(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ColoursEqual(color.White, color.White)
	}
}

func BenchmarkColoursEqualFalse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ColoursEqual(color.White, color.Black)
	}
}

func BenchmarkBlendColours(b *testing.B) {
	for n := 0; n < b.N; n++ {
		BlendColours(uint8(n), uint8(n*2))
	}
}

func BenchmarkHslToRgb(b *testing.B) {
	for n := 0; n < b.N; n++ {
		HslToRgb(float64(n), float64(n), float64(n))
	}
}
