//go:build android

package android

import (
	"path"
	"sync"
	"unsafe"

	"fyne.io/fyne/v2/driver"
)

// #include "intent.h"
import "C"

var initOnce sync.Once

func ensureInit() {
	initOnce.Do(func() {
		driver.RunNative(func(ctx interface{}) error {
			android := ctx.(*driver.AndroidContext)
			C.initJNI(
				C.uintptr_t(android.VM),
				C.uintptr_t(android.Env),
				C.uintptr_t(android.Ctx),
			)
			return nil
		})
	})
}

func GetIntent() Intent {
	ensureInit()

	var intent Intent

	driver.RunNative(func(ctx interface{}) error {
		android := ctx.(*driver.AndroidContext)

		output := C.getIntent(
			C.uintptr_t(android.VM),
			C.uintptr_t(android.Env),
			C.uintptr_t(android.Ctx),
		)

		intent.Action = Action(C.GoString(output.action))
		intent.Type = C.GoString(output._type)

		uris := unsafe.Slice(output.uri, output.uris)
		intent.URI = make([]string, len(uris))
		for i, uri := range uris {
			intent.URI[i] = C.GoString(uri)
		}

		intent.Text = C.GoString(output.text)

		return nil
	})

	return intent
}

func ReadContent(uri string) ([]byte, string) {
	ensureInit()

	var cData *C.uint8_t
	var cDataLength C.uint32_t

	var filename string

	driver.RunNative(func(ctx interface{}) error {
		android := ctx.(*driver.AndroidContext)

		cDisplayName := C.readContent(
			C.uintptr_t(android.VM),
			C.uintptr_t(android.Env),
			C.uintptr_t(android.Ctx),
			C.CString(uri),
			&cData, &cDataLength,
		)

		if cDisplayName != nil {
			filename = C.GoString(cDisplayName)
			C.free(unsafe.Pointer(cDisplayName))
		}

		return nil
	})

	data := make([]byte, cDataLength)
	copy(data, unsafe.Slice((*byte)(cData), cDataLength))
	C.free(unsafe.Pointer(cData))

	if filename == "" {
		filename = path.Base(uri) // not filepath because uri
	}

	return data, filename
}
