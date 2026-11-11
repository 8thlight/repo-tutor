package reader

import (
	"runtime"
	"runtime/debug"

	"repo-tutor/internal/about"
)

var version string

func Read() about.Build {
	info, found := debug.ReadBuildInfo()
	build := about.FromBuildInfo(info, found, runtime.GOOS, runtime.GOARCH)
	if version != "" {
		build.Version = version
	}
	return build
}
