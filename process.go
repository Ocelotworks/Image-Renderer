package main

import (
	"bytes"
	"fmt"
	"github.com/getsentry/sentry-go"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
	"image/color"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
	"golang.org/x/image/draw"
)

const _defaultDelay = 10

var filters = map[string]interface{}{
	"rectangle": filter.Rectangle{},
	"text":      filter.Text{},
	"rainbow":   filter.Rainbow{},
	"hyper":     filter.Hyper{},
}

// ProcessImage processes an incoming ImageRequest and outputs a finished ImageResult
func ProcessImage(request *entity.ImageRequest) *entity.ImageResult {

	for _, component := range request.ImageComponents {
		for _, filterData := range component.Filters {
			var filterObj interface{}
			var ok bool
			if filterObj, ok = filters[filterData.Name]; !ok {
				log.Println("Unknown filter type", filterData)
				sentry.CaptureMessage(fmt.Sprintf("Unknown filter type '%s'", filterData))
				continue
			}
			if processFilter, ok := filterObj.(filter.BeforeStacking); ok {
				processFilter.BeforeStacking(request, component, filterData)
			}
		}
	}

	componentFrameImages := make([][]*image.Image, len(request.ImageComponents))
	componentFrameDelays := make([][]int, len(request.ImageComponents))

	//var wg sync.WaitGroup
	for comp, component := range request.ImageComponents {
		//wg.Add(1)
		//go func(comp int, component *entity.ImageComponent) {
		//defer wg.Done()
		if component.URL == "" {
			continue
		}

		if component.Position.X == nil {
			component.Position.X = float64(0)
		}

		if component.Position.Y == nil {
			component.Position.Y = float64(0)
		}

		if component.Position.Width == nil {
			component.Position.Width = float64(0)
		}

		if component.Position.Height == nil {
			component.Position.Height = float64(0)
		}

		fmt.Println(component)

		// decide which function to get the image with (explicitly typed)
		var getImageFunc = getImageURL
		if component.Local {
			getImageFunc = getLocalImage
		}

		// get the image, returns all the frames if the image is a gif
		frameImages, frameDelay, exception := getImageFunc(component.URL)
		if exception != nil {
			log.Println("Unable to get image:", exception)
			sentry.CaptureException(exception)
			return &entity.ImageResult{Error: "get_image"}
		}

		for _, filterData := range component.Filters {
			var filterObj interface{}
			var ok bool
			if filterObj, ok = filters[filterData.Name]; !ok {
				log.Println("Unknown filter type", filterData)
				continue
			}
			if processFilter, ok := filterObj.(filter.AfterStacking); ok {
				processFilter.AfterStacking(filterData, request, component, &frameImages, &frameDelay)
			}
		}

		go helper.WriteDebugPNG(*frameImages[0], fmt.Sprintf("comp-%d.frame-0.AfterStacking", comp))

		// Set the component width/height to the width/height of the first frame if it's not currently set
		if component.Position.Width == float64(0) {
			component.Position.Width = float64((*frameImages[0]).Bounds().Dx())
		}

		if component.Position.Height == float64(0) {
			component.Position.Height = float64((*frameImages[0]).Bounds().Dy())
		}

		componentFrameImages[comp] = frameImages
		componentFrameDelays[comp] = frameDelay
		//}(comp, component)
	}

	//wg.Done()

	// holds all the contexts for each frame of the final output image
	outputContexts := make([]*gg.Context, 0)

	// the delay for each frame of the output image
	var outputDelay []int

	var unmaskedPrevious image.Image

	shouldDiff := false
	// loop over each image component
	for comp, component := range request.ImageComponents {

		if component.Background != "" {
			shouldDiff = true
		}

		var frameContexts []*gg.Context
		componentDelay := []int{}

		frameImages := componentFrameImages[comp]
		frameDelay := componentFrameDelays[comp]
		componentDelay = frameDelay

		if ppw, ok := component.Position.Width.(string); ok {
			component.Position.Width = helper.GetRelativeDimension(request.Width, ppw)
			fmt.Println("Transforming width to ", component.Position.Width)
		}

		if pph, ok := component.Position.Height.(string); ok {
			component.Position.Height = helper.GetRelativeDimension(request.Height, pph)
			fmt.Println("Transforming height to ", component.Position.Height)
		}

		if len(frameImages) == 0 {
			ctx := gg.NewContext(int(component.Position.Width.(float64)), int(component.Position.Height.(float64)))
			if comp == 0 {
				if component.Background != "" {
					ctx.SetHexColor(component.Background)
					ctx.DrawRectangle(0, 0, float64(request.Width), float64(request.Height))
					ctx.Fill()
				}
			}
			frameContexts = []*gg.Context{ctx}
		} else {
			// create an image context for the image (or each frame for a gif)
			frameContexts = make([]*gg.Context, len(frameImages))
			for i, img := range frameImages {
				dx := (*img).Bounds().Dx()
				dy := (*img).Bounds().Dy()
				ctx := gg.NewContext(dx, dy)
				// this is a replacement for me figuring out the actual problems
				if component.Background != "" {
					ctx.SetHexColor(component.Background)
					ctx.DrawRectangle(0, 0, float64(dx), float64(dy))
					ctx.Fill()
				}
				ctx.DrawImage(*img, 0, 0)
				frameContexts[i] = ctx
			}
		}

		var wg sync.WaitGroup

		totalFrames := max(len(frameContexts), len(outputContexts))
		// get the image context for each frame (only 1 frame if not a gif)
		for frameNum := 0; frameNum < totalFrames; frameNum++ {
			// loop over a gif and apply it to all canvases (or apply a static image to every frame)
			inputFrameCtx := frameContexts[frameNum%len(frameContexts)]

			// apply any filters set for the component
			for _, filterObject := range component.Filters {
				// check the filter exists and apply it
				var filterObj interface{}
				var ok bool
				if filterObj, ok = filters[filterObject.Name]; !ok {
					log.Println("Unknown filter type", filterObject)
					continue
				}
				if processFilter, ok := filterObj.(filter.BeforeRender); ok {
					log.Println("Applying filter", filterObject.Name, filterObject.Arguments)
					processFilter.BeforeRender(inputFrameCtx, filterObject.Arguments, frameNum)
				}

			}

			// check if there is an existing context for this frame
			var outputCtx *gg.Context
			if frameNum < len(outputContexts) {
				outputCtx = outputContexts[frameNum]
			} else {
				if request.Width == 0 && component.Position.Width != nil {
					request.Width = int(component.Position.Width.(float64))
				}
				if request.Height == 0 && component.Position.Height != nil {
					request.Height = int(component.Position.Height.(float64))
				}
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

			if request.Debug {
				inputFrameCtx.SetLineWidth(10)
				inputFrameCtx.SetHexColor("#ff0000")
				inputFrameCtx.DrawRectangle(0, 0, float64(inputFrameCtx.Width()), float64(inputFrameCtx.Height()))
				inputFrameCtx.Stroke()
				inputFrameCtx.DrawStringWrapped(fmt.Sprintf("%dx%d", inputFrameCtx.Width(), inputFrameCtx.Height()), float64(inputFrameCtx.Width()), float64(inputFrameCtx.Height()), 1, 1, float64(inputFrameCtx.Width()), 1, gg.AlignLeft)
			}

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

			log.Printf("Drawing component %d frame %d at %f,%f\n", comp, frameNum, component.Position.X, component.Position.Y)
			outputCtx.DrawImage(frameImage, int(component.Position.X.(float64)), int(component.Position.Y.(float64)))

			// Reset the rotation
			outputCtx.RotateAbout(-component.Rotation, component.Position.X.(float64), component.Position.Y.(float64))

			// (Slow) optimisation for animated gifs
			if shouldDiff && comp == len(request.ImageComponents)-1 {
				// The image must be copied here, to avoid using the mask as the unmasked previous frame
				maskCopy := image.NewRGBA(outputCtx.Image().Bounds())
				draw.Copy(maskCopy, image.Point{X: 0, Y: 0}, outputCtx.Image(), outputCtx.Image().Bounds(), draw.Src, nil)
				if frameNum > 0 {
					wg.Add(1)
					go diffMask(outputCtx, unmaskedPrevious, &wg, frameNum)
				}
				unmaskedPrevious = maskCopy
			}
		}
		log.Println("Waiting for diff to finish...")
		wg.Wait()
		log.Println("Done!")
	}

	outputImages := make([]image.Image, len(outputContexts))
	for i, canvas := range outputContexts {
		outputImages[i] = canvas.Image()
	}

	if os.Getenv("DEBUG_DISABLE_RESPONSE") == "1" {
		return &entity.ImageResult{Error: "debug"}
	}

	result, extension, length, exception := OutputImage(outputImages, outputDelay, request.Metadata, !shouldDiff)
	if exception != nil {
		sentry.CaptureException(exception)
		return &entity.ImageResult{Error: "output"}
	}
	return &entity.ImageResult{
		Data:      result,
		Extension: extension,
		Size:      length,
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

// Erases pixels on `context` that are different to those on `image2`
func diffMask(context *gg.Context, image2 image.Image, wg *sync.WaitGroup, num int) {
	defer wg.Done()
	log.Printf("Diff for frame %d has finished", num)
	image1 := context.Image()
	dx := image1.Bounds().Dx()
	dy := image1.Bounds().Dy()
	context.SetRGBA255(0, 0, 0, 0)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			if coloursEqual(image1.At(x, y), image2.At(x, y)) {
				context.SetPixel(x, y)
			}
		}
	}
}

func coloursEqual(colour1 color.Color, colour2 color.Color) bool {
	r1, b1, g1, a1 := colour1.RGBA()
	r2, b2, g2, a2 := colour2.RGBA()
	return r1 == r2 && b1 == b2 && g1 == g2 && a1 == a2
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
		log.Println("Decoding the gif...")
		gifFile, err := gif.DecodeAll(reader)
		if err != nil {

			log.Printf("Error decoding gif: %s\n", err)
			return nil, nil, err
		}
		output := make([]*image.Image, len(gifFile.Image))

		log.Println("Stacking frames...")
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
		return output, gifFile.Delay, nil
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, nil, err
	}
	return []*image.Image{&img}, []int{}, nil
}
