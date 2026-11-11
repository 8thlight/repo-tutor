package reader

import (
	"runtime"
	"runtime/debug"

	"repo-tutor/internal/about"
)

var version string

func Read() about.Build {
	info, found := debug.ReadBuildInfo()
	build := about.FromBuildInfo(aboutBuildInfo(info), found, runtime.GOOS, runtime.GOARCH)
	if version != "" {
		build.Version = version
	}
	return build
}

func aboutBuildInfo(info *debug.BuildInfo) *about.BuildInfo {
	if info == nil {
		return nil
	}
	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return &about.BuildInfo{GoVersion: info.GoVersion, MainVersion: info.Main.Version, Settings: settings}
}
