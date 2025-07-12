package android

import "C"

const (
	ACTION_SEND          = "android.intent.action.SEND"
	ACTION_SEND_MULTIPLE = "android.intent.action.SEND_MULTIPLE"
)

type Intent struct {
	Action string
	Type   string
	URI    []string
	Text   string
}
