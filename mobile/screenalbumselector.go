package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/makinori/maki-immich/immich"
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
		image, ok := cachedFileCanvasImageMap[file]
		if !ok {
			var reader io.Reader
			if len(file.Thumbnail) > 0 {
				reader = bytes.NewReader(file.Thumbnail)
			} else {
				reader = bytes.NewReader(file.Data)
			}

			image = canvas.NewImageFromReader(reader, file.Name)
			image.FillMode = canvas.ImageFillContain
			cachedFileCanvasImageMap[file] = image
		}
		fyne.Do(func() {
			imageStack.Objects[1] = image
		})
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

	albumDisableSelect := binding.NewBool()

	var onAlbumSelected func(i int)

	albumRadioList := radioList(
		albumNames,
		albumDisableSelect,
		&onAlbumSelected,
		func() {
			os.Exit(0)
		},
	)

	uploadingLabel := widget.NewLabel("uploading...")

	var box *fyne.Container

	onAlbumSelected = func(i int) {
		go func() {
			fyne.Do(func() {
				box.Objects[0] = container.NewCenter(uploadingLabel)
			})

			messages := immich.UploadFiles(albums[i], currentFiles)

			fyne.Do(func() {
				uploadingLabel.SetText("")

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
		}()
	}

	var updateImagesGrid func(errIfNoneLeft bool)

	updateImagesGrid = func(errIfNoneLeft bool) {
		if errIfNoneLeft {
			showScreenError(ScreenError{Text: []string{
				"no files", "none at all",
			}})
			return
		}

		if len(currentFiles) == 0 {
			albumDisableSelect.Set(true)
		} else {
			albumDisableSelect.Set(false)
		}

		imagesGrid := getImagesGrid(currentFiles, func(f *immich.File) {
			i := slices.Index(currentFiles, f)
			if i == -1 {
				return
			}
			currentFiles = slices.Delete(currentFiles, i, i+1)
			updateImagesGrid(true)
		})

		box = container.NewBorder(imagesGrid, nil, nil, nil, albumRadioList)

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
