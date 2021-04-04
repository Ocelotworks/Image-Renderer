package stage

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"golang.org/x/image/draw"
	"image"
	"log"
)

func RotateAndResize(inputFrameCtx *gg.Context, outputCtx *gg.Context, component *entity.ImageComponent) {
	// move the specified component to its target position
	outputCtx.RotateAbout(component.Rotation, component.Position.X.(float64), component.Position.Y.(float64))

	// check if the frame needs to be resized
	var frameImage *image.RGBA
	if component.Position.Width != inputFrameCtx.Width() || component.Position.Height != inputFrameCtx.Height() {
		// make a rectangle with the target bounds
		newSize := image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: int(component.Position.Width.(float64)),
				Y: int(component.Position.Height.(float64)),
			},
		}

		log.Println("New size:", newSize)
		frameImage = image.NewRGBA(newSize)

		// scale the image
		draw.BiLinear.Scale(frameImage, frameImage.Bounds(), inputFrameCtx.Image(), inputFrameCtx.Image().Bounds(), draw.Over, &draw.Options{})
	} else {
		frameImage = inputFrameCtx.Image().(*image.RGBA)
	}

	log.Printf("Drawing component %s at %d %d\n", component.URL, component.Position.X, component.Position.Y)
	outputCtx.DrawImage(frameImage, int(component.Position.X.(float64)), int(component.Position.Y.(float64)))

	// Reset the rotation
	outputCtx.RotateAbout(-component.Rotation, component.Position.X.(float64), component.Position.Y.(float64))
}
