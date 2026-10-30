package acceptance_test

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"
)

const screenshotVersion = "v0.1.0"

var screenshotThemes = map[string]tuitest.Colors{
	"dark": {
		Foreground: "#F4F4F4", Background: "#141414", Cursor: "#F4F4F4",
		Black: "#3C3C3C", Red: "#F32735", Green: "#93D500", Yellow: "#FFD100", Blue: "#3D8FE0", Magenta: "#D04CB4", Cyan: "#00B6BC", White: "#DEDEDE",
		BrightBlack: "#808080", BrightRed: "#F32735", BrightGreen: "#93D500", BrightYellow: "#FFD100", BrightBlue: "#3D8FE0", BrightMagenta: "#D04CB4", BrightCyan: "#00B6BC", BrightWhite: "#F4F4F4",
	},
	"light": {
		Foreground: "#141414", Background: "#FFFFFF", Cursor: "#141414",
		Black: "#141414", Red: "#D41F2B", Green: "#4F7A00", Yellow: "#8A6D00", Blue: "#0071CE", Magenta: "#A81B8D", Cyan: "#00767B", White: "#3C3C3C",
		BrightBlack: "#808080", BrightRed: "#D41F2B", BrightGreen: "#4F7A00", BrightYellow: "#8A6D00", BrightBlue: "#0071CE", BrightMagenta: "#A81B8D", BrightCyan: "#00767B", BrightWhite: "#141414",
	},
}

func buildForScreenshots(t *testing.T) string {
	t.Helper()
	target := filepath.Join(t.TempDir(), "repo-tutor")
	flags := "-X repo-tutor/internal/about.version=" + screenshotVersion
	if output, err := exec.Command("go", "build", "-buildvcs=false", "-ldflags", flags, "-o", target, "../binary").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	return target
}

func withoutWindowFrame(t *testing.T, source, target string, background color.Color) {
	t.Helper()
	file, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	captured, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	bounds := captured.Bounds()
	middle := bounds.Min.X + bounds.Dx()/2
	left, right, top, bottom := bounds.Max.X, bounds.Min.X, bounds.Max.Y, bounds.Min.Y
	contentTop := -1
	wantRed, wantGreen, wantBlue, _ := background.RGBA()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if _, _, _, alpha := captured.At(x, y).RGBA(); alpha < 0xff00 {
				continue
			}
			left, right, top, bottom = min(left, x), max(right, x+1), min(top, y), max(bottom, y+1)
		}
		if red, green, blue, alpha := captured.At(middle, y).RGBA(); contentTop < 0 && alpha == 0xffff && red == wantRed && green == wantGreen && blue == wantBlue {
			contentTop = y
		}
	}
	if contentTop < 0 || right <= left || bottom <= contentTop {
		t.Fatalf("could not find the terminal inside %s", source)
	}
	cropped := image.NewNRGBA(image.Rect(0, 0, right-left, bottom-contentTop))
	draw.Draw(cropped, cropped.Bounds(), captured, image.Pt(left, contentTop), draw.Src)
	output, err := os.Create(target)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	if err := png.Encode(output, cropped); err != nil {
		t.Fatal(err)
	}
}

func TestScreenshots(t *testing.T) {
	directory := os.Getenv("REPO_TUTOR_SCREENSHOTS")
	if directory == "" {
		t.Skip("set REPO_TUTOR_SCREENSHOTS to a directory to write the README screenshots")
	}
	program := buildForScreenshots(t)
	for theme, colors := range screenshotThemes {
		terminal := tuitesttest.New(t, tuitesttest.Options{
			Program: program,
			Args:    []string{"testdata/bookshelf", "--agent", "codex", "--state-dir", t.TempDir(), "--goal", "Understand how bookshelf adds and saves a book", "--no-prefetch"},
			Spawn: tuitest.SpawnOptions{Profile: &tuitest.Profile{Colors: colors},
				Cols: tuitest.Ptr(uint16(120)), Rows: tuitest.Ptr(uint16(36)), WaitReady: tuitest.Ptr(false),
				Env: map[string]string{"PATH": agentDirectory + ":/usr/bin:/bin", "HOME": t.TempDir(), "COLORTERM": "truecolor", "FAKE_AGENT_SCENARIO": "screenshots"},
			},
		})
		waitForText(t, terminal, "Lesson ready")
		capture := func(name string) {
			t.Helper()
			raw := filepath.Join(t.TempDir(), name+".png")
			if _, err := terminal.Screenshot(raw, tuitest.ScreenshotOptions{Transparent: true}); err != nil {
				t.Fatal(err)
			}
			background, _ := parseHexColor(colors.Background)
			withoutWindowFrame(t, raw, filepath.Join(directory, name+"-"+theme+".png"), background)
		}
		capture("home")
		for _, page := range []struct{ key, heading, name string }{{"F2", "TAB STYLE", "settings"}, {"F3", "About repo-tutor", "about"}} {
			if err := terminal.Press(page.key); err != nil {
				t.Fatal(err)
			}
			waitForText(t, terminal, page.heading)
			capture(page.name)
			if err := terminal.Press("Escape"); err != nil {
				t.Fatal(err)
			}
			waitUntilGone(t, terminal, page.heading)
		}
		if err := terminal.Press("Enter"); err != nil {
			t.Fatal(err)
		}
		waitForText(t, terminal, "Every claim cites a file")
		capture("course")
	}
}

func parseHexColor(hex string) (color.Color, error) {
	var red, green, blue uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &red, &green, &blue)
	return color.RGBA{red, green, blue, 255}, err
}
