package filter

import (
	"github.com/fogleman/gg"
	"gl.ocelotworks.com/ocelotbotv5/image-renderer/entity"
)

type Filter interface {
	Preprocess(request *entity.ImageRequest, component *entity.ImageComponent, filter *entity.Filter)
	ApplyFilter(ctx *gg.Context, args map[string]interface{}) *gg.Context
}
