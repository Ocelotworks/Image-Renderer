package filter

import (
	"fmt"
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
)

type Animate struct{}

func (a Animate) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {
	fmt.Println("Filter: ", filter)
	imageDeficit := len(filter.Arguments["frames"].([]interface{})) - len(*images)
	i := 0
	for imageDeficit > 0 {
		*images = append(*images, (*images)[i%len(*images)])
		if len(*delays) == 0 {
			*delays = []int{5}
		}
		*delays = append(*delays, (*delays)[i%len(*delays)])
		i++
		imageDeficit--
	}
}

func (a Animate) BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context {
	animFrames := args["frames"].([]interface{})

	animFrame := animFrames[frameNum%len(animFrames)].(map[string]interface{})

	component.Position.X = animFrame["x"]
	component.Position.Y = animFrame["y"]
	component.Position.Width = animFrame["w"]
	component.Position.Height = animFrame["h"]
	if animFrame["background"] != nil {
		component.Background = animFrame["background"].(string)
	}

	if animFrame["rotation"] != nil {
		component.Rotation = animFrame["rotation"].(float64)
	}

	return ctx
}
