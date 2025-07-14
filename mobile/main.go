package main

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/makinori/maki-immich/immich"
	"github.com/makinori/maki-immich/mobile/android"
	"github.com/makinori/maki-immich/mobile/makitheme"
	"github.com/makinori/maki-immich/scrape"
)

var (
	fyneApp fyne.App
	window  fyne.Window

	currentIntent *android.Intent
	currentFiles  []immich.File
)

func showUnknownIntent() {
	lines := []string{
		"action: " + currentIntent.Action,
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
	showScreenError("unknown intent", strings.Join(lines, "\n"))
}

func showFetchingImages(from string) {
	showScreenError(
		"fetching", "from "+from+"...",
		ScreenTextOptionNoError,
		ScreenTextOptionNoDismiss,
	)
}

func loop() {
	if currentIntent != nil {
		return
	}

	intent := android.GetIntent()
	if intent.Action != android.ACTION_SEND &&
		intent.Action != android.ACTION_SEND_MULTIPLE {
		return
	}

	currentIntent = &intent

	if intent.Type == "text/plain" {
		intentURL, err := url.Parse(intent.Text)
		if err != nil {
			showScreenError("failed to parse url", err.Error())
			return
		}

		tryScrape := func(
			name string, hosts []string,
			scrapeFn func(url *url.URL) ([]immich.File, error),
		) bool {
			if !slices.Contains(hosts, intentURL.Host) {
				return false
			}

			showFetchingImages(name)

			currentFiles, err = scrapeFn(intentURL)
			if err == nil && len(currentFiles) == 0 {
				err = errors.New("no images")
			}
			if err != nil {
				showScreenError("failed to retrieve", err.Error())
				return true
			}

			showScreenAlbumSelector()
			return true
		}

		if tryScrape("twitter", scrape.TwitterHosts, scrape.VXTwitter) {
			return
		}

		showScreenError("unknown url", intentURL.String())

	} else if strings.HasPrefix(intent.Type, "image/") ||
		strings.HasPrefix(intent.Type, "video/") {

		showFetchingImages("content://")

		currentFiles = make([]immich.File, len(intent.URI))

		for i, uri := range intent.URI {
			data := android.ReadContent(uri)
			if len(data) == 0 {
				showScreenError(
					"failed to read content",
					"returned 0 for uri:\n"+uri,
				)
				return
			}

			currentFiles[i] = immich.File{
				Name: path.Base(uri), Data: data,
			}
		}

		showScreenAlbumSelector()
	} else {
		showUnknownIntent()
	}
}

func main() {
	fyneApp = app.New()
	fyneApp.Settings().SetTheme(&makitheme.Theme{})

	window = fyneApp.NewWindow("maki immich")

	if runtime.GOOS == "linux" {
		window.Resize(fyne.Size{
			Width:  300,
			Height: 500,
		})
	}

	showScreenError(
		"maki immich", "share an image to this app",
		ScreenTextOptionNoError,
	)

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
