package stage

import (
	"bytes"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var filters = map[string]interface{}{
	"rectangle": filter.Rectangle{},
	"text":      filter.Text{},
	"rainbow":   filter.Rainbow{},
	"hyper":     filter.Hyper{},
	"animate":   filter.Animate{},
}

var (
	componentStackDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "component_stack_duration",
		Help:      "Duration taken to stack component images",
	})
	beforeStackingFilterDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "filter_before_stacking_duration",
		Help:      "Duration taken to process BeforeStacking filters",
	})
)

// Does the BeforeStacking filters
func ProcessBeforeStackingFilters(request *entity.ImageRequest) {
	for _, component := range request.ImageComponents {
		for _, filterData := range component.Filters {
			var filterObj interface{}
			var ok bool
			if filterObj, ok = filter.Filters[filterData.Name]; !ok {
				log.Println("Unknown filter type", filterData)
				sentry.CaptureMessage(fmt.Sprintf("Unknown filter type '%s'", filterData))
				continue
			}
			if processFilter, ok := filterObj.(filter.BeforeStacking); ok {
				beforeStackingStart := time.Now()
				processFilter.BeforeStacking(request, component, filterData)
				beforeStackingFilterDuration.Observe(float64(time.Since(beforeStackingStart).Milliseconds()))
			}
		}

	}
}

func isFloat(value interface{}) bool {
	_, ok := value.(float64)
	return ok
}

// Loads every image in the request and maps them into arrays of images and delays
func MapComponentFrames(request *entity.ImageRequest) ([][]int, [][]*image.Image, error) {
	componentFrameImages := make([][]*image.Image, len(request.ImageComponents))
	componentFrameDelays := make([][]int, len(request.ImageComponents))

	for comp, component := range request.ImageComponents {
		componentStackStart := time.Now()

		if component.Position.X == nil || !isFloat(component.Position.X) {
			component.Position.X = float64(0)
		}

		if component.Position.Y == nil || !isFloat(component.Position.X) {
			component.Position.Y = float64(0)
		}

		if component.Position.Width == nil {
			component.Position.Width = float64(0)
		}

		if component.Position.Height == nil {
			component.Position.Height = float64(0)
		}

		if component.URL == "" {
			continue
		}

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
			return nil, nil, exception
			//return &entity.ImageResult{Error: "get_image"}
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
		componentStackDuration.Observe(float64(time.Since(componentStackStart).Milliseconds()))

	}
	return componentFrameDelays, componentFrameImages, nil
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

// Takes an input image in a supported format and redraws it as an array of NRGBA frames
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

		// Convert the gif into a series of NRGBA frames
		for i, img := range gifFile.Image {
			disposalMethod := gifFile.Disposal[i]
			// Depending on the disposal method, reset frameBg to a blank slate
			//  - DisposalNone: sum of previous frames
			//  - DisposalBackground or DisposalPrevious: blank
			if disposalMethod != 0 && disposalMethod != gif.DisposalNone {
				frameBg = image.NewNRGBA(img.Bounds())
			}
			// Iterate over the palette image, converting it into sets of RGBA bytes
			for i, colour := range img.Pix {
				r, g, b, a := img.Palette[colour].RGBA()
				// Skip drawing transparent pixels to avoid overwriting the previous frame in those Disposal modes
				if a == 0 {
					continue
				}
				// Each pixel contains 4 bytes
				p := i * 4
				frameBg.Pix[p] = uint8(r)
				frameBg.Pix[p+1] = uint8(g)
				frameBg.Pix[p+2] = uint8(b)
				frameBg.Pix[p+3] = uint8(a)
			}
			// Set the output to a clone of the current state of frameBg
			// Copy the pix array for speed
			clone := image.NewNRGBA(frameBg.Bounds())
			copy(clone.Pix, frameBg.Pix)
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
