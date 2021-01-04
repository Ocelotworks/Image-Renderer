package main

import (
	"github.com/fogleman/gg"
	"image"
	"image/png"
	"os"
	"path"
	"testing"
)

func BenchmarkProcessImage(b *testing.B) {
	file, exception := os.Open(path.Join("test_res", "stacking.gif"))
	if exception != nil {
		b.Fatal(exception)
	}
	defer file.Close()

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, _, _ = getImage(file)
	}
	b.StopTimer()
}

func BenchmarkDiffFrames(b *testing.B) {
	diff1, exception := loadTestPNG("diff_1.png")

	if exception != nil {
		b.Fatal(exception)
	}

	diff2, exception := loadTestPNG("diff_2.png")

	if exception != nil {
		b.Fatal(exception)
	}

	ctx := gg.NewContextForImage(diff1)

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		diffMask(ctx, diff2)
	}
	b.StopTimer()
}

func loadTestPNG(file string) (image.Image, error) {
	infile, exception := os.Open(path.Join("test_res", file))
	if exception != nil {
		return nil, exception
	}
	defer infile.Close()

	img, exception := png.Decode(infile)

	if exception != nil {
		return nil, exception
	}

	return img, nil
}
