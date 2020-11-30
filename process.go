package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
	"golang.org/x/image/draw"
	"image"
	"log"
	"net/http"
)

var filters = map[string]filter.Filter{
	"rectangle": filter.Rectangle{},
}

func ProcessImage(request *entity.ImageRequest) {
	canvas := gg.NewContext(request.Width, request.Height)
	for _, component := range request.ImageComponents {
		var img *image.Image
		var exception error
		if !component.Local {
			img, exception = getImageUrl(component.Url)
		} else {
			// TODO: local images
		}

		if exception != nil {
			log.Println(exception)
			continue
		}
		imageContext := gg.NewContextForImage(*img)

		for _, filterObject := range component.Filters {
			if filters[filterObject.Name] != nil {
				log.Println("Applying filter ", filterObject.Name, filterObject.Arguments)
				filters[filterObject.Name].ApplyFilter(imageContext, filterObject.Arguments)
			} else {
				log.Println("Unknown filter type ", filterObject)
			}
		}

		canvas.RotateAbout(component.Rotation, float64(component.Position.X), float64(component.Position.Y))
		if component.Position.Width == 0 {
			component.Position.Width = imageContext.Width()
		}
		if component.Position.Height == 0 {
			component.Position.Height = imageContext.Height()
		}
		var dstImage *image.RGBA
		if component.Position.Width != imageContext.Width() || component.Position.Height != imageContext.Height() {
			newSize := image.Rectangle{
				Min: image.Point{
					X: component.Position.X,
					Y: component.Position.Y,
				},
				Max: image.Point{
					X: component.Position.X + component.Position.Width,
					Y: component.Position.Y + component.Position.Height,
				},
			}
			dstImage = image.NewRGBA(newSize)
			draw.BiLinear.Scale(dstImage, dstImage.Bounds(), imageContext.Image(), imageContext.Image().Bounds(), draw.Over, &draw.Options{})
		} else {
			dstImage = imageContext.Image().(*image.RGBA)
		}

		canvas.DrawImage(dstImage, component.Position.X, component.Position.X)
		// Reset the rotation
		canvas.RotateAbout(-component.Rotation, float64(component.Position.X), float64(component.Position.Y))
	}
	fmt.Println(OutputImage([]image.Image{canvas.Image()}))
}

func getImageUrl(url string) (*image.Image, error) {
	response, exception := http.Get(url)
	if exception != nil {
		return nil, exception
	}
	defer response.Body.Close()
	img, _, exception := image.Decode(response.Body)
	if exception != nil {
		return nil, exception
	}
	return &img, nil
}
