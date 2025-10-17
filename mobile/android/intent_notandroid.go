//go:build !android

package android

import (
	"os"
	"path/filepath"
)

func GetIntent() Intent {
	_, intentTest := os.LookupEnv("INTENT_TEST")
	if intentTest {
		// return Intent{
		// 	Action: ACTION_SEND,
		// 	Type:   "video/mp4",
		// 	URI: []string{
		// 		// "/home/maki/Videos/worm.webm",
		// 		"/home/maki/Videos/sigma-event-horizon.mp4",
		// 	},
		// }

		// below uses scrape package

		// return Intent{
		// 	Action: ACTION_SEND,
		// 	Type:   "text/plain",
		// 	// Text:   "https://x.com/youmu_i19/status/1944012392939106547", // video
		// 	// Text: "https://x.com/ShitpostRock/status/1943487719578927504", // quote reply
		// 	// Text: "https://x.com/Yukimachis/status/1946112791187988643", // video
		// 	// Text: "https://x.com/abmayo_mfp/status/1949651267304693771", // image (slight nsfw)
		// 	// Text: "https://x.com/birdman46049238/status/1852311426666537322", // gif
		// 	// Text: "https://x.com/withNagi_7/status/1951586795487216038", // multiple images
		// 	Text: "https://x.com/chibikki_ikki/status/1953081885631918090", // only png
		// }

		// return Intent{
		// 	Action: ACTION_SENDTO,
		// 	// Text:   "hi there,\nthis is a test url https://posh.mk/rLZ4pGHLaWb\nbye now",
		// 	// Text: "https://posh.mk/l9ZJ5hR7bWb", // has video
		// 	// Text: "https://posh.mk/aOxv4li5sWb", // sold item
		// 	Text: "https://posh.mk/lXoDoOrDsWb", // large images missing
		// }

		// return Intent{
		// 	Action: ACTION_SENDTO,
		// 	Text:   "https://bsky.app/profile/chibikki.bsky.social/post/3lyt3tr353c2p",
		// }

		// return Intent{
		// 	Action: ACTION_SENDTO,
		// 	// Text:   "https://mastodon.hotmilk.space/@maki/113338487636859072",
		// 	// Text: "https://mastodon.hotmilk.space/@maki/113172074826912415", // video
		// 	// Text: "https://mastodon.hotmilk.space/@maki/114674021457567710", // gif
		// 	// not mastodon
		// 	// Text: "https://misskey.design/notes/acoult3ez23m0skq",
		// 	// Text: "https://labyrinth.zone/objects/ab982cfd-c27a-499d-9664-7555ea1839ed",
		// 	Text: "https://mk.absturztau.be/notes/acsmqmgpkequ010q",
		// 	// Text: "https://social.edist.ro/@CosmickTrigger/115210316299208892",
		// }

		return Intent{
			Action: ACTION_SENDTO,
			// Text:   "https://www.instagram.com/reels/DOY2Bz0jH9n/", // reel
			// Text: "https://www.instagram.com/p/CMRVWdeAClC/", // single image
			// Text: "https://www.instagram.com/p/CFugc67oPNu/", // video (doesnt work with uuinstagram)
			Text: "https://www.instagram.com/p/CD3LDKvAujc/", // multiple images
		}

		// return Intent{
		// 	Action: ACTION_SENDTO,
		// 	// Text:   "https://hotmilk.space/u/bQ2LEe.jpg",
		// 	Text: "https://hotmilk.space/u/dDYfyS.webm",
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

func ReadContent(uri string) ([]byte, string) {
	data, err := os.ReadFile(uri)
	if err != nil {
		return []byte{}, ""
	}
	return data, filepath.Base(uri) // on desktop use filepath
}
