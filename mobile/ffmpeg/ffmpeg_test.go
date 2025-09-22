package ffmpeg

import (
	"os"
	"testing"
)

func TestGetMiddleFrameFromVideo(t *testing.T) {
	fileData, err := os.ReadFile("/home/maki/Videos/sigma-event-horizon.mp4")
	if err != nil {
		t.Fatal(err)
	}

	frame, err := GetMiddleFrameFromVideo(fileData)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("got frame", len(frame))
}
