package scrape

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/makinori/maki-save/immich"
)

var (
	//go:embed nitter.txt
	nitterURLString string
	nitterURL       *url.URL
)

func init() {
	var err error
	nitterURL, err = url.Parse(nitterURLString)
	if err != nil {
		panic(err)
	}
}

func nitterPathToURL(escapedPath string) (string, error) {
	unescapedPath, err := url.PathUnescape(escapedPath)
	if err != nil {
		return "", err
	}

	resultURL := *nitterURL
	resultURL.Path = unescapedPath

	return resultURL.String(), nil
}

func nitterRawPathToURL(rawPath string) string {
	resultURL := *nitterURL
	resultURL.Path = ""
	return resultURL.String() + rawPath
}

func nitterImageURL(escapedPath string) (string, error) {
	path, err := url.PathUnescape(escapedPath)
	if err != nil {
		return "", err
	}

	prefixMatches := twitterMediaPathPrefix.FindStringSubmatch(path)
	if len(prefixMatches) == 0 {
		return "", errors.New("failed to match media prefix")
	}

	url := *nitterURL
	url.Path = prefixMatches[1] + ".jpg?name=orig&format=jpg"

	return url.String(), nil

}

func nitter(inputURL *url.URL) ([]immich.File, error) {
	twitterPathMatches := twitterPathRegexp.FindStringSubmatch(inputURL.Path)
	if len(twitterPathMatches) == 0 {
		return []immich.File{}, errors.New("failed to match username and id from url")
	}

	username := twitterPathMatches[1]
	id := twitterPathMatches[2]

	nitterTweetURL, err := nitterPathToURL(inputURL.Path)
	if err != nil {
		return []immich.File{}, err
	}

	req, err := http.NewRequest("GET", nitterTweetURL, nil)
	if err != nil {
		return []immich.File{}, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "none")
	req.Header.Add("Sec-GPC", "1")

	req.AddCookie(&http.Cookie{
		Name:  "hlsPlayback",
		Value: "on",
	})

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []immich.File{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return []immich.File{}, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []immich.File{}, err
	}

	usernameEl := doc.Find(".username").First()
	if usernameEl == nil {
		return []immich.File{}, errors.New("failed to find username in response")
	}

	imageURLs := []string{}

	// TODO: improve error handling

	doc.Find(".main-tweet .attachment img").Each(
		func(i int, imgEl *goquery.Selection) {
			imagePath, ok := imgEl.Attr("src")
			if !ok {
				return
			}

			imageURL, err := nitterImageURL(imagePath)
			if err != nil {
				return
			}

			imageURLs = append(imageURLs, imageURL)
		},
	)

	doc.Find(".main-tweet .attachment video").Each(
		func(i int, videoEl *goquery.Selection) {
			thumbnailPath, ok := videoEl.Attr("poster")
			if !ok {
				return
			}

			thumbnailURL, err := nitterImageURL(thumbnailPath)
			if err != nil {
				return
			}

			fmt.Println("thumbnail", thumbnailURL)

			mp4Source := videoEl.Find("source[type='video/mp4']")
			if mp4Source != nil {
				gifPath, ok := mp4Source.Attr("src")

				if ok {
					gifURL, err := nitterPathToURL(gifPath)
					if err != nil {
						return
					}
					fmt.Println("gif", gifURL)
				}
			}

			m3u8Source, ok := videoEl.Attr("data-url")
			if ok {
				fmt.Println(m3u8Source)
				m3u8Url := nitterRawPathToURL(m3u8Source)
				fmt.Println(m3u8Url)

			}
		},
	)

	files := getTwitterImageURLs(
		username, id, imageURLs,
	)

	errText := ""
	for _, file := range files {
		if file.Err != nil {
			errText += file.Err.Error() + "\n"
		}
	}

	if errText != "" {
		return []immich.File{}, errors.New(errText)
	}

	return files, nil
}
