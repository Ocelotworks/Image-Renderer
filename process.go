package main

import (
	"bytes"
	"fmt"
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
	"golang.org/x/image/draw"
	"image"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

var filters = map[string]filter.Filter{
	"rectangle": filter.Rectangle{},
}

func ProcessImage(request *entity.ImageRequest) *entity.ImageResult {
	canvases := make([]*gg.Context, 1)
	canvases[0] = gg.NewContext(request.Width, request.Height)
	for _, component := range request.ImageComponents {
		var imageContexts []*gg.Context
		if len(component.Url) > 0 {
			var imgs []*image.Image
			var exception error
			if !component.Local {
				imgs, exception = getImageUrl(component.Url)
			} else {
				imgs, exception = getLocalImage(component.Url)
			}

			if exception != nil {
				log.Println(exception)
				continue
			}
			imageContexts = make([]*gg.Context, len(imgs))
			for i, img := range imgs {
				imageContexts[i] = gg.NewContextForImage(*img)
			}
		} else {
			imageContexts = []*gg.Context{gg.NewContext(request.Width, request.Height)}
		}
		for frame, imageContext := range imageContexts {
			for _, filterObject := range component.Filters {
				if filters[filterObject.Name] != nil {
					log.Println("Applying filter ", filterObject.Name, filterObject.Arguments)
					filters[filterObject.Name].ApplyFilter(imageContext, filterObject.Arguments)
				} else {
					log.Println("Unknown filter type ", filterObject)
				}
			}
			if len(canvases)-1 < frame {
				canvases = append(canvases, gg.NewContext(request.Width, request.Height))
			}
			canvases[frame].RotateAbout(component.Rotation, float64(component.Position.X), float64(component.Position.Y))
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
						X: 0,
						Y: 0,
					},
					Max: image.Point{
						X: component.Position.Width,
						Y: component.Position.Height,
					},
				}
				log.Println("New size: ", newSize)
				dstImage = image.NewRGBA(newSize)

				draw.BiLinear.Scale(dstImage, dstImage.Bounds(), imageContext.Image(), imageContext.Image().Bounds(), draw.Over, &draw.Options{})
			} else {
				dstImage = imageContext.Image().(*image.RGBA)
			}

			log.Println("Drawing image at ", component.Position.X, component.Position.Y)
			canvases[frame].DrawImage(dstImage, component.Position.X, component.Position.Y)
			// Reset the rotation
			canvases[frame].RotateAbout(-component.Rotation, float64(component.Position.X), float64(component.Position.Y))
		}
	}
	outputImages := make([]image.Image, len(canvases))
	for i, canvas := range canvases {
		outputImages[i] = canvas.Image()
	}
	result, extension := OutputImage(outputImages, request.Metadata)
	return &entity.ImageResult{
		Data:      result,
		Extension: extension,
	}
}

func getImageUrl(url string) ([]*image.Image, error) {
	response, exception := http.Get(url)
	if exception != nil {
		return nil, exception
	}
	defer response.Body.Close()
	return getImage(response.Body)
}

func getLocalImage(url string) ([]*image.Image, error) {
	file, exception := os.Open(path.Join("res", url))
	if exception != nil {
		return nil, exception
	}
	defer file.Close()
	return getImage(file)
}

func getImage(input io.Reader) ([]*image.Image, error) {
	body, exception := ioutil.ReadAll(input)
	if exception != nil {
		return nil, exception
	}
	reader := bytes.NewReader(body)
	_, format, exception := image.DecodeConfig(reader)
	if exception != nil {
		return nil, exception
	}
	_, exception = reader.Seek(0, 0)
	if exception != nil {
		return nil, exception
	}
	if format == "gif" {
		fmt.Println("Decode the gif")
		gifFile, exception := gif.DecodeAll(reader)
		if exception != nil {
			return nil, exception
		}
		output := make([]*image.Image, len(gifFile.Image))
		for i, img := range gifFile.Image {
			genericImage := image.Image(img)
			output[i] = &genericImage
		}
		return output, nil
	} else {
		img, _, exception := image.Decode(reader)
		if exception != nil {
			return nil, exception
		}
		return []*image.Image{&img}, nil
	}
}
