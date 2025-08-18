//go:build !android

package ffmpeg

import "io/fs"

var (
	libsFS fs.FS
)
