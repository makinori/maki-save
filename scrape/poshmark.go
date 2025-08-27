package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/makinori/maki-immich/immich"
	"mvdan.cc/xurls/v2"
)

var PoshmarkHosts = []string{
	"posh.mk",
	"poshmark.com",
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

type poshmarkInitialState struct {
	ListingDetails struct {
		ListingDetails struct {
			ID        string `json:"id"`
			CoverShot struct {
				URLLarge string `json:"url_large"`
				// URLLargeWebP string `json:"url_large_webp"`
			} `json:"cover_shot"`
			Pictures []struct {
				URLLarge string `json:"url_large"`
				// URLLargeWebP string `json:"url_large_webp"`
			} `json:"pictures"`
		} `json:"listingDetails"`
		ListerData struct {
			// ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"listerData"`
	} `json:"$_listing_details"`
}

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

	if htmlRes.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", htmlRes.StatusCode, htmlRes.Status)
	}

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

	fileURLs := make([]string, 1+len(listingDetails.Pictures))
	thumbnailURLs := make([]string, 1+len(listingDetails.Pictures))

	fileURLs[0] = listingDetails.CoverShot.URLLarge
	for i, picture := range listingDetails.Pictures {
		fileURLs[1+i] = picture.URLLarge
	}

	return getFilesFromURLs(
		fmt.Sprintf("%s-%s-", listerData.Username, listingDetails.ID),
		fileURLs, thumbnailURLs,
	), nil
}
