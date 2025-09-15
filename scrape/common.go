package scrape

import (
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
	cleanUpExt = regexp.MustCompile(`(?i)[^\.a-z0-9].*$`)
)

type TestFn func(url *url.URL) bool

type ScrapeFn func(url *url.URL) ([]immich.File, error)

func getFilesFromURLs(
	prefix string, fileURLs []string, thumbnailURLs []string,
) []immich.File {
	files := make([]immich.File, len(fileURLs))

	wg := sync.WaitGroup{}
	for i, imageURL := range fileURLs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ext := path.Ext(imageURL)
			ext = cleanUpExt.ReplaceAllString(ext, "")

			files[i].Name = fmt.Sprintf("%s%02d%s", prefix, i+1, ext)

			{
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
			}

			thumbnailURL := thumbnailURLs[i]
			if thumbnailURL == "" {
				return
			}

			{
				res, err := http.Get(thumbnailURL)
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

	return files
}

func Test(scrapeURL *url.URL) (string, ScrapeFn) {
	switch {
	case TestTwitter(scrapeURL):
		return "Twitter", Twitter
	case TestPoshmark(scrapeURL):
		return "Poshmark", Poshmark
	case TestMastodon(scrapeURL):
		return "Mastodon", Mastodon
	}
	return "", func(url *url.URL) ([]immich.File, error) {
		return []immich.File{}, nil
	}
}
