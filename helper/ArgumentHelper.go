package helper

import (
	"github.com/Knetic/govaluate"
	"log"
)

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

func EvaluateFloat(input string, defaultValue float64, parameters map[string]interface{}) float64 {
	expression, exception := govaluate.NewEvaluableExpression(input)
	if exception != nil {
		log.Println("Couldn't parse expression", exception)
		return defaultValue
	}
	value, exception := expression.Evaluate(parameters)
	if exception != nil {
		log.Println("Couldn't evaluate expression", exception)
		return defaultValue
	}

	castValue, ok := value.(float64)
	if !ok {
		log.Println("Returned value was not a float")
		return defaultValue
	}

	return castValue
}

func ParseFloat(value interface{}, defaultValue float64, parameters map[string]interface{}) float64 {
	if value == nil {
		return defaultValue
	}
	floatCast, ok := value.(float64)
	if ok {
		return floatCast
	}

	stringCast, ok := value.(string)
	if ok {
		return EvaluateFloat(stringCast, defaultValue, parameters)
	}

	return defaultValue
}
