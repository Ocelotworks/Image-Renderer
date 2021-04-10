package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
)

type Rectangle struct{}

func (r Rectangle) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {
	x := helper.GetFloatDefault(args["x"], 0)
	y := helper.GetFloatDefault(args["y"], 0)
	w := helper.GetFloatDefault(args["w"], float64(ctx.Width()))
	h := helper.GetFloatDefault(args["h"], float64(ctx.Height()))
	fill := helper.GetBoolDefault(args["fill"], true)
	colour := helper.GetStringDefault(args["colour"], "#000000")

	ctx.SetHexColor(colour)
	ctx.DrawRectangle(x, y, w, h)

	if fill {
		ctx.Fill()
	} else {
		ctx.Stroke()
	}

	return ctx
}
