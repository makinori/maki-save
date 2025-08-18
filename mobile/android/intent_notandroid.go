//go:build !android

package android

import (
	"os"
)

func GetIntent() Intent {
	_, intentTest := os.LookupEnv("INTENT_TEST")
	if intentTest {
		return Intent{
			Action: ACTION_SEND,
			Type:   "text/plain",
			// Text:   "https://x.com/youmu_i19/status/1944012392939106547", // video
			// Text: "https://x.com/ShitpostRock/status/1943487719578927504", // quote reply
			// Text: "https://x.com/Yukimachis/status/1946112791187988643", // video
			// Text: "https://x.com/abmayo_mfp/status/1949651267304693771", // image (slight nsfw)
			// Text: "https://x.com/birdman46049238/status/1852311426666537322", // gif
			// Text: "https://x.com/withNagi_7/status/1951586795487216038", // multiple images
			Text: "https://x.com/chibikki_ikki/status/1953081885631918090", // only png
		}
		// return Intent{
		// 	Action: ACTION_SEND,
		// 	Type:   "video/mp4",
		// 	URI: []string{
		// 		"/home/maki/Videos/worm.webm",
		// 	},
		// }
	}

	if len(os.Args) <= 1 {
		return Intent{}
	}

	return Intent{
		Action: ACTION_SEND,
		Type:   "image/*", // doesnt matter
		URI:    os.Args[1:],
	}
}

func ReadContent(uri string) []byte {
	data, err := os.ReadFile(uri)
	if err != nil {
		return []byte{}
	}
	return data
}
