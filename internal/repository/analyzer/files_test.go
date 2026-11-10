package analyzer_test

import (
	"testing"

	"repo-tutor/internal/repository/analyzer"
)

func TestExcludedDirectoriesHoldNoSourceToLearnFrom(t *testing.T) {
	for _, name := range []string{".git", ".scratch", "vendor", "node_modules", "output", "dist", "build", "testdata", "fixtures", "__pycache__", "coverage"} {
		if !analyzer.Excluded(name) {
			t.Errorf("%s should be excluded", name)
		}
	}
	for _, name := range []string{"internal", "src", "cmd", "docs"} {
		if analyzer.Excluded(name) {
			t.Errorf("%s should be read", name)
		}
	}
}

func TestSupportedFilesAreSourceManifestsAndDocuments(t *testing.T) {
	for _, name := range []string{"main.go", "go.mod", "README.md", "app.py", "index.ts", "Main.java", "package.json"} {
		if !analyzer.Supported(name) {
			t.Errorf("%s should be read", name)
		}
	}
	for _, name := range []string{".hidden.go", "image.png", "notes.txt", "Makefile", "archive.zip"} {
		if analyzer.Supported(name) {
			t.Errorf("%s should be skipped", name)
		}
	}
}
