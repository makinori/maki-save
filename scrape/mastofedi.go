package scrape

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unsafe"

	"github.com/makinori/maki-save/immich"
)

var (
	//go:embed mastofedi.txt
	MASTODON_TXT          string
	MASTODON_INSTANCE_URL *url.URL
	MASTODON_ACCESS_TOKEN string
)

func init() {
	mastodonTxtLines := strings.Split(MASTODON_TXT, "\n")
	if len(mastodonTxtLines) < 2 {
		panic("mastodon.txt needs 2 lines")
	}

	var err error
	MASTODON_INSTANCE_URL, err = url.Parse(
		strings.TrimSpace(mastodonTxtLines[0]),
	)
	if err != nil {
		panic(err)
	}

	MASTODON_ACCESS_TOKEN = strings.TrimSpace(mastodonTxtLines[1])
}

// https://docs.joinmastodon.org/entities/Status/
type mastodonStatus struct {
	ID      string `json:"id"`
	URI     string `json:"uri"` // more activitypub?
	Account struct {
		Username string `json:"username"`
	} `json:"account"`
	MediaAttachments []struct {
		Type        string `json:"type"`
		URL         string `json:"url"`
		PreviewURL  string `json:"preview_url"`
		Description string `json:"description"`
	} `json:"media_attachments"`
}

type mastodonSearch struct {
	Statuses []mastodonStatus `json:"statuses"`
}

func TestMastodonFediverse(tootURL *url.URL, extraData *unsafe.Pointer) bool {
	params := url.Values{}
	params.Add("resolve", "true")
	params.Add("limit", "1")
	params.Add("type", "statuses")
	params.Add("q", tootURL.String())

	requestURL := *MASTODON_INSTANCE_URL // copy
	requestURL.Path = "/api/v2/search"
	requestURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return false
	}

	req.Header.Add("Authorization", "Bearer "+MASTODON_ACCESS_TOKEN)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return false
	}

	// os.WriteFile("test.json", data, 0644)

	var search mastodonSearch
	err = json.Unmarshal(data, &search)
	if err != nil {
		return false
	}

	if len(search.Statuses) == 0 {
		return false
	}

	*extraData = unsafe.Pointer(&search.Statuses[0])

	return true
}

func MastodonFediverse(
	extraData unsafe.Pointer,
) ([]immich.File, error) {
	if extraData == nil {
		return []immich.File{}, errors.New("extra data is nil")
	}

	status := (*mastodonStatus)(extraData)

	// id will be different on federated servers from ours
	// guess by extracting supposed activitypub note id from uri
	// TODO: is there a better way to do this cause this is just guessing

	noteIDMatches := activityPubNoteIDRegexp.FindStringSubmatch(status.URI)
	if len(noteIDMatches) == 0 {
		return []immich.File{}, errors.New("failed to find note id")
	}

	noteID := noteIDMatches[1]

	// resolve handle with webfinger
	// e.g. mastodon.hotmilk.space is actually just hotmilk.space

	noteURI, err := url.Parse(status.URI)
	if err != nil {
		return []immich.File{}, err
	}

	handle, err := activityPubResolveHandle(
		status.Account.Username, noteURI.Host,
	)
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
		"%s-%s-", handle, noteID,
	)

	files := getFilesFromURLs(prefix, fileURLs, thumbnailURLs)

	for i, media := range status.MediaAttachments {
		files[i].Description = media.Description // why not lol
		switch media.Type {
		case "video", "gifv":
			files[i].UIIsVideo = true
		}
	}

	return files, nil
}
