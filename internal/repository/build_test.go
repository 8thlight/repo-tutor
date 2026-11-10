package repository_test

import (
	"strings"
	"testing"

	"repo-tutor/internal/repository"
)

func goSource() repository.Source {
	return repository.Source{
		Root: "/work/course",
		Files: []repository.File{
			{Path: "main.go", Content: "package main\nfunc main() {\n\tif true {\n\t}\n}\n"},
			{Path: "go.mod", Content: "module course\n"},
		},
		Warnings: []string{"large.md: skipped oversized or non-text file"},
	}
}

func TestBuildFingerprintFollowsTheContentAndNotTheOrder(t *testing.T) {
	first, err := repository.Build(goSource())
	if err != nil {
		t.Fatal(err)
	}
	reordered := goSource()
	reordered.Files[0], reordered.Files[1] = reordered.Files[1], reordered.Files[0]
	second, err := repository.Build(reordered)
	if err != nil || first.Fingerprint != second.Fingerprint {
		t.Fatalf("file order changed the fingerprint: %v", err)
	}
	changed := goSource()
	changed.Files[0].Content += "\n"
	third, err := repository.Build(changed)
	if err != nil || first.Fingerprint == third.Fingerprint {
		t.Fatalf("changed content kept the fingerprint: %v", err)
	}
}

func TestBuildKeepsTheRootAndPutsReadWarningsFirst(t *testing.T) {
	source := goSource()
	source.Files = append(source.Files, repository.File{Path: "go.mod.bak/go.mod", Content: "no directive\n"})
	snapshot, err := repository.Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Root != "/work/course" || len(snapshot.Warnings) < 2 || !strings.HasPrefix(snapshot.Warnings[0], "large.md") {
		t.Fatalf("root or warnings are wrong: %q %q", snapshot.Root, snapshot.Warnings)
	}
}

func TestBuildTurnsTheChangeLogIntoReadingHints(t *testing.T) {
	source := goSource()
	source.ChangeLog = "main.go\x00main.go\x00go.mod\x00"
	source.History = true
	snapshot, err := repository.Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.History || len(snapshot.Hotspots) == 0 || snapshot.Hotspots[0].Path != "main.go" || snapshot.Hotspots[0].Commits != 2 {
		t.Fatalf("the change log did not reach the hints: %+v", snapshot.Hotspots)
	}
}

func TestBuildReportsASourceWithNothingToRead(t *testing.T) {
	snapshot, err := repository.Build(repository.Source{Root: "/work/empty", Files: []repository.File{{Path: "notes.txt", Content: "hi\n"}}})
	if err == nil || !strings.Contains(err.Error(), "/work/empty") || len(snapshot.Hotspots) != 0 {
		t.Fatalf("an empty source was accepted: %v %+v", err, snapshot)
	}
}
