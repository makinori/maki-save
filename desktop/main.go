package main

import (
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/makinori/maki-immich/immich"
)

func selectAlbumKdialog() (immich.Album, error) {
	albums, err := immich.GetAlbums()
	if err != nil {
		return immich.Album{}, err
	}

	kdialogArgs := []string{
		"--radiolist", "Select an album to upload to",
	}

	for _, album := range albums {
		kdialogArgs = append(kdialogArgs, album.Id, album.AlbumName, "off")
	}

	cmd := exec.Command("kdialog", kdialogArgs...)

	output, err := cmd.Output()
	if err != nil {
		return immich.Album{}, err
	}

	albumId := strings.TrimSpace(string(output))

	for _, album := range albums {
		if album.Id == albumId {
			return album, nil
		}
	}

	return immich.Album{}, errors.New("failed to match album id")
}

func dialog(message string, isError bool) {
	var args []string

	if isError {
		args = []string{"--error", message}
	} else {
		args = []string{"--msgbox", message}
	}

	cmd := exec.Command("kdialog", args...)
	cmd.Run()
}

func main() {
	// get files

	if len(os.Args) <= 1 {
		dialog("Please select one or more files", true)
		os.Exit(1)
	}

	filePaths := os.Args[1:]

	_, usingNautilus := os.LookupEnv("NAUTILUS")
	if usingNautilus {
		fileUrlPaths := strings.Split(filePaths[0], " ")
		filePaths = []string{}

		for _, fileUrlStr := range fileUrlPaths {
			fileUrl, _ := url.Parse(fileUrlStr)
			filePaths = append([]string{fileUrl.Path}, filePaths...)
		}
	}

	// get album id

	album, err := selectAlbumKdialog()

	if err != nil {
		// didnt select
		if err.Error() == "exit status 1" || err.Error() == "no album selected" {
			os.Exit(0)
		}

		dialog(err.Error(), true)
		os.Exit(1)
	}

	// upload files

	files := make([]*immich.File, len(filePaths))

	for i, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		files[i] = &immich.File{
			Data: data,
			Name: path.Base(filePath),
			Err:  err,
		}
	}

	// display

	messages := immich.UploadFiles(album, files)

	var outputMsg string
	for i, msg := range messages {
		outputMsg += msg + "\n"
		if i%2 == 1 {
			outputMsg += "\n"
		}
	}
	outputMsg = strings.TrimSpace(outputMsg)

	dialog(outputMsg, false)
}
