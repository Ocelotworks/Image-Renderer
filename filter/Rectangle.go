package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
)

type Rectangle struct{}

func (r Rectangle) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {
	colour, ok := args["colour"].(string)
	if ok {
		ctx.SetHexColor(colour)
	}

	ctx.DrawRectangle(conditionalFloat(args["x"], 0), conditionalFloat(args["y"], 0), conditionalFloat(args["w"], float64(ctx.Width())), conditionalFloat(args["h"], float64(ctx.Height())))

	fill, ok := args["fill"].(bool)
	if !ok || fill {
		ctx.Fill()
	} else {
		ctx.Stroke()
	}
	return ctx
}

func conditionalFloat(input interface{}, defaultValue float64) float64 {
	castInput, ok := input.(float64)
	if !ok {
		return defaultValue
	}
	return castInput
}
