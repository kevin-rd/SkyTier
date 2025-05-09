package packet

const (
	ProtocolVersion = 0x01
)

const (
	TypeData = 0x01

	// TypeAuxPeers request Peers list data
	TypeAuxPeers      = 0x02
	TypeAuxPeersReply = 0x03

	// TypeHandshake is handshake packet type
	TypeHandshake      = 0x04
	TypeHandshakeReply = 0x05

	// TypePing is ping packet type
	TypePing = 0x06
	TypePong = 0x07
)
