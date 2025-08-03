package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/makinori/maki-immich/immich"
)

// thank you so much https://github.com/dylanpdx/BetterTwitFix
// please please please donate at https://ko-fi.com/dylanpdx

type VXTwitterMediaType string

const (
	VXTwitterMediaTypeVideo VXTwitterMediaType = "video"
	VXTwitterMediaTypeImage VXTwitterMediaType = "image"
	VXTwitterMediaTypeGIF   VXTwitterMediaType = "gif"
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

func vxTwitterImageURL(inputURL string) string {
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

	fileURLs := make([]string, len(foundMedia))
	thumbnailURLs := make([]string, len(foundMedia))

	for i, media := range foundMedia {
		switch media.Type {
		case VXTwitterMediaTypeImage:
			fileURLs[i] = vxTwitterImageURL(media.URL)
		case VXTwitterMediaTypeVideo, VXTwitterMediaTypeGIF:
			fileURLs[i] = media.URL
			thumbnailURLs[i] = vxTwitterImageURL(media.ThumbnailURL)
		}
	}

	files := getFilesFromURLs(
		fmt.Sprintf("%s-%s-", vxTwitterRes.Username, vxTwitterRes.ID),
		fileURLs, thumbnailURLs,
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
