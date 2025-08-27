package android

import "C"

type Action string

const (
	ACTION_SEND          Action = "android.intent.action.SEND"
	ACTION_SEND_MULTIPLE Action = "android.intent.action.SEND_MULTIPLE"
	ACTION_SENDTO        Action = "android.intent.action.SENDTO"
)

var Actions = []Action{
	ACTION_SEND, ACTION_SEND_MULTIPLE, ACTION_SENDTO,
}

type Intent struct {
	Action Action
	Type   string
	URI    []string
	Text   string
}
