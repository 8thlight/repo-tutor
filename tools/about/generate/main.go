package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var licenseNames = []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE", "COPYING", "COPYING.md"}

//go:embed go-license.txt
var goLicense string

type module struct{ path, version, directory string }

func linkedModules() ([]module, error) {
	listing := "{{with .Module}}{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}{{end}}"
	data, err := exec.Command("go", "list", "-deps", "-f", listing, "./binary").Output()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var modules []module
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) != 3 || seen[parts[0]] {
			continue
		}
		seen[parts[0]] = true
		modules = append(modules, module{parts[0], parts[1], parts[2]})
	}
	sort.Slice(modules, func(first, second int) bool { return modules[first].path < modules[second].path })
	return modules, nil
}

func licenseText(directory string) (string, error) {
	for _, name := range licenseNames {
		if data, err := os.ReadFile(filepath.Join(directory, name)); err == nil {
			return strings.TrimSpace(string(data)), nil
		}
	}
	return "", fmt.Errorf("no license file in %s", directory)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: generate OUTPUT")
		os.Exit(2)
	}
	modules, err := linkedModules()
	if err != nil {
		fmt.Fprintln(os.Stderr, "list modules:", err)
		os.Exit(1)
	}
	var notices strings.Builder
	notices.WriteString("Third-party software linked into repo-tutor, with each license. The Go standard library comes first.\n")
	fmt.Fprintf(&notices, "\n================\nGo standard library\n================\n\n%s\n", strings.TrimSpace(goLicense))
	for _, current := range modules {
		text, err := licenseText(current.directory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", current.path, err)
			os.Exit(1)
		}
		fmt.Fprintf(&notices, "\n================\n%s %s\n================\n\n%s\n", current.path, current.version, text)
	}
	if err := os.WriteFile(os.Args[1], []byte(notices.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
