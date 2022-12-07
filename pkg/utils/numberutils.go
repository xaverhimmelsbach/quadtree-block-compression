package utils

import "math"

// InRange returns wether two numbers are up to a maximum range away from each other
func InRange(number1 float64, number2 float64, maxRange float64) bool {
	difference := math.Abs(number1 - number2)
	return difference <= maxRange
}
