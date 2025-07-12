package android

import "C"

const (
	ACTION_SEND = "android.intent.action.SEND"
)

type Intent struct {
	Action string
	Type   string
	URI    string
	Text   string
}
