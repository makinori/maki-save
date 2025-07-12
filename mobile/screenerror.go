package main

import (
	"os"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ScreenTextOption uint8

const (
	ScreenTextOptionNoError ScreenTextOption = iota
	ScreenTextOptionNoSelfDestruct
	ScreenTextOptionNoDismiss
)

func showScreenError(
	title string, text string, options ...ScreenTextOption,
) {
	var titleColorName fyne.ThemeColorName = theme.ColorNameError

	if slices.Contains(options, ScreenTextOptionNoError) {
		titleColorName = ""
	}

	if !slices.Contains(options, ScreenTextOptionNoSelfDestruct) {
		go func() {
			time.Sleep(time.Second * 10)
			os.Exit(0)
		}()
	}

	var content = []fyne.CanvasObject{
		layout.NewSpacer(),
	}

	if title != "" {
		content = append(content, widget.NewRichText(
			&widget.TextSegment{
				Text: title,
				Style: widget.RichTextStyle{
					Alignment: fyne.TextAlignCenter,
					SizeName:  theme.SizeNameSubHeadingText,
					ColorName: titleColorName,
					TextStyle: fyne.TextStyle{
						Bold: true,
					},
				},
			},
		))
	}

	textLabel := widget.NewLabel(text)
	textLabel.Truncation = fyne.TextTruncateEllipsis
	textLabel.Alignment = fyne.TextAlignCenter

	dismissButton := widget.NewButton("dismiss", func() {
		os.Exit(0)
	})

	paddedContent := []fyne.CanvasObject{
		textLabel,
	}

	if !slices.Contains(options, ScreenTextOptionNoDismiss) {
		paddedContent = append(paddedContent, container.NewCenter(dismissButton))
	}

	content = append(content,
		container.New(layout.NewCustomPaddedVBoxLayout(16),
			paddedContent...,
		),
		layout.NewSpacer(),
	)

	fyne.DoAndWait(func() {
		window.SetContent(container.NewVBox(content...))
	})
}
