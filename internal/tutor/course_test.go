package tutor_test

import (
	"testing"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

func TestPackagesNamedLikeTheSpecialTopicsKeepTheirOwnIds(t *testing.T) {
	snapshot := repository.Analyze([]repository.File{
		{Path: "go.mod", Content: "module overview\n"},
		{Path: "main.go", Content: "package main\nfunc main() {}\n"},
		{Path: "hotspots/hotspots.go", Content: "package hotspots\nfunc Rank() {}\n"},
	})
	snapshot.Hotspots = []repository.Hotspot{{Path: "main.go", Line: 2, Language: "go", Commits: 1, Branches: 1, Lines: 2, Score: 4}}
	seen := map[string]int{}
	for _, topic := range tutor.Course(snapshot) {
		seen[topic.ID]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Fatalf("topic id %q appears %d times", id, count)
		}
	}
	if seen[tutor.Overview] != 1 || seen[tutor.HotspotsTopic] != 1 {
		t.Fatalf("the overview and the hotspots topic should both be present once: %v", seen)
	}
	if tutor.Overview == "overview" || tutor.HotspotsTopic == "hotspots" {
		t.Fatal("the special ids must differ from plain package names")
	}
}

func TestNoPackagePathCanStartWithTheReservedPrefix(t *testing.T) {
	for _, id := range []string{tutor.Overview, tutor.HotspotsTopic} {
		if id[0] != ':' {
			t.Fatalf("%q should start with a colon, which no package path can", id)
		}
	}
}
