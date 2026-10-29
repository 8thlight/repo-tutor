package about

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	Name               = "repo-tutor"
	Copyright          = "Copyright © 2026 8th Light, LLC."
	License            = "Licensed under the GNU General Public License, version 3 or later."
	Website            = "https://8thlight.com"
	developmentVersion = "0.1.0-dev"
	unknown            = "unknown"
)

//go:embed splash.ansi
var splash string

const (
	SplashWidth   = 80
	SplashCaption = "Prince Charles Louis of the Palatinate with his Tutor Wolrad von Plessen in Historical Dress. Jan Lievens, 1631. Public domain, from the J. Paul Getty Museum via Google Art Project."
)

const (
	splashCanvas = "\x1b[48;2;0;0;0m"
	colorReset   = "\x1b[0m"
)

func Splash() string {
	lines := strings.Split(strings.TrimRight(splash, "\n"), "\n")
	for index, line := range lines {
		lines[index] = splashCanvas + strings.ReplaceAll(line, colorReset, colorReset+splashCanvas) + colorReset
	}
	return strings.Join(lines, "\n")
}

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

var version string

func Read() Build {
	info, found := debug.ReadBuildInfo()
	build := FromBuildInfo(info, found, runtime.GOOS, runtime.GOARCH)
	if version != "" {
		build.Version = version
	}
	return build
}

func FromBuildInfo(info *debug.BuildInfo, found bool, goos, goarch string) Build {
	build := Build{Version: developmentVersion, Revision: unknown, Time: unknown, GoVersion: unknown, System: goos + "/" + goarch}
	if !found || info == nil {
		return build
	}
	build.GoVersion = info.GoVersion
	if tagged := taggedVersion(info.Main.Version); tagged != "" {
		build.Version = tagged
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

var releaseVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?$`)

var pseudoVersion = regexp.MustCompile(`\d{14}-[0-9a-f]{12}$`)

func taggedVersion(version string) string {
	version, _, _ = strings.Cut(version, "+")
	if !releaseVersion.MatchString(version) || pseudoVersion.MatchString(version) {
		return ""
	}
	return version
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
		License,
		Website,
	}
}

func (build Build) Text() string {
	return strings.Join(build.Lines(), "\n") + "\n"
}
