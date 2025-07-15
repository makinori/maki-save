package main

import (
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ScreenError struct {
	// title,text pairs.
	// prefix title with # for no error.
	Text           []string
	NoDismiss      bool
	NoSelfDestruct bool
}

type ScreenTextOption uint8

const (
	ScreenTextOptionNoError ScreenTextOption = iota
	ScreenTextOptionNoDismiss
)

func showScreenError(in ScreenError) {

	if !in.NoSelfDestruct {
		// errors self destruct
		go func() {
			time.Sleep(time.Second * 10)
			os.Exit(0)
		}()
	}

	var segments []widget.RichTextSegment

	for i, str := range in.Text {
		if str == "" {
			continue
		}

		if i%2 == 0 {
			// title
			newStr := strings.TrimPrefix(str, "#")
			var titleColorName fyne.ThemeColorName = theme.ColorNameError
			if str != newStr {
				titleColorName = ""
			}
			segments = append(segments, &widget.TextSegment{
				Text: strings.TrimSpace(newStr),
				Style: widget.RichTextStyle{
					Alignment: fyne.TextAlignCenter,
					SizeName:  theme.SizeNameSubHeadingText,
					ColorName: titleColorName,
					TextStyle: fyne.TextStyle{Bold: true},
				},
			})
		} else {
			// text
			segments = append(segments, &widget.TextSegment{
				Text: strings.TrimSpace(str),
				Style: widget.RichTextStyle{
					Alignment: fyne.TextAlignCenter,
				},
			})
		}

		// spacer
		if i < len(in.Text)-1 {
			segments = append(segments, &widget.TextSegment{
				Style: widget.RichTextStyle{
					SizeName: theme.SizeNamePadding,
				},
			})
		}
	}

	if !in.NoDismiss {
		// smaller spacer
		segments = append(segments, &widget.TextSegment{
			Style: widget.RichTextStyle{
				SizeName: theme.SizeNameSeparatorThickness,
			},
		})
	}

	var content = []fyne.CanvasObject{
		layout.NewSpacer(),
		widget.NewRichText(segments...),
	}

	if !in.NoDismiss {
		dismissButton := widget.NewButton("dismiss", func() { os.Exit(0) })
		content = append(content, container.NewCenter(dismissButton))
	}

	content = append(content,
		layout.NewSpacer(),
	)

	fyne.DoAndWait(func() {
		window.SetContent(container.NewVBox(content...))
	})
}
