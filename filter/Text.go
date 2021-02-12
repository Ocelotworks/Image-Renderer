package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"path"
)

type Text struct{}

func (r Text) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {

	//// Debug
	//ctx.SetLineWidth(5)
	//ctx.SetHexColor("#00ff00")
	//ctx.DrawRectangle(args["x"].(float64), args["y"].(float64), args["w"].(float64), float64(ctx.Height()))
	//ctx.Stroke()
	//
	//ctx.DrawPoint(args["ax"].(float64), args["ay"].(float64), 25)
	//ctx.Fill()

	if args["w"] == 0 {
		args["w"] = ctx.Width()
	}

	ctx.SetHexColor(args["colour"].(string))
	_ = ctx.LoadFontFace(path.Join("res/font/", args["font"].(string)), args["fontSize"].(float64))

	ctx.DrawStringWrapped(args["content"].(string), args["x"].(float64), args["y"].(float64), args["ax"].(float64), args["ay"].(float64), args["w"].(float64), args["spacing"].(float64), gg.Align(args["align"].(float64)))
	return ctx
}
