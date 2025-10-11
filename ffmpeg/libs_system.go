//go:build !android

package ffmpeg

import "io/fs"

var (
	useEmbeddedLibraries   = false
	hasSelfContainedTmpDir = false

	libsFS fs.FS

	preloadLibs = []string{}
)
