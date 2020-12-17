package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
	"golang.org/x/image/draw"
)

var filters = map[string]filter.Filter{
	"rectangle": filter.Rectangle{},
}

// ProcessImage processes an incoming ImageRequest and outputs a finished ImageResult
func ProcessImage(request *entity.ImageRequest) *entity.ImageResult {
	// holds all the canvases for each frame of the final output image
	var outputCanvases []*gg.Context

	// loop over each image component
	for _, component := range request.ImageComponents {
		var imageContexts []*gg.Context
		// fetch the image by URL/path if provided, otherwise make a blank image context
		if len(component.Url) > 0 {
			// decide which function to get the image with (explicitly typed)
			var getImageFunc func(url string) ([]*image.Image, error) = getImageURL
			if component.Local {
				getImageFunc = getLocalImage
			}

			// get the image, returns all the frames if the image is a gif
			imgs, err := getImageFunc(component.Url)
			if err != nil {
				log.Println(err)
				continue
			}

			// create an image context for the image (or each frame for a gif)
			imageContexts = make([]*gg.Context, len(imgs))
			for i, img := range imgs {
				imageContexts[i] = gg.NewContextForImage(*img)
			}
		} else {
			imageContexts = []*gg.Context{gg.NewContext(request.Width, request.Height)}
		}

		totalFrames := max(len(imageContexts), len(outputCanvases))

		// get the image context for each frame (only 1 frame if not a gif)
		for frameNum := 0; frameNum < totalFrames; frameNum++ {
			// loop over a gif and apply it to all canvases (or apply a static image to every frame)
			imageCtx := imageContexts[frameNum%len(imageContexts)]

			// apply any filters set for the component
			for _, filterObject := range component.Filters {
				// check the filter exists and apply it
				var filterObj filter.Filter
				var ok bool
				if filterObj, ok = filters[filterObject.Name]; !ok {
					log.Println("Unknown filter type ", filterObject)
					continue
				}
				log.Println("Applying filter ", filterObject.Name, filterObject.Arguments)
				filterObj.ApplyFilter(imageCtx, filterObject.Arguments)
			}

			// check if there is an existing canvas for this frame
			var frameCanvas *gg.Context
			if len(outputCanvases) > frameNum {
				frameCanvas = outputCanvases[frameNum]
			} else {
				frameCanvas = gg.NewContext(request.Width, request.Height)
				outputCanvases = append(outputCanvases, frameCanvas)
			}

			// create a new canvas image context for this frame
			frameCanvas.RotateAbout(component.Rotation, float64(component.Position.X), float64(component.Position.Y))

			// if there is no width or height, set it from the current frame
			if component.Position.Width == 0 {
				component.Position.Width = imageCtx.Width()
			}
			if component.Position.Height == 0 {
				component.Position.Height = imageCtx.Height()
			}

			// check if the frame needs to be resized
			var dstImage *image.RGBA
			if component.Position.Width != imageCtx.Width() || component.Position.Height != imageCtx.Height() {
				// make a rectangle with the target bounds
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

				// scale the image
				draw.BiLinear.Scale(dstImage, dstImage.Bounds(), imageCtx.Image(), imageCtx.Image().Bounds(), draw.Over, &draw.Options{})
			} else {
				dstImage = imageCtx.Image().(*image.RGBA)
			}

			log.Println("Drawing image at ", component.Position.X, component.Position.Y)
			frameCanvas.DrawImage(dstImage, component.Position.X, component.Position.Y)

			// Reset the rotation
			frameCanvas.RotateAbout(-component.Rotation, float64(component.Position.X), float64(component.Position.Y))
		}
	}

	outputImages := make([]image.Image, len(outputCanvases))
	for i, canvas := range outputCanvases {
		outputImages[i] = canvas.Image()
	}
	result, extension := OutputImage(outputImages, request.Metadata)
	return &entity.ImageResult{
		Data:      result,
		Extension: extension,
	}
}

func max(i, i2 int) int {
	if i < i2 {
		return i2
	}
	return i
}

func getImageURL(url string) ([]*image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return getImage(response.Body)
}

func getLocalImage(url string) ([]*image.Image, error) {
	file, err := os.Open(path.Join("res", url))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return getImage(file)
}

func getImage(input io.Reader) ([]*image.Image, error) {
	body, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(body)
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	_, err = reader.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	if format == "gif" {
		fmt.Println("Decode the gif")
		gifFile, err := gif.DecodeAll(reader)
		if err != nil {
			return nil, err
		}
		output := make([]*image.Image, len(gifFile.Image))

		// use tmp to hold a stacked version of the frame
		firstFrame := gifFile.Image[0]
		tmp := image.NewNRGBA(firstFrame.Bounds())
		for i, img := range gifFile.Image {
			// stack over tmp
			draw.Draw(tmp, tmp.Bounds(), img, image.Point{X: 0, Y: 0}, draw.Over)

			// copy tmp as a new frame
			clone := image.NewPaletted(tmp.Bounds(), img.Palette)
			draw.Draw(clone, clone.Bounds(), tmp, tmp.Bounds().Min, draw.Src)

			genericImage := image.Image(clone)
			output[i] = &genericImage
		}
		return output, nil
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return []*image.Image{&img}, nil
}
