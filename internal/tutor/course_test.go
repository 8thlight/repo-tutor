package tutor_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"repo-tutor/internal/repository/analyzer"
	"repo-tutor/internal/tutor"
)

func TestPackagesNamedLikeTheSpecialTopicsKeepTheirOwnIds(t *testing.T) {
	snapshot := analyzer.Analyze([]analyzer.File{
		{Path: "go.mod", Content: "module overview\n"},
		{Path: "main.go", Content: "package main\nfunc main() {}\n"},
		{Path: "hotspots/hotspots.go", Content: "package hotspots\nfunc Rank() {}\n"},
	})
	snapshot.Hotspots = []analyzer.Hotspot{{Path: "main.go", Line: 2, Language: "go", Commits: 1, Branches: 1, Lines: 2, Score: 4}}
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

func TestATestLayerIsATopicAfterTheCodeItExercises(t *testing.T) {
	snapshot := analyzer.Analyze([]analyzer.File{
		{Path: "src/main/java/com/acme/app/Main.java", Content: "package com.acme.app;\nclass Main {}\n"},
		{Path: "src/test/java/com/acme/integration/FlowTest.java", Content: "package com.acme.integration;\nimport com.acme.app.Main;\n@SpringBootTest\nclass FlowTest {}\n"},
		{Path: "src/test/java/com/acme/loose/LooseTest.java", Content: "package com.acme.loose;\nclass LooseTest {}\n"},
	})
	var ids []string
	for _, topic := range tutor.Course(snapshot) {
		ids = append(ids, topic.ID)
	}
	want := []string{tutor.Overview, "java:src/main/java/com/acme/app", "java:src/test/java/com/acme/integration"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("wrong course: %v", ids)
	}
	context := tutor.Prepare(snapshot, tutor.Request{Action: "lesson", Topic: "java:src/test/java/com/acme/integration"})
	var paths []string
	for _, excerpt := range context.Excerpts {
		paths = append(paths, excerpt.Path)
	}
	for _, expected := range []string{"src/test/java/com/acme/integration/FlowTest.java", "src/main/java/com/acme/app/Main.java"} {
		if !slices.Contains(paths, expected) {
			t.Fatalf("the test layer lesson lacks %s: %v", expected, paths)
		}
	}
	if !strings.Contains(context.Prompt, "test_layer true is a test layer") || !strings.Contains(context.Prompt, `"test_layer":true`) {
		t.Fatal("the prompt does not teach a test layer")
	}
}

func TestNoPackagePathCanStartWithTheReservedPrefix(t *testing.T) {
	for _, id := range []string{tutor.Overview, tutor.HotspotsTopic} {
		if id[0] != ':' {
			t.Fatalf("%q should start with a colon, which no package path can", id)
		}
	}
}
