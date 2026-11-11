package reader

import (
	"runtime/debug"
	"testing"

	"repo-tutor/internal/about"
)

func TestLinkerVersionOverridesBuildInfo(t *testing.T) {
	version = "v0.2.0"
	t.Cleanup(func() { version = "" })
	if got := Read().Version; got != "v0.2.0" {
		t.Fatalf("version %q, want the linker value", got)
	}
}

func TestConversionKeepsVersionControlSettings(t *testing.T) {
	info := aboutBuildInfo(&debug.BuildInfo{GoVersion: "go1.26.0", Main: debug.Module{Version: "v1.2.3"}, Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123"},
		{Key: "vcs.time", Value: "2026-10-26T10:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}})
	build := about.FromBuildInfo(info, true, "linux", "amd64")
	if build.Revision != "0123456789abcdef0123" || build.Time != "2026-10-26T10:00:00Z" || !build.Modified || build.Version != "v1.2.3" || build.GoVersion != "go1.26.0" {
		t.Fatalf("conversion lost build facts: %+v", build)
	}
}
