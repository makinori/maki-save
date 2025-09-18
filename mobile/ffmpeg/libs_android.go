//go:build android

package ffmpeg

import (
	"embed"
)

var (
	useEmbeddedLibraries = true

	//go:embed android-arm64-v8a
	libsFS embed.FS

	preloadLibs = []string{
		"libc++_shared.so", "libavutil.so",
		"libswresample.so", "libavcodec.so",
	}
)
