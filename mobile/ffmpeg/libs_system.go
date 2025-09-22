//go:build !android

package ffmpeg

import "io/fs"

var (
	useEmbeddedLibraries = false

	libsFS fs.FS

	preloadLibs = []string{}
)
