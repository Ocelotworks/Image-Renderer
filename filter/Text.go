package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"path"
)

type Text struct{}

func (r Text) BeforeProcess(request *entity.ImageRequest, component *entity.ImageComponent, filter *entity.Filter) {
	// TODO: emoji parsing and shit
	//parser := emoji.NewEmojiParser()
	//i := 0
	//filter.Arguments["content"] = parser.ReplaceAllStringFunc(filter.Arguments["content"].(string), func(s string) string {
	//	i++
	//	r, _ := utf8.DecodeRuneInString(s)
	//	request.ImageComponents = append(request.ImageComponents, &entity.ImageComponent{
	//		URL:  fmt.Sprintf("https://twemoji.maxcdn.com/%[1]dx%[1]d/%[2]x.png", filter.Arguments["fontSize"], r),
	//		Local: false,
	//		Position: struct {
	//			X      int `json:"x"`
	//			Y      int `json:"y"`
	//			Width  int `json:"w"`
	//			Height int `json:"h"`
	//		}{
	//			X: int(float64(i)*(filter.Arguments["fontSize"].(float64))),
	//		},
	//		Rotation:   0,
	//		Filters:    nil,
	//		Background: "",
	//	})
	//	return " "
	//})
}

func (r Text) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int) *gg.Context {
	ctx.SetHexColor(args["colour"].(string))
	_ = ctx.LoadFontFace(path.Join("res/font/", args["font"].(string)), args["fontSize"].(float64))
	ctx.DrawStringWrapped(args["content"].(string), args["x"].(float64), args["y"].(float64), args["ax"].(float64), args["ay"].(float64), args["w"].(float64), args["spacing"].(float64), gg.Align(args["align"].(float64)))
	return ctx
}
