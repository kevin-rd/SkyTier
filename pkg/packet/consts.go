package packet

import "errors"

// ProtocolVersion is Packet Protocol Version
const (
	ProtocolVersion = 0x01
)

// Packet Type
const (
	TypeData byte = iota

	// TypeAuxPeers request Peers list data
	TypeAuxPeers
	TypeAuxPeersReply

	// TypeHandshake is handshake packet type
	TypeHandshake
	TypeHandshakeReply

	// TypePing is ping packet type
	TypePing
	TypePong

	TypeCmdRequest
	TypeCmdReply
)

// Packet errors
var (
	ErrPacketTooLarge = errors.New("packet too large")
	ErrPacketTooSmall = errors.New("packet too small")

	ErrPacketEncode     = errors.New("packet encode error")
	ErrPacketDecode     = errors.New("packet decode error")
	ErrPacketIncomplete = errors.New("packet incomplete")
)
