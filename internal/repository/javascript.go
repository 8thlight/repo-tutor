package repository

import (
	"encoding/json"
	"path"
	"regexp"
	"strings"
)

var javascriptImport = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*|\brequire\s*\(\s*|\bimport\s*\(\s*)["']([^"']+)["']`)
var javascriptDeclaration = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:declare\s+)?(?:async\s+)?(function|class|interface|type|enum)\s+([\w]+)`)
var javascriptBranch = regexp.MustCompile(`\b(if|for|while|case|catch)\b|&&|\|\||\?\?`)

var javascriptAnalyzer = analyzer{
	name:       "JavaScript/TypeScript",
	extensions: []string{".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx"},
	manifests:  []string{"package.json"},
	branches:   javascriptBranch,
	manifest:   javascriptManifest,
	scan:       javascriptScan,
	index:      javascriptIndex,
}

func javascriptManifest(file File) (Module, string, bool) {
	var manifest struct {
		Name string `json:"name"`
	}
	if json.Unmarshal([]byte(file.Content), &manifest) != nil {
		return Module{}, file.Path + ": invalid package.json", false
	}
	if manifest.Name == "" {
		manifest.Name = path.Dir(file.Path)
	}
	return Module{manifest.Name, path.Dir(file.Path), "JavaScript/TypeScript"}, "", true
}

func javascriptScan(file File) scanned {
	var found scanned
	for index, line := range strings.Split(file.Content, "\n") {
		for _, match := range javascriptImport.FindAllStringSubmatch(line, -1) {
			found.imports = append(found.imports, match[1])
		}
		if match := javascriptDeclaration.FindStringSubmatch(line); match != nil {
			kind := "type"
			if match[1] == "function" {
				kind = "function"
			}
			found.symbols = append(found.symbols, Symbol{match[2], kind, Location{file.Path, index + 1}})
		}
	}
	return found
}

func javascriptIndex(modules []Module, _ []File, owners map[string]string) resolver {
	names := map[string][]string{}
	for _, module := range modules {
		if module.Language == "JavaScript/TypeScript" {
			names[module.Path] = append(names[module.Path], module.Directory)
		}
	}
	return resolver{resolve: func(from, target string) string {
		if strings.HasPrefix(target, ".") {
			return javascriptTarget(path.Join(path.Dir(from), target), owners)
		}
		if directories := names[target]; len(directories) == 1 {
			for _, entry := range []string{"src/index", "index"} {
				if resolved := javascriptTarget(path.Join(directories[0], entry), owners); resolved != "" {
					return resolved
				}
			}
		}
		return ""
	}}
}

func javascriptTarget(name string, owners map[string]string) string {
	if owner := owners[name]; owner != "" {
		return owner
	}
	for _, extension := range []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", "/index.ts", "/index.tsx", "/index.js"} {
		if owner := owners[name+extension]; owner != "" {
			return owner
		}
	}
	if strings.HasSuffix(name, ".js") {
		return owners[strings.TrimSuffix(name, ".js")+".ts"]
	}
	return ""
}
