package repository

import (
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const patternWarning = "Python and JavaScript/TypeScript inventory uses literal imports and declaration patterns; dynamic imports, aliases, re-exports, and entry-point discovery may be incomplete"

type analyzer struct {
	name        string
	extensions  []string
	manifests   []string
	branches    *regexp.Regexp
	gaps        string
	test        func(name string) bool
	system      func(test File) bool
	uncommented func(content string) string
	manifest    func(file File) (Module, string, bool)
	scan        func(file File) scanned
	index       func(modules []Module, sources []File, owners map[string]string) resolver
}

type scanned struct {
	imports     []string
	symbols     []Symbol
	entryPoints []Location
}

type resolver struct {
	resolve     func(from, target string) string
	entryPoints func(files []File, packages map[string]*Package) []Location
	owner       func(test File) string
}

type directoryTest struct {
	file      File
	directory string
}

type importSite struct {
	from   string
	target string
}

var analyzers = []analyzer{pythonAnalyzer, javascriptAnalyzer, javaAnalyzer}

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

func (candidate analyzer) isTest(name string) bool {
	if candidate.test != nil {
		return candidate.test(name)
	}
	return sourceTest(name)
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
	tests := map[string][]directoryTest{}
	sites := map[string][]importSite{}
	present := map[string]bool{}
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
		if language.isTest(file.Path) {
			current.Tests = append(current.Tests, file.Path)
			tests[language.name] = append(tests[language.name], directoryTest{file, identifier})
			continue
		}
		current.Files = append(current.Files, file.Path)
		sources[language.name] = append(sources[language.name], file)
		present[language.name] = true
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
	moveTestsToOwners(packages, tests, resolvers)
	markTestLayers(packages, tests)
	readLayerImports(packages, tests, sites)
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
	var disclosed []string
	for _, language := range analyzers {
		if present[language.name] && !contains(disclosed, language.gaps) {
			disclosed = append(disclosed, language.gaps)
			snapshot.Warnings = append(snapshot.Warnings, language.gaps)
		}
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

func moveTestsToOwners(packages map[string]*Package, tests map[string][]directoryTest, resolvers map[string]resolver) {
	adopted := map[string]bool{}
	for language, candidates := range tests {
		owner := resolvers[language].owner
		if owner == nil {
			continue
		}
		for _, test := range candidates {
			target := owner(test.file)
			production, found := packages[target]
			if !found || target == test.directory {
				continue
			}
			origin := packages[test.directory]
			origin.Tests = slices.DeleteFunc(origin.Tests, func(name string) bool { return name == test.file.Path })
			production.Tests = append(production.Tests, test.file.Path)
			adopted[target] = true
		}
	}
	for identifier := range adopted {
		sort.Strings(packages[identifier].Tests)
	}
	for identifier, current := range packages {
		if len(current.Files) == 0 && len(current.Tests) == 0 {
			delete(packages, identifier)
		}
	}
}

func markTestLayers(packages map[string]*Package, tests map[string][]directoryTest) {
	for language, candidates := range tests {
		analyzed, _ := analyzerNamed(language)
		if analyzed.system == nil {
			continue
		}
		for _, test := range candidates {
			current := packages[test.directory]
			if current != nil && len(current.Files) == 0 && slices.Contains(current.Tests, test.file.Path) && analyzed.system(test.file) {
				current.TestLayer = true
			}
		}
	}
}

func readLayerImports(packages map[string]*Package, tests map[string][]directoryTest, sites map[string][]importSite) {
	for language, candidates := range tests {
		analyzed, _ := analyzerNamed(language)
		for _, test := range candidates {
			current := packages[test.directory]
			if current == nil || !current.TestLayer || !slices.Contains(current.Tests, test.file.Path) {
				continue
			}
			for _, target := range analyzed.scan(test.file).imports {
				current.Imports = append(current.Imports, target)
				sites[test.directory] = append(sites[test.directory], importSite{test.file.Path, target})
			}
		}
	}
}

func sourceTest(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.Contains("/"+name, "/tests/") || strings.Contains("/"+name, "/test/") || strings.Contains("/"+name, "/__tests__/")
}
