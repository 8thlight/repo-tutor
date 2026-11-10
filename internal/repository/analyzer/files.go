package analyzer

import (
	"path"
	"strings"
)

func Excluded(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "vendor", "node_modules", "output", "dist", "build", "result", "testdata", "fixtures", "__pycache__", "coverage":
		return true
	}
	return false
}

func Supported(name string) bool {
	if strings.HasPrefix(name, ".") {
		return false
	}
	if name == "go.mod" || manifestFile(name) || sourceExtension(name) {
		return true
	}
	switch strings.ToLower(path.Ext(name)) {
	case ".go", ".md":
		return true
	}
	return false
}
