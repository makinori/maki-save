//go:build windows

package ffmpeg

import (
	"syscall"
)

func libOpen(path string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(path)
	return uintptr(handle), err
}

func libClose(handle uintptr) {
	syscall.FreeLibrary(syscall.Handle(handle))
}
