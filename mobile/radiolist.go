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

	selectButton := widget.NewButtonWithIcon(
		"Select", theme.ConfirmIcon(),
		func() {
			if onSelect != nil {
				(*onSelect)(selected)
			}
			// w.Close()
		},
	)

	selectButton.Disable()

	list := widget.NewList(
		func() int { return len(options) },
		func() fyne.CanvasObject {
			// return widget.NewLabel("")
			return widget.NewRichText(&widget.TextSegment{
				Text: "",
				Style: widget.RichTextStyle{
					SizeName: theme.SizeNameSubHeadingText,
				},
			})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			// o.(*widget.Label).SetText(options[i])
			seg := o.(*widget.RichText).Segments[0]
			seg.(*widget.TextSegment).Text = options[i]
		},
	)

	list.OnSelected = func(i widget.ListItemID) {
		selected = i
		selectButton.Enable()
	}

	listScroll := container.NewVScroll(list)

	// label := container.NewCenter(
	// 	widget.NewLabel("Select an album to upload to"),
	// )

	buttons := container.NewBorder(nil, nil, cancelButton, nil, selectButton)

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
