//go:build !android

package ffmpeg

import (
	"embed"
)

var (
	useEmbeddedLibraries = false

	libsFS embed.FS

	preloadLibs = []string{}
)
