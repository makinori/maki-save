package scrape

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/makinori/maki-save/immich"
)

var (
	instagramHosts = []string{
		"instagram.com",
		"ddinstagram.com",
		"uuinstagram.com",
		"eeinstagram.com",
	}

	instagramIDRegexp = regexp.MustCompile(`\/(?:p|reels)\/(.+?)(?:\/|$)`)

	noRedirClient = *http.DefaultClient
)

func init() {
	noRedirClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func TestInstagram(url *url.URL) bool {
	return slices.Contains(
		instagramHosts, strings.TrimPrefix(url.Host, "www."),
	)
}

func getMetaContent(
	doc *goquery.Document, name string, resolveUrl *url.URL,
) (string, bool) {
	selection := doc.Find(fmt.Sprintf(`meta[name="%s"]`, name))
	if selection == nil {
		return "", false
	}

	value, ok := selection.Attr("content")
	if !ok {
		return "", false
	}

	if resolveUrl == nil {
		return value, true
	}

	if strings.HasPrefix(value, "http") {
		return value, true
	}

	newUrl := *resolveUrl
	newUrl.Path = value
	return newUrl.String(), true
}

func getInstaFixImageURL(imageURL string) (string, error) {
	res, err := noRedirClient.Get(imageURL)
	if err != nil {
		return "", err
	}

	redirectURL := res.Header.Get("Location")
	if redirectURL == "" {
		return "", nil
	}

	return redirectURL, nil
}

func Instagram(scrapeURL *url.URL) ([]immich.File, error) {
	idMatches := instagramIDRegexp.FindStringSubmatch(scrapeURL.Path)
	if len(idMatches) == 0 {
		return []immich.File{}, errors.New("failed to find id in url")
	}

	id := idMatches[1]

	newUrl := *scrapeURL
	newUrl.Host = "www.uuinstagram.com"

	res, err := noRedirClient.Get(newUrl.String())
	if err != nil {
		return []immich.File{}, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []immich.File{}, err
	}

	// html, _ := doc.Html()
	// os.WriteFile("test.html", []byte(html), 0644)

	username, ok := getMetaContent(doc, "twitter:title", nil)
	if !ok {
		return []immich.File{}, errors.New("failed to get twitter:title")
	}

	prefix := fmt.Sprintf(
		"%s-%s-", strings.TrimPrefix(username, "@"), id,
	)

	twitterCard, ok := getMetaContent(doc, "twitter:card", nil)
	if !ok {
		return []immich.File{}, errors.New("failed to get twitter:card")
	}

	if twitterCard == "player" {
		// video

		twitterPlayerStream, ok := getMetaContent(doc, "twitter:player:stream", &newUrl)
		if !ok {
			return []immich.File{}, errors.New("failed to get twitter:player:stream")
		}

		files := getFilesFromURLs(
			prefix, []string{twitterPlayerStream}, []string{""},
		)

		files[0].UIIsVideo = true
		files[0].UIThumbnail, _ = getMiddleFrameFromVideo(files[0].Data)

		return files, nil
	}

	// pictures

	twitterImage, ok := getMetaContent(doc, "twitter:image", &newUrl)
	if !ok {
		return []immich.File{}, errors.New("failed to get twitter:image")
	}

	var imageURLs []string

	if strings.Contains(twitterImage, "/images/") {
		// only one image
		imageURLs = append(imageURLs, twitterImage)
	} else {
		// multiple. no way to easily find out how many unfortunately

		imageURLPrefix := strings.Replace(twitterImage, "/grid", "/images", 1)
		if !strings.HasSuffix(imageURLPrefix, "/") {
			imageURLPrefix += "/"
		}

		for i := 1; ; i++ {
			imageURL, err := getInstaFixImageURL(
				imageURLPrefix + strconv.Itoa(i),
			)
			if err != nil {
				return []immich.File{}, err
			}
			if imageURL == "" {
				break
			}
			imageURLs = append(imageURLs, imageURL)
		}
	}

	files := getFilesFromURLs(
		prefix, imageURLs, make([]string, len(imageURLs)),
	)

	return files, nil
}
