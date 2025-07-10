package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func showDefaultScreen() {
	header := widget.NewLabel("maki immich")
	header.TextStyle.Bold = true

	text := widget.NewLabel("share an image to this app")

	content := container.NewCenter(container.NewVBox(
		container.NewCenter(header),
		container.NewCenter(text),
	))

	window.SetContent(content)

}
