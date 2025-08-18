//go:build android

package ffmpeg

import (
	"embed"
	"io/fs"
)

var (
	//go:embed android-arm64-v8a
	libsFSAndroid embed.FS
	libsFS        fs.FS
)

func init() {
	var err error
	libsFS, err = fs.Sub(libsFSAndroid, "android-arm64-v8a")
	if err != nil {
		panic(err)
	}
}
