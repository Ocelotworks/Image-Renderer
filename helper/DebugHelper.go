package helper

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
)

func WriteDebugPNG(image image.Image, name string) {
	if os.Getenv("DEBUG_IMAGES") == "1" {

		file, exception := os.Create(path.Join("debug", name+".png"))

		if exception != nil {
			fmt.Print("Failed to open debug PNG for writing: ", exception)
			return
		}

		defer file.Close()

		exception = png.Encode(file, image)

		if exception != nil {
			fmt.Println("Failed to write debug PNG: ", exception)
		}
	}
}
