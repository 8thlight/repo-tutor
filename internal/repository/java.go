package repository

import (
	"path"
	"regexp"
	"strings"
)

const javaGaps = "Java inventory uses literal imports and declaration patterns; fully qualified names without an import, reflection, dependency injection, and Maven or Gradle module links are not followed, tests outside src/test read as source, and branch counts include keywords inside comments and string literals"

var javaPomArtifact = regexp.MustCompile(`<artifactId>\s*([^<\s]+)\s*</artifactId>`)
var javaPomParent = regexp.MustCompile(`(?s)<parent>.*?</parent>`)
var javaGradleName = regexp.MustCompile(`(?m)^\s*rootProject\.name\s*=\s*["']([^"']+)["']`)
var javaPackage = regexp.MustCompile(`^\s*package\s+([\w.]+)\s*;`)
var javaImport = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([\w.]+(?:\.\*)?)\s*;`)
var javaType = regexp.MustCompile(`^\s*(?:(?:public|protected|private|abstract|static|final|sealed|non-sealed|strictfp)\s+)*(class|interface|enum|record|@interface)\s+(\w+)`)
var javaMethod = regexp.MustCompile(`^\s*(?:(?:public|protected|private|static|final|abstract|synchronized|default|native)\s+)+(?:<[^>]+>\s+)?[\w<>\[\],.? ]+?\s+(\w+)\s*\(`)
var javaMain = regexp.MustCompile(`\bstatic\s+void\s+main\s*\(`)
var javaBranch = regexp.MustCompile(`\b(if|for|while|case|catch)\b|&&|\|\|`)

var javaAnalyzer = analyzer{
	name:       "Java",
	extensions: []string{".java"},
	manifests:  []string{"pom.xml", "settings.gradle", "settings.gradle.kts"},
	branches:   javaBranch,
	gaps:       javaGaps,
	manifest:   javaManifest,
	scan:       javaScan,
	index:      javaIndex,
}

func javaManifest(file File) (Module, string, bool) {
	var name string
	if path.Base(file.Path) == "pom.xml" {
		if match := javaPomArtifact.FindStringSubmatch(javaPomParent.ReplaceAllString(file.Content, "")); match != nil {
			name = match[1]
		}
	} else if match := javaGradleName.FindStringSubmatch(file.Content); match != nil {
		name = match[1]
	}
	if name == "" {
		return Module{}, "", false
	}
	return Module{name, path.Dir(file.Path), "Java"}, "", true
}

func javaScan(file File) scanned {
	var found scanned
	for index, line := range strings.Split(file.Content, "\n") {
		location := Location{file.Path, index + 1}
		if match := javaImport.FindStringSubmatch(line); match != nil {
			found.imports = append(found.imports, match[1])
		}
		if match := javaType.FindStringSubmatch(line); match != nil {
			found.symbols = append(found.symbols, Symbol{match[2], "type", location})
		} else if match := javaMethod.FindStringSubmatch(line); match != nil {
			found.symbols = append(found.symbols, Symbol{match[1], "function", location})
		}
		if javaMain.MatchString(line) {
			found.entryPoints = append(found.entryPoints, location)
		}
	}
	return found
}

func javaIndex(_ []Module, sources []File, _ map[string]string) resolver {
	classes := map[string]string{}
	packages := map[string]string{}
	for _, file := range sources {
		name := javaPackageName(file)
		if name == "" {
			continue
		}
		identifier := packageIdentifier("Java", path.Dir(file.Path))
		javaRegister(classes, name+"."+strings.TrimSuffix(path.Base(file.Path), ".java"), identifier)
		javaRegister(packages, name, identifier)
	}
	return resolver{resolve: func(_, target string) string {
		wildcard := strings.HasSuffix(target, ".*")
		name := strings.TrimSuffix(target, ".*")
		for attempt := 0; attempt < 3; attempt++ {
			if identifier := classes[name]; identifier != "" {
				return identifier
			}
			if wildcard && attempt == 0 {
				return packages[name]
			}
			cut := strings.LastIndex(name, ".")
			if cut < 0 {
				break
			}
			name = name[:cut]
		}
		return ""
	}}
}

func javaPackageName(file File) string {
	for _, line := range strings.Split(file.Content, "\n") {
		if match := javaPackage.FindStringSubmatch(line); match != nil {
			return match[1]
		}
	}
	return ""
}

func javaRegister(index map[string]string, key, identifier string) {
	if existing, exists := index[key]; exists && existing != identifier {
		index[key] = ""
		return
	}
	index[key] = identifier
}
