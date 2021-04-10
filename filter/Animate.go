package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/helper"
	"image"
)

type Animate struct{}

func (a Animate) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {
	imageDeficit := float64(len(filter.Arguments["frames"].([]interface{})) - len(*images))
	i := 0

	delay := int(helper.GetFloatDefault(filter.Arguments["delay"], 5))

	for imageDeficit > 0 {
		*images = append(*images, (*images)[i%len(*images)])
		if len(*delays) == 0 {
			*delays = []int{delay}
		}
		*delays = append(*delays, (*delays)[i%len(*delays)])
		i++
		imageDeficit--
	}
}

func (a Animate) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {
	animFrames := args["frames"].([]interface{})

	animFrame := animFrames[frameNum%len(animFrames)].(map[string]interface{})

	if animFrame["x"] != nil {
		component.Position.X = animFrame["x"]
	}
	if animFrame["y"] != nil {
		component.Position.Y = animFrame["y"]
	}
	if animFrame["w"] != nil {
		component.Position.Width = animFrame["w"]
	}
	if animFrame["h"] != nil {
		component.Position.Height = animFrame["h"]
	}
	if animFrame["background"] != nil {
		component.Background = animFrame["background"].(string)
	}

	if animFrame["rotation"] != nil {
		component.Rotation = animFrame["rotation"].(float64)
	}

	return ctx
}
