package scrape

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/makinori/maki-immich/immich"
)

var (
	TwitterHosts = []string{
		"twitter.com",
		"x.com",
		"nitter.net",
		"vxtwitter.com",
	}

	twitterPathRegexp = regexp.MustCompile(`(?i)\/(.+?)\/status\/([0-9]+)`)

	// /pic/media/someid.jpg?name=small&format=webp
	// /pic/amplify_video_thumb/someid/img/someid.jpg?name=small&format=webp
	// removes including and after .jpg
	twitterMediaPathPrefix = regexp.MustCompile(`(.+)\..+$`)

	cleanUpExt = regexp.MustCompile(`(?i)[^\.a-z0-9].*$`)

	//go:embed nitter.txt
	nitterURLString string
)

func getTwitterImageURLs(
	username, id string, imageURLs []string,
) []immich.File {
	files := make([]immich.File, len(imageURLs))

	wg := sync.WaitGroup{}
	for i, imageURL := range imageURLs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ext := path.Ext(imageURL)
			ext = cleanUpExt.ReplaceAllString(ext, "")

			files[i].Name = fmt.Sprintf("%s-%s-%02d%s",
				username, id, i+1, ext,
			)

			var err error
			res, err := http.Get(imageURL)
			if err != nil {
				files[i].Err = err
				return
			}
			defer res.Body.Close()

			files[i].Data, err = io.ReadAll(res.Body)
			if err != nil {
				files[i].Err = err
				return
			}

			// if media.Type == VXTwitterMediaTypeVideo ||
			// 	media.Type == VXTwitterMediaTypeGif {

			// 	// dont need to for thumbnail
			// 	res, err := http.Get(media.ThumbnailURL)
			// 	if err != nil {
			// 		files[i].Err = err
			// 		return
			// 	}
			// 	defer res.Body.Close()

			// 	files[i].Thumbnail, err = io.ReadAll(res.Body)
			// 	if err != nil {
			// 		files[i].Err = err
			// 		return
			// 	}
			// }
		}()
	}
	wg.Wait()

	return files
}

func nitterFixImageURL(nitterURL url.URL, escapedPath string) (string, error) {
	path, err := url.PathUnescape(escapedPath)
	if err != nil {
		return "", err
	}

	prefixMatches := twitterMediaPathPrefix.FindStringSubmatch(path)
	if len(prefixMatches) == 0 {
		return "", errors.New("failed to match media prefix")
	}

	nitterURL.Path = prefixMatches[1] + ".jpg?name=orig&format=jpg"

	return nitterURL.String(), nil
}

func Nitter(inputURL *url.URL) ([]immich.File, error) {
	twitterPathMatches := twitterPathRegexp.FindStringSubmatch(inputURL.Path)
	if len(twitterPathMatches) == 0 {
		return []immich.File{}, errors.New("failed to match username and id from url")
	}

	username := twitterPathMatches[1]
	id := twitterPathMatches[2]

	nitterURL, err := url.Parse(nitterURLString)
	if err != nil {
		return []immich.File{}, err
	}

	nitterURL.Path = inputURL.Path

	req, err := http.NewRequest("GET", nitterURL.String(), nil)
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

	doc.Find(".main-tweet .attachment img").Each(func(i int, s *goquery.Selection) {
		imagePath, ok := s.Attr("src")
		if !ok {
			return
		}

		imageURL, err := nitterFixImageURL(*nitterURL, imagePath)
		if err != nil {
			return
		}

		imageURLs = append(imageURLs, imageURL)
	})

	doc.Find(".main-tweet .attachment video").Each(func(i int, s *goquery.Selection) {
		thumbnailPath, ok := s.Attr("poster")
		if !ok {
			return
		}

		thumbnailURL, err := nitterFixImageURL(*nitterURL, thumbnailPath)
		if err != nil {
			return
		}

		// TODO: get video url

		fmt.Println(thumbnailURL)
	})

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

/*
type VXTwitterMediaType string

const (
	VXTwitterMediaTypeVideo VXTwitterMediaType = "video"
	VXTwitterMediaTypeImage VXTwitterMediaType = "image"
	VXTwitterMediaTypeGif   VXTwitterMediaType = "gif"
)

type VXTwitterMedia struct {
	ThumbnailURL string             `json:"thumbnail_url"`
	Type         VXTwitterMediaType `json:"type"`
	URL          string             `json:"url"`
}

type VXTwitterRes struct {
	Media        []VXTwitterMedia `json:"media_extended"`
	QuoteRetweet struct {
		Media []VXTwitterMedia `json:"media_extended"`
	} `json:"qrt"`
	ID          string `json:"tweetID"`
	DisplayName string `json:"user_name"`
	Username    string `json:"user_screen_name"`
}

func VXTwitter(url *url.URL) ([]immich.File, error) {
	twitterPathMatches := twitterPathRegexp.FindStringSubmatch(url.Path)
	if len(twitterPathMatches) == 0 {
		return []immich.File{}, errors.New("failed to match username and id from url")
	}

	res, err := http.Get("https://api.vxtwitter.com" + url.Path)
	if err != nil {
		return []immich.File{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return []immich.File{}, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return []immich.File{}, err
	}

	vxTwitterRes := VXTwitterRes{}
	err = json.Unmarshal(data, &vxTwitterRes)
	if err != nil {
		return []immich.File{}, err
	}

	foundMedia := append(
		vxTwitterRes.Media,
		vxTwitterRes.QuoteRetweet.Media...,
	)

	imageURLs := make([]string, len(foundMedia))
	for i, media := range foundMedia {
		// medium, large, orig
		// can also use + ":orig"
		imageURLs[i] = media.URL + "?name=orig"
	}

	files := getTwitterImageURLs(
		vxTwitterRes.Username, vxTwitterRes.ID, imageURLs,
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
*/
