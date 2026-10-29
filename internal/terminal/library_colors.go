package terminal

import (
	"image/color"
	"reflect"

	"charm.land/lipgloss/v2"
)

var (
	styleType = reflect.TypeFor[lipgloss.Style]()
	colorType = reflect.TypeFor[color.Color]()
)

func withoutStyleColors(target any) {
	stripColors(reflect.ValueOf(target).Elem())
}

func stripColors(value reflect.Value) {
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		switch {
		case !field.CanSet():
		case field.Type() == styleType:
			style := field.Interface().(lipgloss.Style)
			field.Set(reflect.ValueOf(style.UnsetForeground().UnsetBackground().UnsetBorderForeground().UnsetBorderBackground()))
		case field.Type() == colorType:
			field.Set(reflect.Zero(field.Type()))
		case field.Kind() == reflect.Struct:
			stripColors(field)
		}
	}
}
