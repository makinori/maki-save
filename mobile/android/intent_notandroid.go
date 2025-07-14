//go:build !android

package android

func GetIntent() Intent {
	// testing
	return Intent{
		Action: ACTION_SEND,
		Type:   "text/plain",
		Text:   "https://x.com/youmu_i19/status/1944012392939106547", // video
		// Text: "https://x.com/ShitpostRock/status/1943487719578927504", // quote reply
	}

	return Intent{}
}

func ReadContent(uri string) []byte {
	return []byte{}
}
