package helper

func GetStringDefault(value interface{}, defaultValue string) string {
	if value == nil {
		return defaultValue
	}
	cast, ok := value.(string)
	if !ok {
		return defaultValue
	}

	return cast
}

func GetFloatDefault(value interface{}, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}
	cast, ok := value.(float64)
	if !ok {
		return defaultValue
	}

	return cast
}

func GetBoolDefault(value interface{}, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	cast, ok := value.(bool)
	if !ok {
		return defaultValue
	}

	return cast
}
