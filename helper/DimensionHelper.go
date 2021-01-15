package helper

import (
	"fmt"
	"strconv"
)

func GetRelativeDimension(parent int, relative string) float64 {
	fmt.Println(relative, relative[:len(relative)-1])
	raw, exception := strconv.ParseFloat(relative[:len(relative)-1], 64)

	if exception != nil {
		fmt.Println("Error parsing relative dimension ", exception)
		// i did not consider this
		return float64(parent)
	}

	return float64(parent) * (raw / 100)
}
