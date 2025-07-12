//go:build !android

package android

func GetIntent() Intent {
	// testing
	return Intent{
		Action: ACTION_SEND,
		Type:   "text/plain",
		Text:   "https://nitter.net/14_B4L/status/1944010573924250078",
	}

	return Intent{}
}

func ReadContent(uri string) []byte {
	return []byte{}
}
