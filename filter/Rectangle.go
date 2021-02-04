package filter

import (
	"github.com/fogleman/gg"
)

type Rectangle struct{}

func (r Rectangle) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int) *gg.Context {
	ctx.SetHexColor(args["colour"].(string))
	ctx.DrawRectangle(args["x"].(float64), args["y"].(float64), args["w"].(float64), args["h"].(float64))
	if args["fill"].(bool) {
		ctx.Fill()
	} else {
		ctx.Stroke()
	}
	return ctx
}
