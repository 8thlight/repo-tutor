package terminal

import (
	"reflect"
	"strconv"

	glamouransi "charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

const (
	ansiBlue      = 4
	glamourColor  = "Color"
	glamourBackup = "BackgroundColor"
)

func withoutColors(value reflect.Value) {
	for index := 0; index < value.NumField(); index++ {
		field, name := value.Field(index), value.Type().Field(index).Name
		switch {
		case field.Kind() == reflect.Struct:
			withoutColors(field)
		case (name == glamourColor || name == glamourBackup) && field.Kind() == reflect.Pointer:
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

func terminalMarkdownStyle(accent int, noAccent bool) glamouransi.StyleConfig {
	configuration := styles.DarkStyleConfig
	configuration.CodeBlock.Chroma = nil
	withoutColors(reflect.ValueOf(&configuration).Elem())
	paint := func(index int) *string {
		value := strconv.Itoa(index)
		return &value
	}
	if !noAccent {
		for _, heading := range []*glamouransi.StyleBlock{&configuration.Heading, &configuration.H1, &configuration.H2, &configuration.H3, &configuration.H4, &configuration.H5, &configuration.H6} {
			heading.Color = paint(accent)
		}
		configuration.Code.Color = paint(accent)
	}
	configuration.Link.Color = paint(ansiBlue)
	configuration.LinkText.Color = paint(ansiBlue)
	faint := true
	configuration.HorizontalRule.Faint = &faint
	configuration.BlockQuote.Faint = &faint
	return configuration
}
