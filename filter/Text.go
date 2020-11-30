package filter

import (
	"github.com/fogleman/gg"
)

type Text struct{}

func (r Text) ApplyFilter(ctx *gg.Context, args map[string]interface{}) *gg.Context {
	ctx.SetHexColor(args["colour"].(string))
	ctx.DrawStringWrapped(args["content"].(string), args["x"].(float64), args["y"].(float64), args["ax"].(float64), args["ay"].(float64), args["w"].(float64), args["spacing"].(float64), gg.Align(args["align"].(int)))
	ctx.DrawRectangle(args["X"].(float64), args["Y"].(float64), args["W"].(float64), args["H"].(float64))
	ctx.Fill()
	return ctx
}
