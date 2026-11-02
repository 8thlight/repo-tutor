package repository

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

const patternWarning = "Python and JavaScript/TypeScript inventory uses literal imports and declaration patterns; dynamic imports, aliases, re-exports, and entry-point discovery may be incomplete"

type analyzer struct {
	name       string
	extensions []string
	manifests  []string
	branches   *regexp.Regexp
	manifest   func(file File) (Module, string, bool)
	scan       func(file File) scanned
	index      func(modules []Module, sources []File, owners map[string]string) resolver
}

type scanned struct {
	imports     []string
	symbols     []Symbol
	entryPoints []Location
}

type resolver struct {
	resolve     func(from, target string) string
	entryPoints func(files []File, packages map[string]*Package) []Location
}

type importSite struct {
	from   string
	target string
}

var analyzers = []analyzer{pythonAnalyzer, javascriptAnalyzer}

func analyzerFor(name string) (analyzer, bool) {
	extension := strings.ToLower(path.Ext(name))
	for _, candidate := range analyzers {
		for _, supported := range candidate.extensions {
			if supported == extension {
				return candidate, true
			}
		}
	}
	return analyzer{}, false
}

func analyzerNamed(language string) (analyzer, bool) {
	for _, candidate := range analyzers {
		if candidate.name == language {
			return candidate, true
		}
	}
	return analyzer{}, false
}

func manifestFile(name string) bool {
	for _, candidate := range analyzers {
		for _, manifest := range candidate.manifests {
			if manifest == name {
				return true
			}
		}
	}
	return false
}

func sourceExtension(name string) bool {
	_, supported := analyzerFor(name)
	return supported
}

func packageIdentifier(language, directory string) string {
	return strings.ToLower(language) + ":" + directory
}

func readManifests(snapshot Snapshot) (Snapshot, []Module) {
	var modules []Module
	for _, file := range snapshot.Files {
		for _, candidate := range analyzers {
			if !contains(candidate.manifests, path.Base(file.Path)) {
				continue
			}
			module, warning, found := candidate.manifest(file)
			if warning != "" {
				snapshot.Warnings = append(snapshot.Warnings, warning)
			}
			if found {
				modules = append(modules, module)
			}
		}
	}
	snapshot.Modules = append(snapshot.Modules, modules...)
	return snapshot, modules
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func analyzeLanguages(snapshot Snapshot) Snapshot {
	snapshot, modules := readManifests(snapshot)
	packages := map[string]*Package{}
	owners := map[string]string{}
	sources := map[string][]File{}
	sites := map[string][]importSite{}
	present := false
	for _, file := range snapshot.Files {
		language, supported := analyzerFor(file.Path)
		if !supported {
			continue
		}
		directory := path.Dir(file.Path)
		identifier := packageIdentifier(language.name, directory)
		current, exists := packages[identifier]
		if !exists {
			current = &Package{Path: identifier, Name: path.Base(directory), Directory: directory, Language: language.name}
			packages[identifier] = current
		}
		owners[file.Path] = identifier
		if sourceTest(file.Path) {
			current.Tests = append(current.Tests, file.Path)
			continue
		}
		current.Files = append(current.Files, file.Path)
		sources[language.name] = append(sources[language.name], file)
		present = true
		found := language.scan(file)
		for _, target := range found.imports {
			current.Imports = append(current.Imports, target)
			sites[identifier] = append(sites[identifier], importSite{file.Path, target})
		}
		current.Symbols = append(current.Symbols, found.symbols...)
		snapshot.EntryPoints = append(snapshot.EntryPoints, found.entryPoints...)
	}
	resolvers := map[string]resolver{}
	for _, language := range analyzers {
		resolvers[language.name] = language.index(modules, sources[language.name], owners)
	}
	for identifier, current := range packages {
		current.Imports = unique(current.Imports)
		resolved := resolvers[current.Language]
		for _, site := range sites[identifier] {
			dependency := resolved.resolve(site.from, site.target)
			if dependency != "" && dependency != identifier {
				current.Dependencies = append(current.Dependencies, dependency)
			}
		}
		current.Dependencies = unique(current.Dependencies)
		snapshot.Packages = append(snapshot.Packages, *current)
	}
	if present {
		snapshot.Warnings = append(snapshot.Warnings, patternWarning)
	}
	var registered []Location
	for _, language := range analyzers {
		if entries := resolvers[language.name].entryPoints; entries != nil {
			registered = append(registered, entries(snapshot.Files, packages)...)
		}
	}
	snapshot.EntryPoints = append(registered, snapshot.EntryPoints...)
	sort.Slice(snapshot.Packages, func(i, j int) bool { return snapshot.Packages[i].Path < snapshot.Packages[j].Path })
	return snapshot
}

func sourceTest(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.Contains("/"+name, "/tests/") || strings.Contains("/"+name, "/test/") || strings.Contains("/"+name, "/__tests__/")
}
