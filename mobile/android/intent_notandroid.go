//go:build !android

package android

import "os"

func GetIntent() Intent {
	// testing
	return Intent{
		Action: ACTION_SEND,
		Type:   "text/plain",
		// Text:   "https://x.com/youmu_i19/status/1944012392939106547", // video
		// Text: "https://x.com/ShitpostRock/status/1943487719578927504", // quote reply
		// Text: "https://x.com/Yukimachis/status/1946112791187988643",
		Text: "https://x.com/abmayo_mfp/status/1949651267304693771",
	}
	// return Intent{
	// 	Action: ACTION_SEND,
	// 	Type:   "video/mp4",
	// 	URI: []string{
	// 		"/home/maki/Videos/worm.webm",
	// 	},
	// }
	return Intent{}
}

func ReadContent(uri string) []byte {
	data, err := os.ReadFile(uri)
	if err != nil {
		return []byte{}
	}
	return data
}
