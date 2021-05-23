package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
)

type Rectangle struct{}

func (r Rectangle) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {
	evalParams := map[string]interface{}{
		"frameNum":  frameNum,
		"component": component,
		"url":       component.URL,
		"cx":        component.Position.X,
		"cy":        component.Position.Y,
		"cw":        component.Position.Width,
		"ch":        component.Position.Height,
		"ctxw":      ctx.Width(),
		"ctxh":      ctx.Height(),
	}

	x := helper.ParseFloat(args["x"], 0, evalParams)
	y := helper.ParseFloat(args["y"], 0, evalParams)
	w := helper.ParseFloat(args["w"], float64(ctx.Width()), evalParams)
	h := helper.ParseFloat(args["h"], float64(ctx.Height()), evalParams)
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
