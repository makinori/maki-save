package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/makinori/maki-immich/immich"
)

// thank you so much https://github.com/dylanpdx/BetterTwitFix
// please please please donate at https://ko-fi.com/dylanpdx

type vxTwitterMediaType string

const (
	vxTwitterMediaTypeVideo vxTwitterMediaType = "video"
	vxTwitterMediaTypeImage vxTwitterMediaType = "image"
	vxTwitterMediaTypeGIF   vxTwitterMediaType = "gif"
)

type vxTwitterMedia struct {
	ThumbnailURL string             `json:"thumbnail_url"`
	Type         vxTwitterMediaType `json:"type"`
	URL          string             `json:"url"`
}

type vxTwitterTweet struct {
	Media        []vxTwitterMedia `json:"media_extended"`
	QuoteRetweet *struct {
		Media []vxTwitterMedia `json:"media_extended"`
		ID    string           `json:"tweetID"`
		// DisplayName string `json:"user_name"`
		Username string `json:"user_screen_name"`
	} `json:"qrt"`
	ID string `json:"tweetID"`
	// DisplayName string `json:"user_name"`
	Username string `json:"user_screen_name"`
}

func vxTwitterImageURL(inputURL string) string {
	imageURL, err := url.Parse(inputURL)
	if err != nil {
		return ""
	}

	// can use: medium, large, orig
	imageURL.RawQuery = "name=orig"

	// some images are only png, so keep extension
	// however if webp, turn into jpg
	if strings.HasSuffix(imageURL.Path, ".webp") {
		imageURL.Path = imageURL.Path[:len(imageURL.Path)-5] + ".jpg"
	}

	return imageURL.String()
}

func vxTwitterProcessMedia(
	tweetID string, username string, tweetMedia []vxTwitterMedia,
	outputFiles []immich.File,
) {
	fileURLs := make([]string, len(tweetMedia))
	thumbnailURLs := make([]string, len(tweetMedia))

	for i, media := range tweetMedia {
		switch media.Type {
		case vxTwitterMediaTypeImage:
			fileURLs[i] = vxTwitterImageURL(media.URL)
		case vxTwitterMediaTypeVideo, vxTwitterMediaTypeGIF:
			fileURLs[i] = media.URL
			thumbnailURLs[i] = vxTwitterImageURL(media.ThumbnailURL)
		}
	}

	copy(outputFiles, getFilesFromURLs(
		fmt.Sprintf("%s-%s-", username, tweetID),
		fileURLs, thumbnailURLs,
	))
}

func vxTwitter(url *url.URL) ([]immich.File, error) {
	twitterPathMatches := twitterPathRegexp.FindStringSubmatch(url.Path)
	if len(twitterPathMatches) == 0 {
		return []immich.File{}, errors.New("failed to match username and id from url")
	}

	httpRes, err := http.Get("https://api.vxtwitter.com" + url.Path)
	if err != nil {
		return []immich.File{}, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != 200 {
		return []immich.File{}, fmt.Errorf("%d: %s", httpRes.StatusCode, httpRes.Status)
	}

	jsonData, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return []immich.File{}, err
	}

	tweet := vxTwitterTweet{}
	err = json.Unmarshal(jsonData, &tweet)
	if err != nil {
		return []immich.File{}, err
	}

	// i dont think you can simulatenously post images and quote retweet

	filesFound := len(tweet.Media)
	if tweet.QuoteRetweet != nil {
		filesFound += len(tweet.QuoteRetweet.Media)
	}

	if filesFound == 0 {
		return []immich.File{}, nil
	}

	var files = make([]immich.File, filesFound)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		vxTwitterProcessMedia(
			tweet.ID, tweet.Username, tweet.Media,
			files[0:len(tweet.Media)],
		)
	}()

	if tweet.QuoteRetweet != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vxTwitterProcessMedia(
				tweet.QuoteRetweet.ID, tweet.QuoteRetweet.Username,
				tweet.QuoteRetweet.Media, files[len(tweet.Media):filesFound],
			)
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
