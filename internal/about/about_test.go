package about

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestBuildFactsComeFromBuildInfoAndDegradeToUnknown(t *testing.T) {
	info := &debug.BuildInfo{GoVersion: "go1.26.0", Main: debug.Module{Version: "v1.2.3"}, Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123"},
		{Key: "vcs.time", Value: "2026-10-26T10:00:00Z"},
		{Key: "vcs.modified", Value: "true"},
	}}
	text := FromBuildInfo(info, true, "darwin", "arm64").Text()
	for _, expected := range []string{"repo-tutor v1.2.3", "Revision 0123456789ab (modified)", "2026-10-26T10:00:00Z", "go1.26.0 on darwin/arm64", "Copyright © 2026 8th Light, LLC.", "Licensed under the GNU General Public License, version 3 or later.", "https://8thlight.com"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q:\n%s", expected, text)
		}
	}
	local := FromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true, "linux", "amd64")
	if local.Version != developmentVersion || local.Revision != unknown || local.Time != unknown {
		t.Fatalf("a build without version control data should say unknown: %+v", local)
	}
	if missing := FromBuildInfo(nil, false, "linux", "amd64"); missing.GoVersion != unknown || missing.System != "linux/amd64" {
		t.Fatalf("a binary without build info must still describe itself: %+v", missing)
	}
}

func TestLinkerVersionOverridesBuildInfo(t *testing.T) {
	version = "v0.2.0"
	t.Cleanup(func() { version = "" })
	if got := Read().Version; got != "v0.2.0" {
		t.Fatalf("version %q, want the linker value", got)
	}
}

func TestNoticesAreCurrent(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "NOTICES.txt")
	command := exec.Command("go", "run", "./internal/about/generate", generated)
	command.Dir = "../.."
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != Notices() {
		t.Fatal("NOTICES.txt is out of date; run just notices")
	}
}

func TestSplashDrawsOnItsOwnBlackCanvasSoItReadsOnAnyTerminal(t *testing.T) {
	for index, line := range strings.Split(Splash(), "\n") {
		if !strings.HasPrefix(line, splashCanvas) || !strings.HasSuffix(line, colorReset) {
			t.Fatalf("line %d does not start on the canvas and end with a reset: %q", index, line)
		}
		if strings.Count(line, colorReset) != strings.Count(line, colorReset+splashCanvas)+1 {
			t.Fatalf("line %d resets the canvas without restoring it: %q", index, line)
		}
	}
}

func TestNoticesCoverEveryLinkedModule(t *testing.T) {
	data, err := exec.Command("go", "list", "-deps", "-f", "{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}", "../../binary").Output()
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, path := range strings.Fields(string(data)) {
		checked++
		if !strings.Contains(Notices(), "\n"+path+" ") {
			t.Fatalf("%s is linked into the binary but missing from NOTICES.txt; run just notices", path)
		}
	}
	if checked == 0 || !strings.Contains(Notices(), "MIT License") {
		t.Fatal("the notices should list modules and carry their license text")
	}
}

func TestOnlyReleaseTagsReplaceTheDevelopmentVersion(t *testing.T) {
	for version, want := range map[string]string{
		"v1.2.3":       "v1.2.3",
		"v1.2.3+dirty": "v1.2.3",
		"v1.2.3-rc.1":  "v1.2.3-rc.1",
		"v0.0.0-20261026154011-381cd4d0b81a+dirty": developmentVersion,
		"v0.1.1-0.20261026154011-381cd4d0b81a":     developmentVersion,
		"(devel)":                                  developmentVersion,
		"":                                         developmentVersion,
	} {
		if got := FromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: version}}, true, "linux", "amd64").Version; got != want {
			t.Fatalf("%q read as %q, want %q", version, got, want)
		}
	}
}

func TestNoticesStartWithTheGoStandardLibraryLicense(t *testing.T) {
	text := Notices()
	library := strings.Index(text, "\nGo standard library\n")
	firstModule := strings.Index(text, "\ncharm.land/")
	if library < 0 || firstModule < 0 || library > firstModule {
		t.Fatalf("the Go standard library should be listed before the modules: %d %d", library, firstModule)
	}
	for _, expected := range []string{"The Go Authors", "Redistribution and use in source and binary forms"} {
		if !strings.Contains(text[library:firstModule], expected) {
			t.Fatalf("the standard library entry should carry its license text, missing %q", expected)
		}
	}
}
