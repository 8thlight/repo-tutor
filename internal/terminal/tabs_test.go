package terminal

import (
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func luminance(shade color.Color) float64 {
	red, green, blue, _ := shade.RGBA()
	linear := func(channel uint32) float64 {
		value := float64(channel) / 65535
		if value <= 0.03928 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(red) + 0.7152*linear(green) + 0.0722*linear(blue)
}

func contrastRatio(first, second color.Color) float64 {
	high, low := luminance(first), luminance(second)
	if low > high {
		high, low = low, high
	}
	return (high + 0.05) / (low + 0.05)
}

func TestTabTextKeepsReadableContrastOnBothFills(t *testing.T) {
	for _, dark := range []bool{true, false} {
		model := interfaceFixture()
		model.noColor, model.dark = false, dark
		palette := model.palette()
		for name, fill := range map[string]color.Color{"active": palette.accent, "inactive": palette.muted} {
			if ratio := contrastRatio(palette.cutout, fill); ratio < 4.5 {
				t.Fatalf("dark=%t %s tab text contrast is %.2f, want at least 4.5", dark, name, ratio)
			}
		}
	}
}

func TestTabTextUsesTheReportedTerminalBackground(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	model.background = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	if model.palette().cutout != model.background {
		t.Fatal("tab text should use the background the terminal reported")
	}
}

func TestTabHintsShortenAsTheLineNarrows(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	for width, want := range map[int]string{80: "Tab Next  F1 Help  F2 About  F3 Settings", 70: "F1 Help  F2 About  F3 Settings", 56: "F1 Help  F2 About", 44: "F1 Help", 40: ""} {
		bar := ansi.Strip(model.tabBar(width))
		if ansi.StringWidth(bar) > width {
			t.Fatalf("width %d: the bar is %d columns wide", width, ansi.StringWidth(bar))
		}
		if want != "" && !strings.HasSuffix(bar, want) {
			t.Fatalf("width %d should end with %q:\n%q", width, want, bar)
		}
		if want == "" && strings.Contains(bar, "F1") {
			t.Fatalf("width %d should drop the hints:\n%q", width, bar)
		}
	}
}

func TestTabBarMarksTheActiveTabWithoutColor(t *testing.T) {
	model := interfaceFixture()
	model.noColor = true
	bar := ansi.Strip(model.tabBar(120))
	if !strings.Contains(bar, "[Course]") || strings.Contains(bar, "[Home]") {
		t.Fatalf("only the active tab should be bracketed without color:\n%q", bar)
	}
}
