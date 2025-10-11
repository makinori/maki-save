//go:build wasm

package scrape

import (
	"errors"
)

func getMiddleFrameFromVideo(inputData []byte) ([]byte, error) {
	return []byte{}, errors.New("not supported with wasm")
}
