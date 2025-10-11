//go:build !wasm

package scrape

import "github.com/makinori/maki-save/ffmpeg"

func getMiddleFrameFromVideo(inputData []byte) ([]byte, error) {
	return ffmpeg.GetMiddleFrameFromVideo(inputData)
}
