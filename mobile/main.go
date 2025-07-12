package main

import (
	_ "embed"
	"errors"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/makinori/maki-immich/immich"
	"github.com/makinori/maki-immich/mobile/android"
	"github.com/makinori/maki-immich/mobile/scrape"
)

var (
	fyneApp fyne.App
	window  fyne.Window

	currentIntent *android.Intent
	currentFiles  []immich.File
)

func showUnknownIntent() {
	showScreenError("unknown intent", strings.Join([]string{
		"action: " + currentIntent.Action,
		"type: " + currentIntent.Type,
		"uri: " + currentIntent.URI,
		"text: " + currentIntent.Text,
	}, "\n"))
}

func showFetchingImages(from string) {
	showScreenError(
		"fetching images", "from "+from+"...",
		ScreenTextOptionNoError,
		ScreenTextOptionNoSelfDestruct,
		ScreenTextOptionNoDismiss,
	)
}

func loop() {
	if currentIntent != nil {
		return
	}

	intent := android.GetIntent()
	if intent.Action != android.ACTION_SEND {
		return
	}

	currentIntent = &intent

	if intent.Type == "text/plain" {
		url, err := url.Parse(intent.Text)
		if err != nil {
			showScreenError("failed to parse url", err.Error())
			return
		}

		if slices.Contains(scrape.TwitterHosts, url.Host) {
			showFetchingImages("twitter")

			currentFiles, err = scrape.FromTwitter(url)
			if err == nil && len(currentFiles) == 0 {
				err = errors.New("no images")
			}
			if err != nil {
				showScreenError("failed to retrieve", err.Error())
				return
			}

			showScreenAlbumSelector()
		} else {
			showScreenError("unknown url", url.String())
		}
	} else if strings.HasPrefix(intent.Type, "image/") {
		data := android.ReadContent(intent.URI)
		if len(data) == 0 {
			showUnknownIntent()
			return
		}

		currentFiles = []immich.File{{
			Name: path.Base(intent.URI), Data: data,
		}}

		showScreenAlbumSelector()
	} else {
		showUnknownIntent()
	}
}

func main() {
	fyneApp = app.New()
	window = fyneApp.NewWindow("maki immich")

	showScreenError(
		"maki immich", "share an image to this app",
		ScreenTextOptionNoError,
		ScreenTextOptionNoSelfDestruct,
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
