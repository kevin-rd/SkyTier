package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"kevin-rd/my-tier/pkg/packet/payload"
	"kevin-rd/my-tier/pkg/utils"
	"reflect"
)

type Packable interface {
	Encode() ([]byte, error)
	Decode([]byte) error

	Length() int
}

// Packet format:
// +------------+----------+---------+------------+
// | Version(4) | Flags(4) | Type(8) | Length(16) |
// +------------+----------+---------+------------+
// |               Src Virtual IP(32)             |
// +------------+----------+---------+------------+
// |               Dst Virtual IP(32)             |
// +------------+----------+---------+------------+
// |              Payload(variable)               |
// +------------+----------+---------+------------+
type Packet[T Packable] struct {
	Version byte
	Flags   byte
	Type    byte
	Length  uint16 // Payload Length
	SrcVIP  utils.IPv4
	DstVIP  utils.IPv4
	Payload T
}

func newPayload(typ byte) Packable {
	switch typ {
	case TypeHandshakeInit:
		return &payload.HandshakeInitPayload{}
	case TypeHandshakeReply:
		return &payload.HandshakeReplyPayload{}
	case TypeHandshakeFinalize:
		return new(payload.StringPayload)
	default:
		panic("unhandled default case")
	}
}

func NewPacket[T Packable](typ byte, payload T) *Packet[Packable] {
	return &Packet[Packable]{
		Version: ProtocolVersion,
		Type:    typ,
		Length:  uint16(payload.Length()),
		Payload: payload,
	}
}

// WithPayload constructs a new packet with the given generic payload.
func (p *Packet[T]) WithPayload(payload T) *Packet[T] {
	return &Packet[T]{
		Version: p.Version,
		Type:    p.Type,
		Flags:   p.Flags,
		Length:  uint16(payload.Length()),
		Payload: payload,
	}
}

// Encode converts Packet struct to raw bytes
func (p *Packet[T]) Encode() ([]byte, error) {
	if p.Payload.Length() > 0xFFFF-8 {
		return nil, ErrPacketTooLarge
	}

	buf := new(bytes.Buffer)

	// Version(4) + Flags(4)
	if err := buf.WriteByte(p.Version<<4 | p.Flags); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// Type(8)
	if err := buf.WriteByte(p.Type); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// Length(16)
	if err := binary.Write(buf, binary.BigEndian, uint16(p.Payload.Length())); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// SrcVIP and DstVIP
	if _, err := buf.Write(p.SrcVIP[:]); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	if _, err := buf.Write(p.DstVIP[:]); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}

	// Payload with packable.Encode()
	payloadBytes, err := p.Payload.Encode()
	if err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	if _, err = buf.Write(payloadBytes); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}

	return buf.Bytes(), nil
}

func (p *Packet[T]) Decode(data []byte) error {
	if len(data) < 4+4*2 {
		return errors.Join(ErrPacketDecode, ErrPacketTooSmall)
	}

	buf := bytes.NewReader(data)
	// read version
	var first byte
	if err := binary.Read(buf, binary.BigEndian, &first); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}
	p.Version = first >> 4
	p.Flags = first & 0x0F

	// read packet type
	if err := binary.Read(buf, binary.BigEndian, &p.Type); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	// read packet length
	if err := binary.Read(buf, binary.BigEndian, &p.Length); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	// read srcVIP and dstVIP
	if _, err := buf.Read(p.SrcVIP[:]); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}
	if _, err := buf.Read(p.DstVIP[:]); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	// check packet length
	if int(p.Length)+4+4*2 > len(data) {
		return errors.Join(ErrPacketDecode, ErrPacketIncomplete)
	}

	// read payload
	val := reflect.ValueOf(p.Payload)
	if val.Kind() == reflect.Invalid || val.IsNil() {
		p.Payload = newPayload(p.Type).(T)
	}
	if err := p.Payload.Decode(data[4+4*2 : 4+4*2+p.Length]); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	return nil
}
