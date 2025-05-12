package packet

const (
	ProtocolVersion = 0x01
)

const (
	TypeData byte = 0x01

	// TypeAuxPeers request Peers list data
	TypeAuxPeers      byte = 0x02
	TypeAuxPeersReply byte = 0x03

	// TypeHandshake is handshake packet type
	TypeHandshake      byte = 0x04
	TypeHandshakeReply byte = 0x05

	// TypePing is ping packet type
	TypePing byte = 0x06
	TypePong byte = 0x07
)
