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

const _defaultDelay = 10

var filters = map[string]filter.Filter{
	"rectangle": filter.Rectangle{},
}

// ProcessImage processes an incoming ImageRequest and outputs a finished ImageResult
func ProcessImage(request *entity.ImageRequest) *entity.ImageResult {
	// holds all the contexts for each frame of the final output image
	var outputContexts []*gg.Context

	// the delay for each frame of the output image
	var outputDelay []int

	// loop over each image component
	for _, component := range request.ImageComponents {
		var frameContexts []*gg.Context
		componentDelay := []int{}

		// fetch the image by URL/path if provided, otherwise make a blank image context
		if component.URL == "" {
			frameContexts = []*gg.Context{gg.NewContext(request.Width, request.Height)}
		} else {
			// decide which function to get the image with (explicitly typed)
			var getImageFunc func(url string) ([]*image.Image, []int, error) = getImageURL
			if component.Local {
				getImageFunc = getLocalImage
			}

			// get the image, returns all the frames if the image is a gif
			frameImages, frameDelay, err := getImageFunc(component.URL)
			if err != nil {
				log.Println(err)
				continue
			}
			componentDelay = frameDelay

			// create an image context for the image (or each frame for a gif)
			frameContexts = make([]*gg.Context, len(frameImages))
			for i, img := range frameImages {
				frameContexts[i] = gg.NewContextForImage(*img)
			}
		}

		totalFrames := max(len(frameContexts), len(outputContexts))

		// get the image context for each frame (only 1 frame if not a gif)
		for frameNum := 0; frameNum < totalFrames; frameNum++ {
			// loop over a gif and apply it to all canvases (or apply a static image to every frame)
			inputFrameCtx := frameContexts[frameNum%len(frameContexts)]

			// apply any filters set for the component
			for _, filterObject := range component.Filters {
				// check the filter exists and apply it
				var filterObj filter.Filter
				var ok bool
				if filterObj, ok = filters[filterObject.Name]; !ok {
					log.Println("Unknown filter type", filterObject)
					continue
				}
				log.Println("Applying filter", filterObject.Name, filterObject.Arguments)
				filterObj.ApplyFilter(inputFrameCtx, filterObject.Arguments)
			}

			// check if there is an existing context for this frame
			var outputCtx *gg.Context
			if frameNum < len(outputContexts) {
				outputCtx = outputContexts[frameNum]
			} else {
				outputCtx = gg.NewContext(request.Width, request.Height)
				outputContexts = append(outputContexts, outputCtx)
			}

			// set the delay for this frame if one doesn't exist yet
			if frameNum >= len(outputDelay) {
				delay := _defaultDelay
				// check if one exists from the input frames and set to that
				if frameNum < len(componentDelay) {
					delay = componentDelay[frameNum]
				}
				outputDelay = append(outputDelay, delay)
			}

			// create a new canvas image context for this frame
			outputCtx.RotateAbout(component.Rotation, float64(component.Position.X), float64(component.Position.Y))

			// if there is no width or height, set it from the current frame
			if component.Position.Width == 0 {
				component.Position.Width = inputFrameCtx.Width()
			}
			if component.Position.Height == 0 {
				component.Position.Height = inputFrameCtx.Height()
			}

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
						X: component.Position.Width,
						Y: component.Position.Height,
					},
				}

				log.Println("New size:", newSize)
				frameImage = image.NewRGBA(newSize)

				// scale the image
				draw.BiLinear.Scale(frameImage, frameImage.Bounds(), inputFrameCtx.Image(), inputFrameCtx.Image().Bounds(), draw.Over, &draw.Options{})
			} else {
				frameImage = inputFrameCtx.Image().(*image.RGBA)
			}

			log.Println("Drawing image at", component.Position.X, component.Position.Y)
			outputCtx.DrawImage(frameImage, component.Position.X, component.Position.Y)

			// Reset the rotation
			outputCtx.RotateAbout(-component.Rotation, float64(component.Position.X), float64(component.Position.Y))
		}
	}

	outputImages := make([]image.Image, len(outputContexts))
	for i, canvas := range outputContexts {
		outputImages[i] = canvas.Image()
	}
	result, extension := OutputImage(outputImages, outputDelay, request.Metadata)
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

func getImageURL(url string) ([]*image.Image, []int, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	return getImage(response.Body)
}

func getLocalImage(url string) ([]*image.Image, []int, error) {
	file, err := os.Open(path.Join("res", url))
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	return getImage(file)
}

func getImage(input io.Reader) ([]*image.Image, []int, error) {
	body, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, nil, err
	}

	reader := bytes.NewReader(body)
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, nil, err
	}

	_, err = reader.Seek(0, 0)
	if err != nil {
		return nil, nil, err
	}

	if format == "gif" {
		fmt.Println("Decode the gif")
		gifFile, err := gif.DecodeAll(reader)
		if err != nil {
			log.Fatalf("Error decoding gif: %s", err)
			return nil, nil, err
		}
		output := make([]*image.Image, len(gifFile.Image))

		// use tmp to hold a stacked version of the frame
		firstFrame := gifFile.Image[0]
		frameBg := image.NewNRGBA(firstFrame.Bounds())

		for i, img := range gifFile.Image {
			disposalMethod := gifFile.Disposal[i]

			if disposalMethod != 0 && disposalMethod != gif.DisposalNone {
				// clear the frame background if set to dispose the last frame
				frameBg = image.NewNRGBA(img.Bounds())
			}

			// draw onto the frame background, where frameBg is:
			//  - DisposalNone: sum of previous frames
			//  - DisposalBackground or DisposalPrevious: blank
			draw.Draw(frameBg, frameBg.Bounds(), img, image.Point{X: 0, Y: 0}, draw.Over)

			clone := image.NewPaletted(frameBg.Bounds(), img.Palette)
			draw.Draw(clone, clone.Bounds(), frameBg, image.Point{X: 0, Y: 0}, draw.Src)

			genericImage := image.Image(clone)
			output[i] = &genericImage
		}
		fmt.Println()
		return output, gifFile.Delay, nil
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, nil, err
	}
	return []*image.Image{&img}, []int{}, nil
}
