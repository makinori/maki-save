package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sync"

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
)

/*
func Nitter(url *url.URL) ([]immich.File, error) {
	twitterPathMatches := twitterPathRegexp.FindStringSubmatch(url.Path)
	if len(twitterPathMatches) == 0 {
		return []immich.File{}, errors.New("failed to match username and id from url")
	}

	username := twitterPathMatches[1]
	id := twitterPathMatches[2]

	filenamePrefix := username + "-" + id

	req, err := http.NewRequest("GET", "https://nitter.net"+url.Path, nil)
	if err != nil {
		return []immich.File{}, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "none")
	req.Header.Add("Sec-GPC", "1")

	client := http.Client{}
	res, err := client.Do(req)
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

	imageUrls := []string{}

	doc.Find("meta[property='twitter:image:src']").Each(func(i int, s *goquery.Selection) {
		imageUrl, ok := s.Attr("content")
		if !ok {
			return
		}
		imageUrls = append(imageUrls, imageUrl)
	})

	files := make([]immich.File, len(imageUrls))

	wg := sync.WaitGroup{}
	for i, imageUrl := range imageUrls {
		wg.Add(1)
		go func() {
			defer wg.Done()

			files[i].Name = fmt.Sprintf("%s-%02d%s",
				filenamePrefix, i+1, path.Ext(imageUrl),
			)

			var err error
			res, err := http.Get(imageUrl)
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
		}()
	}
	wg.Wait()

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

	filenamePrefix := vxTwitterRes.Username + "-" + vxTwitterRes.ID

	files := make([]immich.File, len(foundMedia))

	wg := sync.WaitGroup{}
	for i, media := range foundMedia {
		wg.Add(1)
		go func() {
			defer wg.Done()

			files[i].Name = fmt.Sprintf("%s-%02d%s",
				filenamePrefix, i+1, path.Ext(media.URL),
			)

			// medium, large, orig
			// can also use + ":orig"
			var err error
			res, err := http.Get(media.URL + "?name=orig")
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

			if media.Type == VXTwitterMediaTypeVideo ||
				media.Type == VXTwitterMediaTypeGif {

				// dont need to for thumbnail
				res, err := http.Get(media.ThumbnailURL)
				if err != nil {
					files[i].Err = err
					return
				}
				defer res.Body.Close()

				files[i].Thumbnail, err = io.ReadAll(res.Body)
				if err != nil {
					files[i].Err = err
					return
				}
			}
		}()
	}
	wg.Wait()

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
