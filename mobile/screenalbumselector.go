package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

func getImageWidget(file immich.File) fyne.CanvasObject {
	background := canvas.NewRectangle(getImageBgColor())

	emptyImage := canvas.NewRectangle(color.NRGBA{})
	imageStack := container.NewStack(
		background,
		emptyImage,
	)

	go func() {
		image := canvas.NewImageFromReader(bytes.NewReader(file.Data), file.Name)
		image.FillMode = canvas.ImageFillContain
		// image.SetMinSize(emptyImage.MinSize())
		fyne.Do(func() {
			imageStack.Objects[1] = image
		})
	}()

	return imageStack
}

func getImagesGrid(files []immich.File) fyne.CanvasObject {
	cols := int(math.Ceil(math.Sqrt(float64(len(files)))))
	images := make([]fyne.CanvasObject, len(files))
	for i, file := range files {
		images[i] = getImageWidget(file)
	}

	imagesGrid := NewFixedSize(fyne.Size{Height: 150},
		container.NewGridWithColumns(cols, images...),
	)

	text := ""
	for _, file := range files {
		text += fmt.Sprintf("%s (%s)\n",
			file.Name, bytesToString(len(file.Data)),
		)
	}
	text = strings.TrimSpace(text)

	label := widget.NewLabel(text)

	return container.NewBorder(
		nil, container.NewCenter(label), nil, nil, imagesGrid,
	)
}

func showError(err string) {
	errorDialog := dialog.NewError(errors.New(err), window)
	errorDialog.SetOnClosed(func() { os.Exit(1) })
	errorDialog.Show()
}

func showScreenAlbumSelector() {
	if len(currentFiles) == 0 {
		showScreenError("missing files", "can't display album selector")
		return
	}

	imagesGrid := getImagesGrid(currentFiles)

	albums, err := immich.GetAlbums()
	if err != nil {
		showError("failed to get albums: " + err.Error())
		return
	}

	albumNames := make([]string, len(albums))
	for i, album := range albums {
		albumNames[i] = album.AlbumName
	}

	var onSelect func(i int)

	albumRadioList := radioList(
		albumNames,
		&onSelect,
		func() {
			os.Exit(0)
		},
	)

	uploadingLabel := widget.NewLabel("Uploading...")

	box := container.NewBorder(imagesGrid, nil, nil, nil, albumRadioList)

	onSelect = func(i int) {
		go func() {
			fyne.Do(func() {
				box.Objects[0] = container.NewCenter(uploadingLabel)
			})

			info := immich.UploadFiles(albums[i], currentFiles)

			fyne.Do(func() {
				uploadingLabel.SetText("")

				infoDialog := dialog.NewInformation("Info", info, window)
				infoDialog.SetOnClosed(func() {
					os.Exit(0)
				})
				infoDialog.Show()
			})
		}()
	}

	fyne.Do(func() {
		window.SetContent(box)
	})
}
