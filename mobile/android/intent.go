package android

import (
	"sync"
	"unsafe"

	"fyne.io/fyne/v2/driver"
)

// #include "intent.h"
import "C"

type Intent struct {
	Action string
	Type   string
	URI    string
}

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

		intent.Action = C.GoString(output.action)
		intent.Type = C.GoString(output._type)
		intent.URI = C.GoString(output.uri)

		return nil
	})

	return intent
}

func ReadContent(uri string) []byte {
	ensureInit()

	var cData *C.uint8_t
	var cDataLength C.uint32_t

	driver.RunNative(func(ctx interface{}) error {
		android := ctx.(*driver.AndroidContext)

		C.readContent(
			C.uintptr_t(android.VM),
			C.uintptr_t(android.Env),
			C.uintptr_t(android.Ctx),
			C.CString(uri), &cData, &cDataLength,
		)

		return nil
	})

	data := make([]byte, cDataLength)
	copy(data, unsafe.Slice((*byte)(cData), cDataLength))
	C.free(unsafe.Pointer(cData))

	return data
}
