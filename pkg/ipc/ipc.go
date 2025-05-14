package ipc

import "kevin-rd/my-tier/pkg/ipc/message"

type IPC interface {
	Send(message.Message) error
	Receive() (message.Message, error)
}
