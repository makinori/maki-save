package scrape

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"unsafe"

	"github.com/gabriel-vasile/mimetype"
	"github.com/makinori/maki-save/immich"
)

func acceptableMediaContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}

func TestGeneric(contentURL *url.URL, extraData *unsafe.Pointer) bool {
	res, err := http.Get(contentURL.String())
	if err != nil {
		return false
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return false
	}

	mime := mimetype.Detect(data)
	if mime == nil || !acceptableMediaContentType(mime.String()) {
		return false
	}

	*extraData = unsafe.Pointer(&data)

	return true
}

func Generic(contentURL *url.URL, extraData unsafe.Pointer) ([]immich.File, error) {
	if extraData == nil {
		return []immich.File{}, errors.New("extra data is nil")
	}

	file := immich.File{
		Data: *(*[]byte)(extraData),
		Name: path.Base(contentURL.Path),
	}

	mime := mimetype.Detect(file.Data)
	if mime != nil && strings.HasPrefix(mime.String(), "video/") {
		file.UIIsVideo = true
		file.UIThumbnail, _ = getMiddleFrameFromVideo(file.Data)
	}

	return []immich.File{file}, nil
}
