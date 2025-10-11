//go:build !windows

package ffmpeg

import "github.com/ebitengine/purego"

func libOpen(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_LAZY)
}

func libClose(handle uintptr) {
	purego.Dlclose(handle)
}
