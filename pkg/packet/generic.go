package packet

type Packetable interface {
}

type PacketG[T Packetable] struct {
	Version byte
	Type    byte
	Length  uint16
	Payload T
}
