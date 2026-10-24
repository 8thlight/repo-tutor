package repository

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strconv"
	"strings"
)

type File struct {
	Path    string `json:"path"`
	Content string `json:"-"`
}

type Location struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

type Symbol struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Location
}

type Package struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	Directory    string   `json:"directory"`
	Files        []string `json:"files"`
	Tests        []string `json:"tests,omitempty"`
	Imports      []string `json:"imports,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Symbols      []Symbol `json:"symbols,omitempty"`
}

type Module struct {
	Path      string `json:"path"`
	Directory string `json:"directory"`
	Language  string `json:"language"`
}

type Snapshot struct {
	Root        string     `json:"root"`
	Fingerprint string     `json:"fingerprint"`
	Modules     []Module   `json:"modules"`
	Packages    []Package  `json:"packages"`
	Documents   []string   `json:"documents"`
	EntryPoints []Location `json:"entry_points"`
	Hotspots    []Hotspot  `json:"hotspots"`
	History     bool       `json:"history_available"`
	Warnings    []string   `json:"warnings,omitempty"`
	Files       []File     `json:"-"`
}

func Analyze(files []File) Snapshot {
	ordered := append([]File(nil), files...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	result := Snapshot{Files: ordered}
	for _, file := range ordered {
		if path.Base(file.Path) == "go.mod" {
			module := modulePath(file.Content)
			if module == "" {
				result.Warnings = append(result.Warnings, file.Path+": no module directive")
				continue
			}
			result.Modules = append(result.Modules, Module{module, path.Dir(file.Path), "Go"})
		}
		if strings.EqualFold(path.Ext(file.Path), ".md") {
			result.Documents = append(result.Documents, file.Path)
		}
	}
	packages := map[string]*Package{}
	for _, file := range ordered {
		if path.Ext(file.Path) != ".go" {
			continue
		}
		positions := token.NewFileSet()
		parsed, err := parser.ParseFile(positions, file.Path, file.Content, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", file.Path, err))
			continue
		}
		if ast.IsGenerated(parsed) {
			continue
		}
		directory := path.Dir(file.Path)
		module, found := nearestModule(result.Modules, directory)
		if !found {
			result.Warnings = append(result.Warnings, file.Path+": outside a Go module")
			continue
		}
		packagePath := module.Path
		if directory != module.Directory {
			relative := directory
			if module.Directory != "." {
				relative = strings.TrimPrefix(directory, module.Directory+"/")
			}
			packagePath += "/" + relative
		}
		current, exists := packages[packagePath]
		if !exists {
			current = &Package{Path: packagePath, Directory: directory, Name: strings.TrimSuffix(parsed.Name.Name, "_test"), Language: "Go"}
			packages[packagePath] = current
		}
		if strings.HasSuffix(file.Path, "_test.go") {
			current.Tests = append(current.Tests, file.Path)
			continue
		}
		current.Name = parsed.Name.Name
		current.Files = append(current.Files, file.Path)
		for _, imported := range parsed.Imports {
			value, err := strconv.Unquote(imported.Path.Value)
			if err == nil {
				current.Imports = append(current.Imports, value)
			}
		}
		for _, declaration := range parsed.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				name := declaration.Name.Name
				kind := "function"
				if declaration.Recv != nil {
					kind = "method"
				}
				location := Location{file.Path, positions.Position(declaration.Pos()).Line}
				current.Symbols = append(current.Symbols, Symbol{name, kind, location})
				if current.Name == "main" && name == "main" && declaration.Recv == nil {
					result.EntryPoints = append(result.EntryPoints, location)
				}
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					if definition, ok := specification.(*ast.TypeSpec); ok {
						current.Symbols = append(current.Symbols, Symbol{definition.Name.Name, "type", Location{file.Path, positions.Position(definition.Pos()).Line}})
					}
				}
			}
		}
	}
	paths := make([]string, 0, len(packages))
	for packagePath := range packages {
		paths = append(paths, packagePath)
	}
	sort.Strings(paths)
	for _, packagePath := range paths {
		current := packages[packagePath]
		current.Imports = unique(current.Imports)
		for _, imported := range current.Imports {
			if _, exists := packages[imported]; exists && imported != packagePath {
				current.Dependencies = append(current.Dependencies, imported)
			}
		}
		result.Packages = append(result.Packages, *current)
	}
	return analyzeLanguages(result)
}

func (snapshot Snapshot) File(name string) (File, bool) {
	for _, file := range snapshot.Files {
		if file.Path == name {
			return file, true
		}
	}
	return File{}, false
}

func (snapshot Snapshot) Package(name string) (Package, bool) {
	for _, current := range snapshot.Packages {
		if current.Path == name {
			return current, true
		}
	}
	return Package{}, false
}

func Foundations(snapshot Snapshot) []Package {
	visited := map[string]bool{}
	var result []Package
	var visit func(Package)
	visit = func(current Package) {
		if visited[current.Path] {
			return
		}
		visited[current.Path] = true
		for _, dependency := range current.Dependencies {
			if child, exists := snapshot.Package(dependency); exists {
				visit(child)
			}
		}
		result = append(result, current)
	}
	for _, current := range snapshot.Packages {
		visit(current)
	}
	return result
}

func modulePath(content string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			if strings.HasPrefix(fields[1], "\"") || strings.HasPrefix(fields[1], "`") {
				value, _ := strconv.Unquote(fields[1])
				return value
			}
			return fields[1]
		}
	}
	return ""
}

func nearestModule(modules []Module, directory string) (Module, bool) {
	var nearest Module
	found := false
	for _, module := range modules {
		if module.Language == "Go" && (module.Directory == "." || directory == module.Directory || strings.HasPrefix(directory, module.Directory+"/")) {
			if !found || (nearest.Directory == "." && module.Directory != ".") || len(module.Directory) > len(nearest.Directory) {
				nearest, found = module, true
			}
		}
	}
	return nearest, found
}

func unique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return compact(result)
}

func compact(values []string) []string {
	var result []string
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
