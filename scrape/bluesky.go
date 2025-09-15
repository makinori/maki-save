package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/makinori/maki-immich/immich"
)

// what the fuck

var (
	blueskyURLRegexp     = regexp.MustCompile(`^https?:\/\/bsky\.app\/profile\/([^/]+)\/post\/([^/?#]+)(?:[?#].*)?$`)
	blueskyPostURIRegexp = regexp.MustCompile(`at:\/\/(did:plc:.+?)\/app\.bsky\.feed\.post`)
)

func TestBluesky(url *url.URL) bool {
	return blueskyURLRegexp.MatchString(url.String())
}

type blueskyRecord struct {
	URI   string `json:"uri"`
	Value struct {
		Embed struct {
			Type   string `json:"$type"`
			Images []struct {
				Image struct {
					Ref struct {
						Link string `json:"$link"`
					} `json:"ref"`
				} `json:"image"`
			} `json:"images"`
		} `json:"embed"`
	} `json:"value"`
}

func Bluesky(postURL *url.URL) ([]immich.File, error) {
	matches := blueskyURLRegexp.FindStringSubmatch(postURL.String())
	if len(matches) == 0 {
		return []immich.File{}, errors.New("no matches found in url")
	}

	handle := matches[1]
	postID := matches[2]

	recordQuery := url.Values{}
	recordQuery.Add("repo", handle)
	recordQuery.Add("collection", "app.bsky.feed.post")
	recordQuery.Add("rkey", postID)

	recordRes, err := http.Get(
		"https://public.api.bsky.app/xrpc/com.atproto.repo.getRecord?" +
			recordQuery.Encode(),
	)
	if err != nil {
		return []immich.File{}, err
	}
	defer recordRes.Body.Close()

	recordData, err := io.ReadAll(recordRes.Body)
	if err != nil {
		return []immich.File{}, err
	}

	// os.WriteFile("test.json", recordData, 0644)

	var record blueskyRecord
	err = json.Unmarshal(recordData, &record)
	if err != nil {
		return []immich.File{}, err
	}

	postURIMatches := blueskyPostURIRegexp.FindStringSubmatch(record.URI)
	didPlc := postURIMatches[1]

	// get files

	fileURLs := make([]string, len(record.Value.Embed.Images))
	thumbnailURLs := make([]string, len(record.Value.Embed.Images))

	for i, image := range record.Value.Embed.Images {
		fileURLs[i] = fmt.Sprintf(
			"https://cdn.bsky.app/img/feed_thumbnail/plain/%s/%s",
			didPlc, image.Image.Ref.Link,
		)
	}

	prefix := fmt.Sprintf("%s-%s-", handle, postID)

	return getFilesFromURLs(prefix, fileURLs, thumbnailURLs), nil
}
