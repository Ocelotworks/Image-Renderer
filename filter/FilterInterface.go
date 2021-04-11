package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
)

var Filters = map[string]interface{}{
	"rectangle": Rectangle{},
	"text":      Text{},
	"rainbow":   Rainbow{},
	"greyscale": Greyscale{},
	"hyper":     Hyper{},
	"animate":   Animate{},
}

type BeforeRender interface {
	BeforeRender(ctx *gg.Context, args map[string]interface{}, frameNum int, component *entity.ImageComponent) *gg.Context
}

type BeforeStacking interface {
	BeforeStacking(request *entity.ImageRequest, component *entity.ImageComponent, filter *entity.Filter)
}

type AfterStacking interface {
	AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int)
}
