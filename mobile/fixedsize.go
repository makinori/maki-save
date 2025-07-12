package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type fixedSize struct {
	size fyne.Size
}

func (f *fixedSize) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := objects[0].MinSize()
	if f.size.Width > 0 {
		minSize.Width = f.size.Width
	}
	if f.size.Height > 0 {
		minSize.Height = f.size.Height
	}
	return minSize
}

func (f *fixedSize) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	objects[0].Resize(containerSize)
}

func NewFixedSize(size fyne.Size, object fyne.CanvasObject) *fyne.Container {
	return container.New(&fixedSize{
		size: size,
		// layout: layout.New(),
	}, object)
}
