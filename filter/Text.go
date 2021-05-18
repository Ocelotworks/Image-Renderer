package filter

import (
	"fmt"
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image/color"
	"path"
	"strings"
)

type Text struct{}

func (r Text) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {

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
	ax := helper.ParseFloat(args["ax"], 0, evalParams)
	ay := helper.ParseFloat(args["ay"], 0, evalParams)
	w := helper.ParseFloat(args["w"], float64(ctx.Width()), evalParams)
	colour := helper.GetStringDefault(args["colour"], "#ffffff")
	content := helper.GetStringDefault(args["content"], "<no content>")
	spacing := helper.GetFloatDefault(args["spacing"], 1.1)
	align := helper.GetFloatDefault(args["align"], 0)
	fontSize := helper.ParseFloat(args["fontSize"], 24, evalParams)
	font := helper.GetStringDefault(args["font"], "arial.ttf")

	_ = ctx.LoadFontFace(path.Join("res/font/", font), fontSize)
	if args["shadowColour"] != nil {
		shadowColour := helper.GetStringDefault(args["shadowColour"], "#000000")
		ctx.SetHexColor(shadowColour)
		ctx.DrawStringWrapped(content, x-3, y-3, ax, ay, w, spacing, gg.Align(align))
	}

	if args["outlineColour"] != nil {
		outlineColour := helper.GetStringDefault(args["outlineColour"], "#000000")
		ctx.SetHexColor(outlineColour)
		// Loop shmoop
		ctx.DrawStringWrapped(content, x+2, y-2, ax, ay, w, spacing, gg.Align(align))
		ctx.DrawStringWrapped(content, x-2, y-2, ax, ay, w, spacing, gg.Align(align))
		ctx.DrawStringWrapped(content, x+2, y+2, ax, ay, w, spacing, gg.Align(align))
		ctx.DrawStringWrapped(content, x-2, y+2, ax, ay, w, spacing, gg.Align(align))
	}

	if args["gradient"] != nil {
		wrappedText := ctx.WordWrap(content, w)
		textW, textH := ctx.MeasureMultilineString(strings.Join(wrappedText, "\n"), spacing)
		textContext := gg.NewContext(int(textW+5), int(textH+30))
		textContext.SetRGB(0, 0, 0)
		_ = textContext.LoadFontFace(path.Join("res/font/", font), fontSize)
		textContext.DrawStringWrapped(content, 0, 5, ax, ay, w, spacing, gg.Align(align))
		textMask := textContext.AsMask()

		// Create the gradient and set the original context to use it
		gradient := args["gradient"].([]interface{})
		grad := gg.NewLinearGradient(0, 0, 0, fontSize)
		for i, stop := range gradient {
			r, g, b, a := parseHexColor(stop.(string))
			grad.AddColorStop(float64(i), color.RGBA{R: r, G: g, B: b, A: a})
		}
		textContext.SetFillStyle(grad)

		//Set the mask to our text path
		_ = textContext.SetMask(textMask)
		// Fill the region the text is in with the gradient
		textContext.DrawRectangle(0, 0, float64(textContext.Width()), float64(textContext.Height()))
		textContext.Fill()
		// Reset the mask
		//_ = ctx.SetMask(nil)
		ctx.DrawImage(textContext.Image(), int(x), int(y-5))
	} else {
		ctx.SetHexColor(colour)
		ctx.DrawStringWrapped(content, x, y, ax, ay, w, spacing, gg.Align(align))
	}

	return ctx
}

func parseHexColor(x string) (r, g, b, a uint8) {
	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		_, _ = fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		_, _ = fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		_, _ = fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return
}
