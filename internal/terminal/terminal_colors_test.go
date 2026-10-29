package terminal

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

var fixedColorSequence = regexp.MustCompile(`\x1b\[[0-9;]*(?:38|48);(?:2;|5;(?:1[6-9]|[2-9][0-9]|[1-9][0-9]{2})\b)`)

func TestMarkdownStyleUsesOnlyTheSixteenTerminalColors(t *testing.T) {
	for _, highlight := range []string{"none", "red", "green", "cyan"} {
		model := interfaceFixture()
		model.noColor = false
		model.application.settings.Highlight = highlight
		var walk func(value reflect.Value, path string)
		walk = func(value reflect.Value, path string) {
			for index := 0; index < value.NumField(); index++ {
				field, name := value.Field(index), value.Type().Field(index).Name
				switch {
				case field.Kind() == reflect.Struct:
					walk(field, path+"."+name)
				case (name == "Color" || name == "BackgroundColor") && field.Kind() == reflect.Pointer && !field.IsNil():
					number, err := strconv.Atoi(field.Elem().String())
					if err != nil || number < 0 || number > 15 {
						t.Fatalf("%s: %s%s is %q, not one of the sixteen terminal colors", highlight, path, name, field.Elem().String())
					}
				}
			}
		}
		configuration := model.markdownStyle()
		walk(reflect.ValueOf(&configuration).Elem(), "")
		if configuration.CodeBlock.Chroma != nil {
			t.Fatalf("%s: code blocks should not carry their own colors", highlight)
		}
	}
}

func TestSourceHighlightingUsesOnlyTheSixteenTerminalColors(t *testing.T) {
	model := interfaceFixture()
	model.noColor = false
	code := "package main\n\n// run starts the tool.\nfunc run(name string) int {\n\treturn len(name) + 42\n}\n"
	rendered := strings.Join(model.highlightSource("main.go", code), "\n")
	if fixedColorSequence.MatchString(rendered) {
		t.Fatalf("source highlighting drew a color of its own: %q", rendered)
	}
	if !strings.Contains(rendered, "\x1b[") {
		t.Fatal("source highlighting should still color tokens")
	}
}

func TestSourceTokensKeepTheirMeaningWithAttributesAndTerminalColors(t *testing.T) {
	for kind, check := range map[chroma.TokenType]func(string) bool{
		chroma.Comment:              func(code string) bool { return strings.Contains(code, "3;2m") },
		chroma.Keyword:              func(code string) bool { return strings.Contains(code, "38;5;5") },
		chroma.LiteralString:        func(code string) bool { return strings.Contains(code, "38;5;2") },
		chroma.LiteralNumberInteger: func(code string) bool { return strings.Contains(code, "38;5;6") },
	} {
		if rendered := sourceTokenStyle(kind).Render("x"); !check(rendered) {
			t.Fatalf("%v rendered as %q", kind, rendered)
		}
	}
}
