package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func radioList(options []string, onSelect *func(int), onCancel func()) *fyne.Container {
	var selected int

	// w.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
	// 	switch event.Name {
	// 	case "Escape":
	// 		selected = -1
	// 		w.Close()
	// 	case "Return":
	// 		w.Close()
	// 	}
	// })

	cancelButton := widget.NewButtonWithIcon(
		"Cancel", theme.CancelIcon(),
		func() {
			selected = -1
			onCancel()
			// w.Close()
		},
	)

	okButton := widget.NewButtonWithIcon(
		"Ok", theme.ConfirmIcon(),
		func() {
			if onSelect != nil {
				(*onSelect)(selected)
			}
			// w.Close()
		},
	)

	okButton.Disable()

	list := widget.NewList(
		func() int { return len(options) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(options[i])
		},
	)

	list.OnSelected = func(i widget.ListItemID) {
		selected = i
		okButton.Enable()
	}

	listScroll := container.NewVScroll(list)

	// label := container.NewCenter(
	// 	widget.NewLabel("Select an album to upload to"),
	// )

	buttons := container.NewBorder(nil, nil, nil, container.NewHBox(
		cancelButton,
		okButton,
	))

	box := container.NewBorder(
		// label,
		nil,
		buttons,
		nil,
		nil,
		listScroll,
	)

	return box

	// w.SetContent(box)

	// w.ShowAndRun()

	// return selected
}
