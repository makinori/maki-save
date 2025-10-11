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
	"unsafe"

	"github.com/makinori/maki-save/immich"
)

var (
	activityPubNoteIDRegexp = regexp.MustCompile(`\/([^\/]+)$`)

	mastodonVideoURLRegexp = regexp.MustCompile(`(^https:\/\/.+\/system\/media_attachments\/files(?:\/[0-9]+)+)\/original\/(.+)\.mp4$`)
)

type activityPubAttachment struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
	// BlurHash  string `json:"blurhash"`
	// Width     int    `json:"width"`
	// Height    int    `json:"height"`
}

type activityPubNote struct {
	ID           string                  `json:"id"`
	AttributedTo string                  `json:"attributedTo"`
	Attachments  []activityPubAttachment `json:"attachment"`
}

type activityPubPerson struct {
	PreferredUsername string `json:"preferredUsername"`
}

type webFingerResponse struct {
	Subject string `json:"subject"`
}

func getActivityPub(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/activity+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	isActivityJSON := false
	for _, contentType := range res.Header["Content-Type"] {
		if strings.Contains(contentType, "application/activity+json") {
			isActivityJSON = true
			break
		}
	}
	if !isActivityJSON {
		return nil, errors.New("response not activity json")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func TestActivityPub(url *url.URL, extraData *unsafe.Pointer) bool {
	data, err := getActivityPub(url.String())
	*extraData = unsafe.Pointer(&data)
	return err == nil
}

func guessActivityPubThumbnailFromURL(mediaURL string) string {
	mastodonVideoURLMatches := mastodonVideoURLRegexp.FindStringSubmatch(mediaURL)
	if len(mastodonVideoURLMatches) > 0 {
		return mastodonVideoURLMatches[1] + "/small/" +
			mastodonVideoURLMatches[2] + ".png"
	}
	return ""
}

func activityPubResolveHandle(username, host string) (string, error) {
	acct := fmt.Sprintf("acct:%s@%s", username, host)

	webFingerQuery := url.Values{}
	webFingerQuery.Add("resource", acct)

	var webFingerURL url.URL
	webFingerURL.Host = host
	webFingerURL.Scheme = "https"
	webFingerURL.Path = "/.well-known/webfinger"
	webFingerURL.RawQuery = webFingerQuery.Encode()

	webFingerRes, err := http.Get(webFingerURL.String())
	if err != nil {
		return "", err
	}
	defer webFingerRes.Body.Close()

	webFingerData, err := io.ReadAll(webFingerRes.Body)
	if err != nil {
		return "", err
	}

	var webFinger webFingerResponse
	err = json.Unmarshal(webFingerData, &webFinger)
	if err != nil {
		return "", err
	}

	return "@" + strings.TrimPrefix(webFinger.Subject, "acct:"), nil
}

func ActivityPub(noteURL *url.URL, noteData unsafe.Pointer) ([]immich.File, error) {
	// os.WriteFile("test.json", *noteData, 0644)

	var note activityPubNote
	err := json.Unmarshal(*(*[]byte)(noteData), &note)
	if err != nil {
		return []immich.File{}, err
	}

	noteIDMatches := activityPubNoteIDRegexp.FindStringSubmatch(note.ID)
	if len(noteIDMatches) == 0 {
		return []immich.File{}, errors.New("failed to find id")
	}

	noteID := noteIDMatches[1]

	// resolve handle with webfinger
	// e.g. mastodon.hotmilk.space is actually just hotmilk.space

	personData, err := getActivityPub(note.AttributedTo)
	if err != nil {
		return []immich.File{}, err
	}

	var person activityPubPerson
	err = json.Unmarshal(personData, &person)
	if err != nil {
		return []immich.File{}, err
	}

	handle, err := activityPubResolveHandle(
		person.PreferredUsername, noteURL.Host,
	)
	if err != nil {
		return []immich.File{}, err
	}

	// get files

	var foundAttachments []activityPubAttachment
	for _, attachment := range note.Attachments {
		if attachment.Type != "Document" {
			continue
		}
		foundAttachments = append(foundAttachments, attachment)
	}

	fileURLs := make([]string, len(foundAttachments))
	thumbnailURLs := make([]string, len(foundAttachments))

	for i, attachment := range foundAttachments {
		fileURLs[i] = attachment.URL

		if !strings.HasPrefix(attachment.MediaType, "image/") {
			thumbnailURLs[i] = guessActivityPubThumbnailFromURL(attachment.URL)
		}
	}

	prefix := fmt.Sprintf("%s-%s-", handle, noteID)

	files := getFilesFromURLs(prefix, fileURLs, thumbnailURLs)

	for i, attachment := range foundAttachments {
		if strings.HasPrefix(attachment.MediaType, "video/") {
			files[i].UIIsVideo = true
		}
	}

	return files, nil
}
