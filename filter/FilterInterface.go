package filter

import "github.com/fogleman/gg"

type Filter interface {
	ApplyFilter(ctx *gg.Context, args map[string]interface{}) *gg.Context
}
