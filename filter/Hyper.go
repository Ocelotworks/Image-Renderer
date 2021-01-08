package filter

import (
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
	"image"
)

type Hyper struct{}

func (r Hyper) AfterStacking(filter *entity.Filter, request *entity.ImageRequest, component *entity.ImageComponent, images *[]*image.Image, delays *[]int) {

	return
}
