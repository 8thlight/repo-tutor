package terminal

import (
	"fmt"
	"image/color"
	"math"
)

func hexOf(shade color.Color) string {
	red, green, blue, _ := shade.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", red>>8, green>>8, blue>>8)
}

func channels(shade color.Color) (float64, float64, float64) {
	red, green, blue, _ := shade.RGBA()
	return float64(red) / 65535, float64(green) / 65535, float64(blue) / 65535
}

func relativeLuminance(shade color.Color) float64 {
	linear := func(value float64) float64 {
		if value <= 0.03928 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	red, green, blue := channels(shade)
	return 0.2126*linear(red) + 0.7152*linear(green) + 0.0722*linear(blue)
}

func contrastBetween(first, second color.Color) float64 {
	high, low := relativeLuminance(first), relativeLuminance(second)
	if low > high {
		high, low = low, high
	}
	return (high + 0.05) / (low + 0.05)
}

func blend(from, to color.Color, share float64) color.Color {
	fromRed, fromGreen, fromBlue := channels(from)
	toRed, toGreen, toBlue := channels(to)
	mix := func(start, end float64) uint8 { return uint8(math.Round((start + (end-start)*share) * 255)) }
	return color.RGBA{mix(fromRed, toRed), mix(fromGreen, toGreen), mix(fromBlue, toBlue), 255}
}
