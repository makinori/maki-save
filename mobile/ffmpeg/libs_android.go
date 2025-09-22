//go:build android

package ffmpeg

import (
	"embed"
	"io/fs"
)

var (
	useEmbeddedLibraries   = true
	hasSelfContainedTmpDir = true

	//go:embed android-arm64-v8a
	embedLibsFS embed.FS
	libsFS      fs.FS

	preloadLibs = []string{
		"libc++_shared.so", "libavutil.so",
		"libswresample.so", "libavcodec.so",
	}
)

func init() {
	libsFS, _ = fs.Sub(embedLibsFS, "android-arm64-v8a")
}
