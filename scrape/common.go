package scrape

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sync"
	"unsafe"

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

			ext := path.Ext(imageURL)
			ext = cleanUpExt.ReplaceAllString(ext, "")

			if ext == "" {
				contentType := http.DetectContentType(files[i].Data)
				exts, _ := mime.ExtensionsByType(contentType)
				if len(exts) > 0 {
					ext = exts[len(exts)-1]
				}
			}

			files[i].Name = fmt.Sprintf("%s%02d%s", prefix, i+1, ext)

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
	var extraData unsafe.Pointer
	switch {
	case TestTwitter(scrapeURL):
		return "Twitter", Twitter
	case TestPoshmark(scrapeURL):
		return "Poshmark", Poshmark
	case TestBluesky(scrapeURL):
		return "Bluesky", Bluesky
	// case TestActivityPub(scrapeURL, &extraData):
	// 	return "ActivityPub", func(url *url.URL) ([]immich.File, error) {
	// 		return ActivityPub(url, &extraData)
	// 	}
	// authenticated mastodon can search and resolve all links we'll encounter
	// not all activitypub servers allow unauthenticated requests
	case TestMastodonFediverse(scrapeURL, &extraData):
		return "Mastodon Fediverse", func(url *url.URL) ([]immich.File, error) {
			return MastodonFediverse(url, &extraData)
		}
	case TestGeneric(scrapeURL, &extraData):
		return "Generic", func(url *url.URL) ([]immich.File, error) {
			return Generic(url, &extraData)
		}
	}
	return "", func(url *url.URL) ([]immich.File, error) {
		return []immich.File{}, errors.New("no scrape function")
	}
}
