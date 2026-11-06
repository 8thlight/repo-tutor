package repository

import (
	"path"
	"regexp"
	"strings"
)

var pythonProjectName = regexp.MustCompile(`(?m)^name\s*=\s*["']([^"']+)["']`)
var pythonFrom = regexp.MustCompile(`^\s*from\s+([.\w]+)\s+import\s+`)
var pythonImport = regexp.MustCompile(`^\s*import\s+([\w.]+)`)
var pythonDeclaration = regexp.MustCompile(`^(?:async\s+)?(def|class)\s+([\w]+)`)
var pythonScript = regexp.MustCompile(`^\s*[^#=]+\s*=\s*["']([\w.]+):([\w.]+)["']`)
var pythonMain = regexp.MustCompile(`^\s*if\s+__name__\s*==\s*["']__main__["']\s*:`)
var pythonBranch = regexp.MustCompile(`\b(if|elif|for|while|except|case|and|or)\b`)

var pythonAnalyzer = analyzer{
	name:        "Python",
	extensions:  []string{".py"},
	manifests:   []string{"pyproject.toml"},
	branches:    pythonBranch,
	gaps:        patternWarning,
	nestedTests: true,
	importOwned: true,
	manifest:    pythonManifest,
	scan:        pythonScan,
	index:       pythonIndex,
}

func pythonManifest(file File) (Module, string, bool) {
	match := pythonProjectName.FindStringSubmatch(file.Content)
	if match == nil {
		return Module{}, "", false
	}
	return Module{match[1], path.Dir(file.Path), "Python"}, "", true
}

func pythonScan(file File) scanned {
	var found scanned
	for index, line := range strings.Split(file.Content, "\n") {
		location := Location{file.Path, index + 1}
		if match := pythonFrom.FindStringSubmatch(line); match != nil {
			found.imports = append(found.imports, match[1])
		} else if match := pythonImport.FindStringSubmatch(line); match != nil {
			found.imports = append(found.imports, match[1])
		}
		if match := pythonDeclaration.FindStringSubmatch(line); match != nil {
			kind := "function"
			if match[1] == "class" {
				kind = "type"
			}
			found.symbols = append(found.symbols, Symbol{match[2], kind, location})
		}
		if pythonMain.MatchString(line) {
			found.entryPoints = append(found.entryPoints, location)
		}
	}
	return found
}

func pythonIndex(_ []Module, sources []File, _ map[string]string) resolver {
	modules := map[string]string{}
	names := map[string]string{}
	for _, file := range sources {
		name := pythonModuleName(file.Path)
		modules[name] = packageIdentifier("Python", path.Dir(file.Path))
		names[file.Path] = name
	}
	return resolver{
		resolve: func(from, target string) string {
			if strings.HasPrefix(target, ".") {
				level := len(target) - len(strings.TrimLeft(target, "."))
				base := strings.Split(names[from], ".")
				if path.Base(from) == "__init__.py" {
					level--
				}
				if level <= len(base) {
					prefix := strings.Join(base[:len(base)-level], ".")
					target = strings.Trim(prefix+"."+strings.TrimLeft(target, "."), ".")
				}
			}
			return modules[target]
		},
		entryPoints: func(files []File, packages map[string]*Package) []Location {
			return pythonRegistered(files, packages, modules, names)
		},
	}
}

func pythonRegistered(files []File, packages map[string]*Package, modules, names map[string]string) []Location {
	var registered []Location
	for _, file := range files {
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
				if current, exists := packages[modules[match[1]]]; exists {
					for _, symbol := range current.Symbols {
						if symbol.Name == match[2] && names[symbol.Path] == match[1] {
							registered = append(registered, symbol.Location)
						}
					}
				}
			}
		}
	}
	return registered
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
