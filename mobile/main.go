package main

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"github.com/makinori/maki-save/ffmpeg"
	"github.com/makinori/maki-save/immich"
	"github.com/makinori/maki-save/mobile/android"
	"github.com/makinori/maki-save/mobile/makitheme"
	"github.com/makinori/maki-save/scrape"
	"mvdan.cc/xurls/v2"
)

var (
	fyneApp fyne.App
	window  fyne.Window

	currentIntent       *android.Intent
	currentFiles        []*immich.File
	currentFilesChanged chan struct{}

	fetchingText binding.String

	usingDesktop = runtime.GOOS != "android"
)

const (
	BUTTON_HEIGHT_MUL = 1.25
)

func showUnknownIntent() {
	lines := []string{
		"action: " + string(currentIntent.Action),
		"type: " + currentIntent.Type,
	}
	if len(currentIntent.URI) == 0 {
		lines = append(lines, "no uris")
	} else {
		for i, uri := range currentIntent.URI {
			lines = append(lines, fmt.Sprintf("uri[%d]: %s", i, uri))
		}
	}
	lines = append(lines, "text: "+currentIntent.Text)
	showScreenError(ScreenError{Text: []string{
		"unknown intent", strings.Join(lines, "\n"),
	}})
}

func setFetchingText(from string) {
	fetchingText.Set("fetching from " + from + "...")
}

func handleTextIntent() {
	foundURL := xurls.Relaxed().FindString(currentIntent.Text)
	if foundURL == "" {
		showScreenError(ScreenError{Text: []string{
			"failed to find url",
		}})
		return
	}

	intentURL, err := url.Parse(foundURL)
	if err != nil {
		showScreenError(ScreenError{Text: []string{
			"failed to parse url", err.Error(),
		}})
		return
	}

	name, scrapeFn := scrape.Test(intentURL)
	if name == "" {
		showScreenError(ScreenError{Text: []string{
			"unknown url", intentURL.String(),
		}})
		return
	}

	setFetchingText(strings.ToLower(name))

	files, err := scrapeFn(intentURL)
	if err == nil && len(files) == 0 {
		err = errors.New("no images")
	}
	if err != nil {
		showScreenError(ScreenError{Text: []string{
			"failed to retrieve", err.Error(),
		}})
		return
	}

	currentFiles = make([]*immich.File, len(files))
	for i, file := range files {
		currentFiles[i] = &file
	}

	currentFilesChanged <- struct{}{}
}

func handleMediaIntent() {
	setFetchingText("content://")

	currentFiles = make([]*immich.File, len(currentIntent.URI))

	var wg sync.WaitGroup
	for i, uri := range currentIntent.URI {
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, filename := android.ReadContent(uri)
			if len(data) == 0 {
				showScreenError(ScreenError{Text: []string{
					"failed to read content", "returned 0 for uri:\n" + uri,
				}})
				return
			}

			unescapedFilename, err := url.PathUnescape(filename)
			if err == nil {
				filename = unescapedFilename
			}

			currentFiles[i] = &immich.File{
				Name: filename, Data: data,
			}

			// generate thumbnail if video

			contentType := http.DetectContentType(data)

			// some videos aren't being recognized,
			// so handle application/octet-stream too
			if !strings.HasPrefix(contentType, "video/") &&
				contentType != "application/octet-stream" {
				return
			}

			currentFiles[i].UIIsVideo = true

			thumbnail, err := ffmpeg.GetMiddleFrameFromVideo(data)
			if err != nil {
				fmt.Println("failed to get thumbnail: " + err.Error())
			} else {
				currentFiles[i].UIThumbnail = thumbnail
			}
		}()
	}

	wg.Wait()

	currentFilesChanged <- struct{}{}
}

var showingIntroScreen = false

func loop() {
	if currentIntent != nil {
		return
	}

	intent := android.GetIntent()

	if !slices.Contains(android.Actions, intent.Action) {
		if !showingIntroScreen {
			showScreenError(ScreenError{
				Text: []string{
					"#maki save", "share an image to this app",
				},
				// will conditionally self destruct below in main()
				NoSelfDestruct: true,
			})
			showingIntroScreen = true
		}
		return
	}

	currentIntent = &intent
	currentFilesChanged = make(chan struct{}, 1)

	fetchingText = binding.NewString()
	fetchingText.Set("loading...")

	intentType := strings.SplitN(intent.Type, "/", 2)[0]

	if intentType == "image" || intentType == "video" {
		if showScreenAlbumSelector() {
			go handleMediaIntent()
		}
	} else if intentType == "text" || intent.Action == android.ACTION_SENDTO {
		if showScreenAlbumSelector() {
			go handleTextIntent()
		}
	} else {
		showUnknownIntent()
	}
}

func main() {
	fyneApp = app.New()
	fyneApp.Settings().SetTheme(&makitheme.Theme{})

	window = fyneApp.NewWindow("maki save")

	if usingDesktop {
		window.Resize(fyne.Size{
			Width:  400,
			Height: 700,
		})
	}

	go func() {
		for {
			loop()
			if currentIntent != nil {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	// auto close program if no intent after some time
	// cause we're polling, don't want to poll for no reason
	go func() {
		time.Sleep(time.Second * 10)
		if currentIntent == nil && len(currentFiles) == 0 {
			os.Exit(0)
		}
	}()

	window.ShowAndRun()
}
