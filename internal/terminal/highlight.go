package terminal

import (
	"fmt"
	"image/color"
	"math"

	"charm.land/lipgloss/v2"

	"repo-tutor/internal/settings"
)

const readableContrast = 4.6

type highlightColor struct{ id, label, hex string }

var highlightColors = []highlightColor{
	{"red", "Red", "#F32735"},
	{"orange", "Orange", "#FF8300"},
	{"yellow", "Yellow", "#FFD100"},
	{"green", "Green", "#93D500"},
	{"blue", "Blue", "#0071CE"},
	{"dark_blue", "Dark blue", "#004987"},
	{"purple", "Purple", "#A81B8D"},
}

func highlightByID(id string) highlightColor {
	for _, candidate := range highlightColors {
		if candidate.id == id {
			return candidate
		}
	}
	return highlightByID(settings.HighlightGreen)
}

func (model interfaceModel) backgroundColor() color.Color {
	if model.dark {
		return model.cutoutColor("#09090B")
	}
	return model.cutoutColor("#FAFAFA")
}

func (model interfaceModel) highlight(id string) color.Color {
	return readableOn(lipgloss.Color(highlightByID(id).hex), model.backgroundColor())
}

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

func toHSL(shade color.Color) (hue, saturation, lightness float64) {
	red, green, blue := channels(shade)
	high, low := math.Max(red, math.Max(green, blue)), math.Min(red, math.Min(green, blue))
	lightness = (high + low) / 2
	if high == low {
		return 0, 0, lightness
	}
	span := high - low
	saturation = span / (1 - math.Abs(2*lightness-1))
	switch high {
	case red:
		hue = math.Mod((green-blue)/span, 6)
	case green:
		hue = (blue-red)/span + 2
	default:
		hue = (red-green)/span + 4
	}
	return math.Mod(hue*60+360, 360), saturation, lightness
}

func fromHSL(hue, saturation, lightness float64) color.Color {
	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	second := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	offset := lightness - chroma/2
	var red, green, blue float64
	switch {
	case hue < 60:
		red, green, blue = chroma, second, 0
	case hue < 120:
		red, green, blue = second, chroma, 0
	case hue < 180:
		red, green, blue = 0, chroma, second
	case hue < 240:
		red, green, blue = 0, second, chroma
	case hue < 300:
		red, green, blue = second, 0, chroma
	default:
		red, green, blue = chroma, 0, second
	}
	to8 := func(value float64) uint8 { return uint8(math.Round((value + offset) * 255)) }
	return color.RGBA{to8(red), to8(green), to8(blue), 255}
}

func readableOn(brand, background color.Color) color.Color {
	if contrastBetween(brand, background) >= readableContrast {
		return brand
	}
	hue, saturation, lightness := toHSL(brand)
	lighten := relativeLuminance(background) <= 0.179
	near, far := lightness, 0.0
	if lighten {
		far = 1.0
	}
	if contrastBetween(fromHSL(hue, saturation, far), background) < readableContrast {
		return fromHSL(hue, saturation, far)
	}
	for range 40 {
		middle := (near + far) / 2
		if contrastBetween(fromHSL(hue, saturation, middle), background) >= readableContrast {
			far = middle
		} else {
			near = middle
		}
	}
	return fromHSL(hue, saturation, far)
}
