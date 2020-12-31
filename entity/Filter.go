package entity

// Filter describes an image filter
type Filter struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"args"`
}
