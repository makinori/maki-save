package immich

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"time"
)

var (
	//go:embed server.txt
	IMMICH_SERVER string
	//go:embed key.txt
	IMMICH_API_KEY string

	ErrDuplicate = errors.New("duplicate asset")
)

func init() {
	IMMICH_SERVER = strings.TrimSpace(IMMICH_SERVER)
	IMMICH_SERVER = strings.TrimSuffix(IMMICH_SERVER, "/")
	IMMICH_API_KEY = strings.TrimSpace(IMMICH_API_KEY)
}

type Album struct {
	AlbumName         string    `json:"albumName"`
	Id                string    `json:"id"`
	AssetCount        int       `json:"assetCount"`
	LastModifiedAsset time.Time `json:"lastModifiedAssetTimestamp"`
}

type Action struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

func getRandom() (string, error) {
	bytes := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

/*
// doesnt handle all cases. this isn't very reliable

func stripExifDate(data []byte) ([]byte, error) {
	// find more using: exiftool -G -a -s -time:all image.jpg

	toStrip := []string{
		"AllDates",
		"CreateDate",
		"DateCreated",
		// "DateTimeCreated", // not writeable
		"DateTimeDigitized",
		"DateTimeOriginal",
		"GPSDateStamp",
		"GPSDateTime",
		"GPSTimeStamp",
		"ModifyDate",
		"TimeCreated",
		"DigitalCreationDate",
		"DigitalCreationTime",
		// "DigitalCreationDateTime", // not writeable
	}

	var args []string

	for _, tag := range toStrip {
		args = append(args, "-"+tag+"=")
	}

	args = append(args, "-")

	output := new(bytes.Buffer)

	cmd := exec.Command("exiftool", args...)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = output
	cmd.Stderr = io.Discard

	err := cmd.Run()

	if err != nil {
		return []byte{}, err
	}

	return output.Bytes(), nil
}
*/

func GetAlbums() ([]Album, error) {
	req, err := http.NewRequest("GET", IMMICH_SERVER+"/api/albums", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", IMMICH_API_KEY)

	res, err := new(http.Client).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var albums []Album

	err = json.Unmarshal(bytes, &albums)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(albums, func(a, b Album) int {
		return b.LastModifiedAsset.Compare(a.LastModifiedAsset)
	})

	return albums, nil
}

func uploadAsset(data []byte, filename string, dateStr string) (string, error) {
	mpBuf := new(bytes.Buffer)
	mp := multipart.NewWriter(mpBuf)

	mpFile, err := mp.CreateFormFile("assetData", filename)
	if err != nil {
		return "", err
	}

	mpFile.Write(data)

	deviceAssetId, err := getRandom()
	if err != nil {
		return "", err
	}

	mp.WriteField("deviceAssetId", deviceAssetId)
	mp.WriteField("deviceId", "GO") // WEB

	// doesn't actually set the date
	mp.WriteField("fileCreatedAt", dateStr)
	mp.WriteField("fileModifiedAt", dateStr)

	mp.Close()

	req, err := http.NewRequest("POST", IMMICH_SERVER+"/api/assets", mpBuf)

	if err != nil {
		return "", err
	}

	req.Header.Add("x-api-key", IMMICH_API_KEY)
	req.Header.Add("Content-Type", mp.FormDataContentType())

	res, err := new(http.Client).Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var action Action

	err = json.Unmarshal(bytes, &action)
	if err != nil {
		return "", err
	}

	if action.Error != "" {
		return "", errors.New(action.Error)
	}

	return action.Id, nil
}

func updateAssetDate(assetId string, dateStr string) error {
	data, err := json.Marshal(map[string]any{
		"ids":              []string{assetId},
		"dateTimeOriginal": dateStr,
	})

	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	buffer.Write(data)

	req, err := http.NewRequest(
		"PUT", IMMICH_SERVER+"/api/assets", buffer,
	)

	if err != nil {
		return err
	}

	req.Header.Add("x-api-key", IMMICH_API_KEY)
	req.Header.Add("Content-Type", "application/json")

	res, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		return fmt.Errorf("status code: %d", res.StatusCode)
	}

	return nil
}

func addToAlbum(albumId string, assetId string) error {
	data, err := json.Marshal(map[string][]string{
		"ids": {assetId},
	})

	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	buffer.Write(data)

	req, err := http.NewRequest(
		"PUT", IMMICH_SERVER+"/api/albums/"+albumId+"/assets", buffer,
	)

	if err != nil {
		return err
	}

	req.Header.Add("x-api-key", IMMICH_API_KEY)
	req.Header.Add("Content-Type", "application/json")

	res, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var actions []Action

	err = json.Unmarshal(bytes, &actions)
	if err != nil {
		return err
	}

	if len(actions) == 0 {
		return errors.New("bad response")
	}

	action := actions[0]

	if action.Error == "duplicate" {
		return ErrDuplicate
	}

	if action.Error != "" {
		return errors.New(action.Error)
	}

	return nil
}

type File struct {
	Data []byte
	Name string
	Err  error // if failed to read
}

func UploadFile(album Album, file File) error {
	// try to strip exif
	// ignore error
	// {
	// 	strippedData, err := stripExifDate(fileData)
	// 	if err == nil {
	// 		fileData = strippedData
	// 	}
	// }

	fileDateStr := time.Now().Format(time.RFC3339Nano)

	// upload file which also handles deduplication

	assetId, err := uploadAsset(file.Data, file.Name, fileDateStr)
	if err != nil {
		return err
	}

	// add to album

	err = addToAlbum(album.Id, assetId)
	if err != nil {
		return err
	}

	// if there's a duplicate error, it will have returned early
	// update file date now so we dont push duplicates up

	// ignore error but do retry a few times just incase
	retryNoFailNoOutput(3, time.Millisecond*500, func() error {
		return updateAssetDate(assetId, fileDateStr)
	})

	return nil
}

func UploadFiles(album Album, files []File) string {
	// upload files

	var completed []string
	var failed []string

	for _, file := range files {
		var err error
		if file.Err != nil {
			err = file.Err
		} else {
			err = UploadFile(album, file)
		}

		switch err {
		case ErrDuplicate:
			completed = append(completed, file.Name+" (duplicate)")
		case nil:
			completed = append(completed, file.Name)
		default:
			// some other error
			failed = append(failed, file.Name)
		}
	}

	finalMsg := fmt.Sprintf("added %d to: %s", len(completed), album.AlbumName)

	if len(completed) > 0 {
		finalMsg += "\n" + strings.Join(completed, "\n")
	}

	if len(failed) > 0 {
		finalMsg += "\n\nfailed:\n" + strings.Join(failed, "\n")
	}

	return finalMsg
}
