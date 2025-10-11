//go:build wasm

// dev tools in about:debugger

package main

import (
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"syscall/js"

	"github.com/makinori/maki-immich/immich"
	"github.com/makinori/maki-immich/scrape"
	"github.com/makinori/maki-immich/webext/jsutil"
)

func immichFilesToJS(files []immich.File) js.Value {
	values := make([]js.Value, len(files))
	for i, file := range files {
		uiErr := ""
		if file.UIErr != nil {
			uiErr = file.UIErr.Error()
		}

		values[i] = js.ValueOf(map[string]interface{}{
			"data": jsutil.BytesToValue(file.Data),
			"name": file.Name,
			// file.Description,
			"uiErr": uiErr,
			// UIThumbnail
			// UIIsVideo
		})
	}
	return jsutil.ArrayToValue(values)
}

func scrapeURL(args []js.Value) (ret js.Value, ok bool) {
	if len(args) < 1 {
		return js.ValueOf("not enough args"), false
	}

	urlString := args[0].String()

	scrapeURL, err := url.Parse(urlString)
	if err != nil {
		return js.ValueOf(err.Error()), false
	}

	name, scrapeFn := scrape.Test(scrapeURL)

	files, err := scrapeFn(scrapeURL)
	if err != nil {
		return js.ValueOf(err.Error()), false
	}

	return js.ValueOf(map[string]interface{}{
		"name":  name,
		"files": immichFilesToJS(files),
	}), true
}

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	global := js.Global()
	global.Set("wasmScrapeURL", jsutil.AsyncFunc(scrapeURL))

	<-done
}
