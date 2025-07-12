//go:build !android

package android

func GetIntent() Intent {
	// testing
	return Intent{
		Action: ACTION_SEND,
		Type:   "text/plain",
		Text:   "https://x.com/youmu_i19/status/1944012392939106547",
	}

	return Intent{}
}

func ReadContent(uri string) []byte {
	return []byte{}
}
