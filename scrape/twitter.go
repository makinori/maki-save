package scrape

import (
	"net/url"
	"regexp"
	"slices"

	"github.com/makinori/maki-save/immich"
)

var (
	twitterHosts = []string{
		"twitter.com",
		"x.com",
		"nitter.net",
		"vxtwitter.com",
	}

	twitterPathRegexp = regexp.MustCompile(`(?i)\/(.+?)\/status\/([0-9]+)`)

	/*
		// /pic/media/someid.jpg?name=small&format=webp
		// /pic/amplify_video_thumb/someid/img/someid.jpg?name=small&format=webp
		// removes including and after .jpg
		twitterMediaPathPrefix = regexp.MustCompile(`(.+)\..+$`)
	*/
)

/*
func twitterImageURL(inputURL string) string {
	// medium, large, orig
	// can also use + ":orig"

	imageURL, err := url.Parse(inputURL)
	if err != nil {
		return ""
	}

	prefixMatches := twitterMediaPathPrefix.FindStringSubmatch(imageURL.Path)
	if len(prefixMatches) == 0 {
		return ""
	}

	imageURL.RawQuery = "name=orig&format=jpg"
	imageURL.Path = prefixMatches[1] + ".jpg"

	return imageURL.String()
}
*/

func TestTwitter(url *url.URL) bool {
	return slices.Contains(twitterHosts, url.Host)
}

func Twitter(url *url.URL) ([]immich.File, error) {
	return vxTwitter(url)
}
