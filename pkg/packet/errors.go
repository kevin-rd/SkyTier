package packet

import (
	"errors"
)

var (
	ErrPacketTooLarge = errors.New("packet too large")
	ErrPacketTooSmall = errors.New("packet too small")

	ErrPacketEncode     = errors.New("packet encode error")
	ErrPacketDecode     = errors.New("packet decode error")
	ErrPacketIncomplete = errors.New("packet incomplete")
)
