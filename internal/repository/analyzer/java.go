package analyzer

import (
	"path"
	"regexp"
	"strings"
)

const javaGaps = "Java inventory uses literal imports and declaration patterns; fully qualified names without an import, reflection, dependency injection, and Maven or Gradle module links are not followed, tests that start the whole system with a framework other than Spring Boot are not grouped, and branch counts include keywords inside string literals"

var javaPomArtifact = regexp.MustCompile(`<artifactId>\s*([^<\s]+)\s*</artifactId>`)
var javaPomParent = regexp.MustCompile(`(?s)<parent>.*?</parent>`)
var javaGradleName = regexp.MustCompile(`(?m)^\s*rootProject\.name\s*=\s*["']([^"']+)["']`)
var javaPackage = regexp.MustCompile(`^\s*package\s+([\w.]+)\s*;`)
var javaImport = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([\w.]+(?:\.\*)?)\s*;`)
var javaType = regexp.MustCompile(`^\s*(?:(?:public|protected|private|abstract|static|final|sealed|non-sealed|strictfp)\s+)*(class|interface|enum|record|@interface)\s+(\w+)`)
var javaMethod = regexp.MustCompile(`^\s*(?:(?:public|protected|private|static|final|abstract|synchronized|default|native)\s+)+(?:<[^>]+>\s+)?[\w<>\[\],.? ]+?\s+(\w+)\s*\(`)
var javaSliceTest = regexp.MustCompile(`@(?:WebMvcTest|WebFluxTest)\s*\(\s*(?:(?:value|controllers)\s*=\s*)?(?:\{\s*)?([\w.]+)\.class\s*\}?\s*\)`)
var javaSystemTest = regexp.MustCompile(`@SpringBootTest\b`)
var javaApplication = regexp.MustCompile(`@SpringBootApplication\b`)
var javaMain = regexp.MustCompile(`\bstatic\s+void\s+main\s*\(`)
var javaTestName = regexp.MustCompile(`(?:Tests?|[a-z0-9]IT)$`)
var javaBranch = regexp.MustCompile(`\b(if|for|while|case|catch)\b|&&|\|\|`)

var javaAnalyzer = analyzer{
	name:        "Java",
	extensions:  []string{".java"},
	manifests:   []string{"pom.xml", "settings.gradle", "settings.gradle.kts"},
	branches:    javaBranch,
	gaps:        javaGaps,
	test:        javaTest,
	system:      javaSystem,
	uncommented: javaWithoutComments,
	manifest:    javaManifest,
	scan:        javaScan,
	index:       javaIndex,
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

func javaTest(name string) bool {
	if javaTestName.MatchString(strings.TrimSuffix(path.Base(name), ".java")) {
		return true
	}
	for _, directory := range strings.Split(path.Dir(name), "/") {
		if directory == "test" || directory == "integrationTest" || directory == "integration-test" || directory == "androidTest" {
			return true
		}
	}
	return false
}

func javaSystem(test File) bool {
	return javaSystemTest.MatchString(test.Content)
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
	applications := map[string]bool{}
	for _, file := range sources {
		name := javaPackageName(file)
		if name == "" {
			continue
		}
		identifier := packageIdentifier("Java", path.Dir(file.Path))
		if javaApplication.MatchString(file.Content) {
			applications[identifier] = true
		}
		javaRegister(classes, name+"."+strings.TrimSuffix(path.Base(file.Path), ".java"), identifier)
		javaRegister(packages, name, identifier)
	}
	return resolver{owner: func(test File) string {
		for _, candidate := range javaSliceTargets(test) {
			if identifier := classes[candidate]; identifier != "" {
				return identifier
			}
		}
		identifier := packages[javaPackageName(test)]
		if applications[identifier] && javaSystem(test) {
			return ""
		}
		return identifier
	}, resolve: func(_, target string) string {
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

func javaSliceTargets(test File) []string {
	match := javaSliceTest.FindStringSubmatch(test.Content)
	if match == nil {
		return nil
	}
	name := match[1]
	if strings.Contains(name, ".") {
		return []string{name}
	}
	targets := []string{javaPackageName(test) + "." + name}
	for _, line := range strings.Split(test.Content, "\n") {
		if imported := javaImport.FindStringSubmatch(line); imported != nil {
			if wildcard, isWildcard := strings.CutSuffix(imported[1], ".*"); isWildcard {
				targets = append(targets, wildcard+"."+name)
			} else if strings.HasSuffix(imported[1], "."+name) {
				targets = append(targets, imported[1])
			}
		}
	}
	return targets
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

func javaWithoutComments(content string) string {
	result := []byte(content)
	for position := 0; position < len(content); {
		rest := content[position:]
		switch {
		case strings.HasPrefix(rest, "//"):
			end := strings.Index(rest, "\n")
			if end < 0 {
				end = len(rest)
			}
			javaBlank(result, position, position+end)
			position += end
		case strings.HasPrefix(rest, "/*"):
			end := len(rest)
			if close := strings.Index(rest[2:], "*/"); close >= 0 {
				end = close + 4
			}
			javaBlank(result, position, position+end)
			position += end
		case strings.HasPrefix(rest, `"""`):
			position += javaLiteralLength(rest, `"""`)
		case rest[0] == '"' || rest[0] == '\'':
			position += javaLiteralLength(rest, rest[:1])
		default:
			position++
		}
	}
	return string(result)
}

func javaLiteralLength(rest, quote string) int {
	for offset := len(quote); offset < len(rest); offset++ {
		switch {
		case rest[offset] == '\\':
			offset++
		case strings.HasPrefix(rest[offset:], quote):
			return offset + len(quote)
		case rest[offset] == '\n' && len(quote) == 1:
			return offset
		}
	}
	return len(rest)
}

func javaBlank(result []byte, from, to int) {
	for index := from; index < to; index++ {
		if result[index] != '\n' && result[index] != '\r' {
			result[index] = ' '
		}
	}
}
