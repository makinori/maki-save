package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type minSize struct {
	size fyne.Size
}

func (f *minSize) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := objects[0].MinSize()
	if f.size.Width > 0 {
		minSize.Width = f.size.Width
	}
	if f.size.Height > 0 {
		minSize.Height = f.size.Height
	}
	return minSize
}

func (f *minSize) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	objects[0].Resize(containerSize)
}

func NewMinSize(size fyne.Size, object fyne.CanvasObject) *fyne.Container {
	return container.New(&minSize{
		size: size,
	}, object)
}

/*
type maxSize struct {
	size fyne.Size
}

func (f *maxSize) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[0].MinSize()
}

func (f *maxSize) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	size := objects[0].MinSize()
	if f.size.Width > 0 {
		size.Width = f.size.Width
	}
	if f.size.Height > 0 {
		size.Height = f.size.Height
	}
	objects[0].Resize(size)
}

func NewMaxSize(size fyne.Size, object fyne.CanvasObject) *fyne.Container {
	return container.New(&maxSize{
		size: size,
	}, object)
}
*/
