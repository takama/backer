package helper

import (
	"math"
)

// Round rounds float64 value with unit and precision
func Round(value float64, unit float64, precision int) float64 {

	pow := math.Pow(10, float64(precision))
	digit := pow * value
	_, frac := math.Modf(digit)

	var round float64
	if value > 0 {
		if frac >= unit {
			round = math.Ceil(digit)
		} else {
			round = math.Floor(digit)
		}
	} else {
		if math.Abs(frac) >= unit {
			round = math.Floor(digit)
		} else {
			round = math.Ceil(digit)
		}
	}

	return round / pow
}

// RoundPrice rounds price values presented in float32
func RoundPrice(price float32) float32 {
	return float32(Round(float64(price), float64(0.5), 2))
}

// TruncatePrice truncate a float32 to two levels of precision
func TruncatePrice(value float32) float32 {
	return float32(int(value*100)) / 100
}
