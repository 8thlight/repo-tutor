package about

import (
	_ "embed"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	Name               = "repo-tutor"
	Copyright          = "Copyright © 2026 Eighth Light LLC."
	Website            = "https://8thlight.com"
	developmentVersion = "0.1.0-dev"
	unknown            = "unknown"
)

//go:embed NOTICES.txt
var notices string

func Notices() string {
	return notices
}

type Build struct {
	Version   string
	Revision  string
	Time      string
	Modified  bool
	GoVersion string
	System    string
}

func Read() Build {
	info, found := debug.ReadBuildInfo()
	return FromBuildInfo(info, found, runtime.GOOS, runtime.GOARCH)
}

func FromBuildInfo(info *debug.BuildInfo, found bool, goos, goarch string) Build {
	build := Build{Version: developmentVersion, Revision: unknown, Time: unknown, GoVersion: unknown, System: goos + "/" + goarch}
	if !found || info == nil {
		return build
	}
	build.GoVersion = info.GoVersion
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		build.Version = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			build.Revision = setting.Value
		case "vcs.time":
			build.Time = setting.Value
		case "vcs.modified":
			build.Modified = setting.Value == "true"
		}
	}
	return build
}

func (build Build) ShortRevision() string {
	revision := build.Revision
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if build.Modified {
		revision += " (modified)"
	}
	return revision
}

func (build Build) Lines() []string {
	return []string{
		fmt.Sprintf("%s %s", Name, build.Version),
		fmt.Sprintf("Revision %s", build.ShortRevision()),
		fmt.Sprintf("Built from commit time %s", build.Time),
		fmt.Sprintf("%s on %s", build.GoVersion, build.System),
		Copyright,
		Website,
	}
}

func (build Build) Text() string {
	return strings.Join(build.Lines(), "\n") + "\n"
}
