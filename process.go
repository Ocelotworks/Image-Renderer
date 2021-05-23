package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/stage"
	"image"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/filter"
)

const _defaultDelay = 10

// Performance metrics
var (
	processDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "process_duration",
		Help:      "Duration taken for the entire processing",
	})

	componentDrawDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "component_draw_duration",
		Help:      "Duration taken to stack component images",
	})

	beforeRenderFilterDuration = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "image_renderer",
		Name:      "filter_before_render_duration",
		Help:      "Duration taken to process BeforeRender filters",
	})
)

// ProcessImage processes an incoming ImageRequest and outputs a finished ImageResult
func ProcessImage(request *entity.ImageRequest) *entity.ImageResult {
	processDurationStart := time.Now()

	stage.ProcessBeforeStackingFilters(request)

	componentFrameDelays, componentFrameImages, exception := stage.MapComponentFrames(request)

	if exception != nil {
		return &entity.ImageResult{Error: "get_image"}
	}

	// holds all the contexts for each frame of the final output image
	outputContexts := make([]*gg.Context, 0)

	// the delay for each frame of the output image
	var outputDelay []int

	// Used to determine if the diff should be calculated
	shouldDiff := false

	for comp, component := range request.ImageComponents {
		componentDrawStart := time.Now()
		// Only components with a background should be diffed
		if component.URL != "" && component.Background != "" {
			shouldDiff = true
		}

		var frameContexts []*gg.Context

		frameImages := componentFrameImages[comp]
		frameDelay := componentFrameDelays[comp]
		componentDelay := frameDelay

		// Check for relative width/height and set to the correct value
		if ppw, ok := component.Position.Width.(string); ok {
			component.Position.Width = helper.GetRelativeDimension(request.Width, ppw)
			fmt.Println("Transforming width to ", component.Position.Width)
		}

		if pph, ok := component.Position.Height.(string); ok {
			component.Position.Height = helper.GetRelativeDimension(request.Height, pph)
			fmt.Println("Transforming height to ", component.Position.Height)
		}

		// If there are no frames in this image, create a new blank context of the correct width/height
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

			// Only apply the filter to the first frame of animated GIFs
			if frameNum == 0 || len(frameContexts) > 1 {
				// apply any filters set for the component
				for _, filterObject := range component.Filters {
					// check the filter exists and apply it
					var filterObj interface{}
					var ok bool
					if filterObj, ok = filter.Filters[filterObject.Name]; !ok {
						log.Println("Unknown filter type", filterObject)
						continue
					}
					if processFilter, ok := filterObj.(filter.BeforeRender); ok {
						log.Println("Applying filter", filterObject.Name, filterObject.Arguments)
						beforeRenderFilterStart := time.Now()
						processFilter.BeforeRender(inputFrameCtx, filterObject.Arguments, frameNum, component)
						beforeRenderFilterDuration.Observe(float64(time.Since(beforeRenderFilterStart).Milliseconds()))
					}

				}
			}

			// check if there is an existing context for this frame
			var outputCtx *gg.Context
			if frameNum < len(outputContexts) {
				outputCtx = outputContexts[frameNum]
			} else {
				// Check for a MaxWidth param, or default to 1920 and resize the image accordingly
				if request.MaxWidth > -1 && component.Position.Width != nil && component.Position.Height != nil {
					if request.MaxWidth == 0 {
						request.MaxWidth = 1920
					}
					componentWidth := int(component.Position.Width.(float64))
					componentHeight := int(component.Position.Height.(float64))
					if componentWidth > request.MaxWidth {
						component.Position.Height = float64(request.MaxWidth * componentHeight / componentWidth)
						component.Position.Width = float64(request.MaxWidth)
					}
				}

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

			stage.RotateAndResize(inputFrameCtx, outputCtx, component)

			// (Slow) optimisation for animated gifs
			if shouldDiff && comp == len(request.ImageComponents)-1 {
				stage.GIFOptimise(outputCtx, frameNum, &wg)
			}
			componentDrawDuration.Observe(float64(time.Since(componentDrawStart).Milliseconds()))
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

	output := OutputImage(outputImages, outputDelay, !shouldDiff, request)
	processDuration.Observe(float64(time.Since(processDurationStart).Milliseconds()))
	return output
}

func max(i, i2 int) int {
	if i < i2 {
		return i2
	}
	return i
}
