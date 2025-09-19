package main

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/makinori/maki-immich/immich"
	"golang.org/x/image/webp"
)

func bytesToString(bytes int) string {
	s := float32(bytes)
	if s < 1000 {
		return fmt.Sprintf("%.0f bytes", s)
	}
	s /= 1000
	if s < 1000 {
		return fmt.Sprintf("%.1f kB", s)
	}
	s /= 1000
	return fmt.Sprintf("%.1f MB", s)
}

/*
func getImageBgColor() color.NRGBA {
	r, g, b, a := fyne.CurrentApp().Settings().Theme().Color(
		theme.ColorNameDisabledButton, fyne.CurrentApp().Settings().ThemeVariant(),
	).RGBA()
	var alpha float32 = 0.5
	return color.NRGBA{
		R: uint8((float32(r) / math.MaxUint16) * 255),
		G: uint8((float32(g) / math.MaxUint16) * 255),
		B: uint8((float32(b) / math.MaxUint16) * 255),
		A: uint8((float32(a) / math.MaxUint16) * alpha * 255),
	}
}
*/

var cachedFileCanvasImageMap = map[*immich.File]*canvas.Image{}

func getCanvasImage(file *immich.File) *canvas.Image {
	canvasImage, ok := cachedFileCanvasImageMap[file]
	if ok {
		return canvasImage
	}

	var data []byte
	if len(file.Thumbnail) > 0 {
		data = file.Thumbnail
	} else {
		data = file.Data
	}

	contentType := http.DetectContentType(data)

	if contentType == "image/webp" {
		image, err := webp.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Println(err)
			return nil
		}
		canvasImage = canvas.NewImageFromImage(image)
	} else {
		canvasImage = canvas.NewImageFromReader(
			bytes.NewReader(data), file.Name,
		)
	}

	canvasImage.FillMode = canvas.ImageFillContain

	cachedFileCanvasImageMap[file] = canvasImage

	return canvasImage
}

func getImageWidget(file *immich.File, onClick func()) fyne.CanvasObject {
	// background := canvas.NewRectangle(getImageBgColor())
	button := widget.NewButton("", onClick)

	emptyImage := canvas.NewRectangle(color.NRGBA{})
	imageStack := container.NewStack(
		// background,
		button,
		emptyImage,
	)

	if len(file.Thumbnail) > 0 {
		icon := widget.NewIcon(theme.MediaVideoIcon())
		imageStack.Add(container.NewCenter(NewMinSize(
			fyne.Size{Width: 36, Height: 36},
			icon,
		)))
	}

	go func() {
		image := getCanvasImage(file)
		if image != nil {
			fyne.Do(func() {
				imageStack.Objects[1] = image
			})
		}
	}()

	return imageStack
}

func getImagesGrid(files []*immich.File, onClick func(*immich.File)) fyne.CanvasObject {
	var minHeight float32 = 300

	if len(files) == 0 {
		return NewMinSize(fyne.Size{Height: minHeight},
			container.NewCenter(widget.NewLabelWithData(fetchingText)),
		)
	}

	cols := int(math.Ceil(math.Sqrt(float64(len(files)))))
	images := make([]fyne.CanvasObject, len(files))
	for i, file := range files {
		images[i] = getImageWidget(file, func() {
			onClick(file)
		})
	}

	imagesGrid := container.NewGridWithColumns(cols, images...)

	textNames := ""
	textSizes := ""
	for _, file := range files {
		textNames += file.Name + "\n"
		textSizes += "(" + bytesToString(len(file.Data)) + ")\n"
	}
	textNames = strings.TrimSpace(textNames)
	textSizes = strings.TrimSpace(textSizes)

	labelNames := widget.NewLabel(textNames)
	labelNames.Truncation = fyne.TextTruncateEllipsis

	labelSizes := widget.NewLabel(textSizes)

	textContainer := container.NewBorder(
		nil, nil, nil, labelSizes, labelNames,
	)

	return NewMinSize(fyne.Size{Height: minHeight},
		container.NewBorder(
			nil, textContainer, nil, nil, imagesGrid,
		),
	)
}

var albums []immich.Album

func showScreenAlbumSelector() bool {
	if len(albums) == 0 {
		var err error
		albums, err = immich.GetAlbums()
		if err != nil {
			showScreenError(ScreenError{Text: []string{
				"failed to get albums", err.Error(),
			}})
			return false
		}
	}

	albumNames := make([]string, len(albums))
	for i, album := range albums {
		albumNames[i] = album.AlbumName
	}

	// albumDisableSelect := binding.NewBool()

	var onAlbumSelected func(i int)

	albumRadioList := radioList(
		albumNames,
		// albumDisableSelect,
		&onAlbumSelected,
		func() {
			os.Exit(0)
		},
	)

	var box *fyne.Container

	uploading := false
	selectedAlbumIndex := -1

	uploadNow := func() {
		if uploading || selectedAlbumIndex < 0 {
			return
		}

		uploading = true

		// reverse so it appears in the same order when uploaded
		currentFilesReversed := currentFiles
		slices.Reverse(currentFilesReversed)

		messages := immich.UploadFiles(
			albums[selectedAlbumIndex], currentFilesReversed,
		)

		fyne.Do(func() {
			// infoDialog := dialog.NewInformation("Info", info, window)
			// infoDialog.SetOnClosed(func() {
			// 	os.Exit(0)
			// })
			// infoDialog.Show()

			messages[0] = "#" + messages[0]
			showScreenError(ScreenError{
				Text: messages,
				// we want it to self destruct
			})
		})
	}

	var updateImagesGrid func(errIfNone bool)

	onAlbumSelected = func(i int) {
		selectedAlbumIndex = i

		updateImagesGrid(false)

		if len(currentFiles) > 0 {
			go uploadNow()
		}
	}

	updateImagesGrid = func(errIfNone bool) {
		if errIfNone && len(currentFiles) == 0 {
			showScreenError(ScreenError{Text: []string{
				"no files", "none at all",
			}})
			return
		}

		imagesGrid := getImagesGrid(currentFiles, func(f *immich.File) {
			i := slices.Index(currentFiles, f)
			if i == -1 {
				return
			}
			currentFiles = slices.Delete(currentFiles, i, i+1)
			updateImagesGrid(true)
		})

		centerContainer := albumRadioList

		if selectedAlbumIndex >= 0 {
			text := ""
			if len(currentFiles) > 0 {
				text = "uploading to "
			} else {
				text = "about to upload to "
			}
			text += albums[selectedAlbumIndex].AlbumName + "..."

			centerContainer = container.NewCenter(widget.NewLabel(text))

			if len(currentFiles) > 0 {
				go uploadNow()
			}
		}

		box = container.NewBorder(imagesGrid, nil, nil, nil, centerContainer)

		fyne.DoAndWait(func() {
			window.SetContent(box)
		})
	}

	updateImagesGrid(false)

	go func() {
		<-currentFilesChanged

		if len(currentFiles) == 0 {
			return
		}

		updateImagesGrid(false)
	}()

	return true
}
