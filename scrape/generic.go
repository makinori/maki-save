package scrape

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/makinori/maki-immich/immich"
)

func acceptableMediaContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}

func TestGeneric(contentURL *url.URL, extraData *[]byte) bool {
	// extraData might contain data from previous response with a different
	// accept header. check if this is already content we can upload

	contentType := http.DetectContentType(*extraData)

	if acceptableMediaContentType(contentType) {
		return true
	}

	// ok request normally instead

	res, err := http.Get(contentURL.String())
	if err != nil {
		return false
	}
	defer res.Body.Close()

	*extraData, err = io.ReadAll(res.Body)
	if err != nil {
		return false
	}

	contentType = http.DetectContentType(*extraData)

	return acceptableMediaContentType(contentType)
}

func Generic(contentURL *url.URL, extraData *[]byte) ([]immich.File, error) {
	file := immich.File{
		Data: *extraData,
		Name: path.Base(contentURL.Path),
	}
	return []immich.File{file}, nil
}
