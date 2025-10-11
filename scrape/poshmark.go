package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/makinori/maki-save/immich"
	"mvdan.cc/xurls/v2"
)

var poshmarkHosts = []string{
	"posh.mk",
	"poshmark.com",
}

func TestPoshmark(url *url.URL) bool {
	return slices.Contains(poshmarkHosts, url.Host)
}

func poshmarkHandleRedirect(url *url.URL) (*url.URL, error) {
	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}

	html, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	foundURLs := xurls.Relaxed().FindAllString(string(html), -1)

	for _, foundURL := range foundURLs {
		if strings.Contains(foundURL, "//poshmark.com/listing/") {
			return url.Parse(foundURL)
		}
	}

	return nil, errors.New("listing url not found")
}

var poshmarkInitialStateJSONRegexp = regexp.MustCompile(`window.__INITIAL_STATE__=({.*});`)

type poshmarkPicture struct {
	URLLarge  string `json:"url_large"`
	URLMedium string `json:"url"`
	URLSmall  string `json:"url_small"`
	// URLLargeWebP string `json: "url_large_webp"`
}

func (picture *poshmarkPicture) getLargest() string {
	if picture.URLLarge != "" {
		return picture.URLLarge
	}
	if picture.URLMedium != "" {
		return picture.URLMedium
	}
	return picture.URLSmall
}

type poshmarkInitialState struct {
	ListingDetails struct {
		ListingDetails struct {
			ID          string            `json:"id"`
			CoverShot   poshmarkPicture   `json:"cover_shot"`
			Description string            `json:"description"`
			Pictures    []poshmarkPicture `json:"pictures"`
			Videos      []struct {
				Media struct {
					VideoMediaContent map[string]string `json:"video_media_content"`
					ThumbnailContent  poshmarkPicture   `json:"thumbnail_content"`
				} `json:"media"`
			} `json:"Videos"`
			Title   string `json:"title"`
			Size    string `json:"size"`
			SizeObj struct {
				DisplayWithSizeSystem string `json:"display_with_size_system"`
				// b string `json:"display_with_system_and_set"`
			} `json:"size_obj"`
		} `json:"listingDetails"`
		ListerData struct {
			// ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"listerData"`
	} `json:"$_listing_details"`
}

var allButNumbersRegexp = regexp.MustCompile(`[^0-9]`)

func Poshmark(url *url.URL) ([]immich.File, error) {
	if url.Host == "posh.mk" {
		var err error
		url, err = poshmarkHandleRedirect(url)
		if err != nil {
			return []immich.File{}, errors.New(
				"failed to handle redirect: " + err.Error(),
			)
		}
	}

	htmlRes, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer htmlRes.Body.Close()

	// sold items are 404
	// if htmlRes.StatusCode != 200 {
	// 	return nil, fmt.Errorf("%d: %s", htmlRes.StatusCode, htmlRes.Status)
	// }

	html, err := io.ReadAll(htmlRes.Body)
	if err != nil {
		return nil, err
	}

	initialStateMatches := poshmarkInitialStateJSONRegexp.FindSubmatch(html)
	if len(initialStateMatches) == 0 {
		return nil, errors.New("failed to find initial state json")
	}

	// os.WriteFile("test.json", []byte(initialStateMatches[1]), 0644)

	var initialState poshmarkInitialState
	err = json.Unmarshal(initialStateMatches[1], &initialState)
	if err != nil {
		return nil, err
	}

	// test, _ := json.MarshalIndent(initialState, "", "\t")
	// fmt.Println(string(test))

	listerData := &initialState.ListingDetails.ListerData
	listingDetails := &initialState.ListingDetails.ListingDetails

	mediaLength := 1 + len(listingDetails.Videos) + len(listingDetails.Pictures)
	fileURLs := make([]string, mediaLength)
	thumbnailURLs := make([]string, mediaLength)
	isVideo := make([]bool, mediaLength)

	i := 0

	fileURLs[i] = listingDetails.CoverShot.getLargest()
	i++

	for _, video := range listingDetails.Videos {
		biggestQuality := 0
		biggestQualityKey := ""

		qualityKeys := maps.Keys(video.Media.VideoMediaContent)
		for qualityKey := range qualityKeys {
			quality, _ := strconv.Atoi(
				allButNumbersRegexp.ReplaceAllString(qualityKey, ""),
			)
			if quality > biggestQuality {
				biggestQuality = quality
				biggestQualityKey = qualityKey
			}
		}

		fileURLs[i] = video.Media.VideoMediaContent[biggestQualityKey]
		thumbnailURLs[i] = video.Media.ThumbnailContent.getLargest()
		isVideo[i] = true

		i++
	}

	for _, picture := range listingDetails.Pictures {
		fileURLs[i] = picture.getLargest()
		i++
	}

	description := fmt.Sprintf(
		"%s\nSize %s\n%s",
		listingDetails.Title,
		listingDetails.SizeObj.DisplayWithSizeSystem,
		listingDetails.Description,
	)

	// fmt.Println(description)

	files := getFilesFromURLs(
		fmt.Sprintf("%s-%s-", listerData.Username, listingDetails.ID),
		fileURLs, thumbnailURLs,
	)

	for i := range files {
		files[i].Description = description
		if isVideo[i] {
			files[i].UIIsVideo = true
		}
	}

	return files, nil
}
