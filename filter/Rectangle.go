package filter

import (
	"github.com/fogleman/gg"
)

type Rectangle struct{}

func (r Rectangle) ApplyFilter(ctx *gg.Context, args map[string]interface{}) *gg.Context {
	ctx.SetHexColor(args["colour"].(string))
	ctx.DrawRectangle(args["x"].(float64), args["y"].(float64), args["w"].(float64), args["h"].(float64))
	ctx.Fill()
	return ctx
}