package main

import (
	_ "embed"
	"os"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/makinori/maki-immich/mobile/android"
)

var (
	fyneApp fyne.App
	window  fyne.Window

	currentIntent *android.Intent
	currentData   []byte
	currentName   string
)

func loop() {
	if currentIntent != nil {
		return
	}

	intent := android.GetIntent()
	if intent.Action != "android.intent.action.SEND" || intent.URI == "" {
		return
	}

	data := android.ReadContent(intent.URI)
	if len(data) == 0 {
		return
	}

	currentIntent = &intent
	currentData = data
	currentName = path.Base(intent.URI)

	showAlbumSelector()
}

func main() {
	fyneApp = app.New()
	window = fyneApp.NewWindow("maki immich")

	showDefaultScreen()

	go func() {
		for {
			fyne.Do(loop)
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
		if currentIntent == nil {
			os.Exit(0)
		}
	}()

	window.ShowAndRun()
}
