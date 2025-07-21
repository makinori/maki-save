package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/forPelevin/gomoji"
	"github.com/rivo/uniseg"
)

func splitByEmoji(input string) []string {
	out := []string{}
	currentStr := ""

	gr := uniseg.NewGraphemes(input)

	for gr.Next() {
		str := gr.Str()

		emoji, err := gomoji.GetInfo(str)
		if err == nil {
			if currentStr != "" {
				out = append(out, currentStr)
				currentStr = ""
			}
			out = append(out, emoji.Character)
			continue
		}

		currentStr += str
	}

	if currentStr != "" {
		out = append(out, currentStr)
	}

	return out
}

func radioList(
	options []string, disableSelect binding.Bool,
	onSelect *func(int), onCancel func(),
) *fyne.Container {
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

	// create binding and && with disabled binding

	hasSelected := binding.NewBool()

	updateSelectButton := func() {
		a, _ := hasSelected.Get()
		if !a {
			selectButton.Disable()
			return
		}
		b, _ := disableSelect.Get()
		if b {
			selectButton.Disable()
			return
		}
		selectButton.Enable()
	}

	updateSelectButton()

	hasSelected.AddListener(binding.NewDataListener(updateSelectButton))
	disableSelect.AddListener(binding.NewDataListener(updateSelectButton))

	// fyne removes spaces after emojis in instance of: <emoji><space>text
	// so make multiple segments i guess
	optionsSegmented := make([][]string, len(options))
	for i, option := range options {
		// might be a heavy function cause of iteration
		optionsSegmented[i] = splitByEmoji(option)
	}

	listScroll := widget.NewList(
		func() int {
			return len(options)
		},
		func() fyne.CanvasObject {
			// return widget.NewLabel("")
			return widget.NewRichText()
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			// obj.(*widget.Label).SetText(options[i])

			text := obj.(*widget.RichText)
			// segment := text.Segments[0].(*widget.TextSegment)
			// segment.Text = options[i]

			text.Segments = make([]widget.RichTextSegment, len(optionsSegmented[i]))

			for i, segment := range optionsSegmented[i] {
				text.Segments[i] = &widget.TextSegment{
					Text: segment,
					Style: widget.RichTextStyle{
						Inline:   true,
						SizeName: theme.SizeNameSubHeadingText,
					},
				}
			}

			text.Refresh()
		},
	)

	listScroll.OnSelected = func(i widget.ListItemID) {
		selected = i
		hasSelected.Set(true)
	}

	// label := container.NewCenter(
	// 	widget.NewLabel("Select an album to upload to"),
	// )

	buttons := NewMinSize(
		fyne.Size{Height: selectButton.MinSize().Height * BUTTON_HEIGHT_MUL},
		container.NewBorder(nil, nil, cancelButton, nil, selectButton),
	)

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
