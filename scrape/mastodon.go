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

var (
	mastodonUrlRegexp = regexp.MustCompile(`^https?:\/\/[^/]+\/@([^/]+)\/([0-9]{15,19})`)
)

func TestMastodon(url *url.URL) bool {
	return mastodonUrlRegexp.MatchString(url.String())
}

type mastodonInstance struct {
	URI string `json:"uri"`
}

type mastodonStatus struct {
	ID      string `json:"id"`
	Account struct {
		Username string `json:"username"`
	} `json:"account"`
	MediaAttachments []struct {
		Type       string `json:"type"`
		URL        string `json:"url"`
		PreviewURL string `json:"preview_url"`
	} `json:"media_attachments"`
}

func Mastodon(url *url.URL) ([]immich.File, error) {
	matches := mastodonUrlRegexp.FindStringSubmatch(url.String())
	if len(matches) == 0 {
		return []immich.File{}, errors.New("no matches found in url")
	}

	// get instance

	instanceRes, err := http.Get(fmt.Sprintf("https://%s/api/v1/instance", url.Host))
	if err != nil {
		return []immich.File{}, err
	}
	defer instanceRes.Body.Close()

	instanceData, err := io.ReadAll(instanceRes.Body)
	if err != nil {
		return []immich.File{}, err
	}

	var instance mastodonInstance
	err = json.Unmarshal(instanceData, &instance)
	if err != nil {
		return []immich.File{}, err
	}

	// get status

	statusRes, err := http.Get(fmt.Sprintf("https://%s/api/v1/statuses/%s", url.Host, matches[2]))
	if err != nil {
		return []immich.File{}, err
	}
	defer statusRes.Body.Close()

	statusData, err := io.ReadAll(statusRes.Body)
	if err != nil {
		return []immich.File{}, err
	}

	// os.WriteFile("test.json", statusData, 0644)

	var status mastodonStatus
	err = json.Unmarshal(statusData, &status)
	if err != nil {
		return []immich.File{}, err
	}

	// get files

	fileURLs := make([]string, len(status.MediaAttachments))
	thumbnailURLs := make([]string, len(status.MediaAttachments))

	for i, media := range status.MediaAttachments {
		fileURLs[i] = media.URL
		// https://docs.joinmastodon.org/entities/MediaAttachment/#type
		switch media.Type {
		case "video", "gifv":
			thumbnailURLs[i] = media.PreviewURL
		}
	}

	prefix := fmt.Sprintf(
		"@%s@%s-%s-", status.Account.Username, instance.URI, status.ID,
	)

	return getFilesFromURLs(prefix, fileURLs, thumbnailURLs), nil
}
