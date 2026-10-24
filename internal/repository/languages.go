package repository

import (
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strings"
)

var pythonProjectName = regexp.MustCompile(`(?m)^name\s*=\s*["']([^"']+)["']`)
var pythonFrom = regexp.MustCompile(`^\s*from\s+([.\w]+)\s+import\s+`)
var pythonImport = regexp.MustCompile(`^\s*import\s+([\w.]+)`)
var pythonDeclaration = regexp.MustCompile(`^(?:async\s+)?(def|class)\s+([\w]+)`)
var javascriptImport = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*|\brequire\s*\(\s*|\bimport\s*\(\s*)["']([^"']+)["']`)
var javascriptDeclaration = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:declare\s+)?(?:async\s+)?(function|class|interface|type|enum)\s+([\w]+)`)
var pythonScript = regexp.MustCompile(`^\s*[^#=]+\s*=\s*["']([\w.]+):([\w.]+)["']`)
var pythonMain = regexp.MustCompile(`^\s*if\s+__name__\s*==\s*["']__main__["']\s*:`)

func analyzeLanguages(snapshot Snapshot) Snapshot {
	packages := map[string]*Package{}
	fileOwners := map[string]string{}
	pythonModules := map[string]string{}
	pythonNames := map[string]string{}
	nodeNames := map[string][]string{}
	var importLocations = map[string][]string{}
	hasPython, hasJavascript := false, false
	var registered []Location
	for _, file := range snapshot.Files {
		switch path.Base(file.Path) {
		case "pyproject.toml":
			if match := pythonProjectName.FindStringSubmatch(file.Content); match != nil {
				snapshot.Modules = append(snapshot.Modules, Module{match[1], path.Dir(file.Path), "Python"})
			}
		case "package.json":
			var manifest struct {
				Name string `json:"name"`
			}
			if json.Unmarshal([]byte(file.Content), &manifest) != nil {
				snapshot.Warnings = append(snapshot.Warnings, file.Path+": invalid package.json")
				continue
			}
			if manifest.Name == "" {
				manifest.Name = path.Dir(file.Path)
			}
			snapshot.Modules = append(snapshot.Modules, Module{manifest.Name, path.Dir(file.Path), "JavaScript/TypeScript"})
			nodeNames[manifest.Name] = append(nodeNames[manifest.Name], path.Dir(file.Path))
		}
	}
	for _, file := range snapshot.Files {
		language := sourceLanguage(file.Path)
		if language == "" {
			continue
		}
		directory := path.Dir(file.Path)
		identifier := strings.ToLower(language) + ":" + directory
		current, exists := packages[identifier]
		if !exists {
			current = &Package{Path: identifier, Name: path.Base(directory), Directory: directory, Language: language}
			packages[identifier] = current
		}
		fileOwners[file.Path] = identifier
		if sourceTest(file.Path) {
			current.Tests = append(current.Tests, file.Path)
			continue
		}
		current.Files = append(current.Files, file.Path)
		if language == "Python" {
			hasPython = true
			moduleName := pythonModuleName(file.Path)
			pythonModules[moduleName] = identifier
			pythonNames[file.Path] = moduleName
		} else {
			hasJavascript = true
		}
		for index, line := range strings.Split(file.Content, "\n") {
			location := Location{file.Path, index + 1}
			if language == "Python" {
				if match := pythonFrom.FindStringSubmatch(line); match != nil {
					current.Imports = append(current.Imports, match[1])
					importLocations[identifier] = append(importLocations[identifier], file.Path+"\x00"+match[1])
				} else if match := pythonImport.FindStringSubmatch(line); match != nil {
					current.Imports = append(current.Imports, match[1])
					importLocations[identifier] = append(importLocations[identifier], file.Path+"\x00"+match[1])
				}
				if match := pythonDeclaration.FindStringSubmatch(line); match != nil {
					kind := "function"
					if match[1] == "class" {
						kind = "type"
					}
					current.Symbols = append(current.Symbols, Symbol{match[2], kind, location})
				}
				if pythonMain.MatchString(line) {
					snapshot.EntryPoints = append(snapshot.EntryPoints, location)
				}
			} else {
				for _, match := range javascriptImport.FindAllStringSubmatch(line, -1) {
					current.Imports = append(current.Imports, match[1])
					importLocations[identifier] = append(importLocations[identifier], file.Path+"\x00"+match[1])
				}
				if match := javascriptDeclaration.FindStringSubmatch(line); match != nil {
					kind := "type"
					if match[1] == "function" {
						kind = "function"
					}
					current.Symbols = append(current.Symbols, Symbol{match[2], kind, location})
				}
			}
		}
	}
	for identifier, current := range packages {
		current.Imports = unique(current.Imports)
		for _, imported := range importLocations[identifier] {
			parts := strings.SplitN(imported, "\x00", 2)
			name, target := parts[0], parts[1]
			dependency := ""
			if current.Language == "Python" {
				if strings.HasPrefix(target, ".") {
					level := len(target) - len(strings.TrimLeft(target, "."))
					base := strings.Split(pythonNames[name], ".")
					if path.Base(name) == "__init__.py" {
						level--
					}
					if level <= len(base) {
						prefix := strings.Join(base[:len(base)-level], ".")
						target = strings.Trim(prefix+"."+strings.TrimLeft(target, "."), ".")
					}
				}
				dependency = pythonModules[target]
			} else if strings.HasPrefix(target, ".") {
				dependency = javascriptTarget(path.Join(path.Dir(name), target), fileOwners)
			} else {
				if directories := nodeNames[target]; len(directories) == 1 {
					for _, entry := range []string{"src/index", "index"} {
						if resolved := javascriptTarget(path.Join(directories[0], entry), fileOwners); resolved != "" {
							dependency = resolved
							break
						}
					}
				}
			}
			if dependency != "" && dependency != identifier {
				current.Dependencies = append(current.Dependencies, dependency)
			}
		}
		current.Dependencies = unique(current.Dependencies)
		snapshot.Packages = append(snapshot.Packages, *current)
	}
	if hasPython || hasJavascript {
		snapshot.Warnings = append(snapshot.Warnings, "Python and JavaScript/TypeScript inventory uses literal imports and declaration patterns; dynamic imports, aliases, re-exports, and entry-point discovery may be incomplete")
	}
	for _, file := range snapshot.Files {
		if path.Base(file.Path) != "pyproject.toml" {
			continue
		}
		inScripts := false
		for index, line := range strings.Split(file.Content, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") {
				inScripts = trimmed == "[project.scripts]"
				continue
			}
			if !inScripts {
				continue
			}
			if match := pythonScript.FindStringSubmatch(line); match != nil {
				registered = append(registered, Location{file.Path, index + 1})
				if current, exists := packages[pythonModules[match[1]]]; exists {
					for _, symbol := range current.Symbols {
						if symbol.Name == match[2] && pythonNames[symbol.Path] == match[1] {
							registered = append(registered, symbol.Location)
						}
					}
				}
			}
		}
	}
	snapshot.EntryPoints = append(registered, snapshot.EntryPoints...)
	sort.Slice(snapshot.Packages, func(i, j int) bool { return snapshot.Packages[i].Path < snapshot.Packages[j].Path })
	return snapshot
}

func sourceLanguage(name string) string {
	switch path.Ext(name) {
	case ".py":
		return "Python"
	case ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx":
		return "JavaScript/TypeScript"
	}
	return ""
}

func sourceTest(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.Contains("/"+name, "/tests/") || strings.Contains("/"+name, "/test/") || strings.Contains("/"+name, "/__tests__/")
}

func pythonModuleName(name string) string {
	name = strings.TrimSuffix(name, ".py")
	if index := strings.LastIndex(name, "/src/"); index >= 0 {
		name = name[index+5:]
	} else {
		name = strings.TrimPrefix(name, "src/")
	}
	name = strings.TrimSuffix(name, "/__init__")
	return strings.ReplaceAll(name, "/", ".")
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
