package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

// Packet format
// +------------+-------------+---------+------------+--------------------+
// | Version(4) | Reserved(4) | Type(8) | Length(16) | Payload (variable) |
// +------------+-------------+---------+------------+--------------------+
type Packet[T Packable] struct {
	Version byte
	Type    byte
	Length  uint16
	Payload T
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

	// Version(4) + Reserved(4)
	if err := buf.WriteByte(p.Version<<4 | 0x00); err != nil {
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
	if len(data) < 4 {
		return errors.Join(ErrPacketDecode, ErrPacketTooSmall)
	}

	buf := bytes.NewReader(data)
	// read version
	if err := binary.Read(buf, binary.BigEndian, &p.Version); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}
	p.Version = p.Version >> 4

	// read packet type
	if err := binary.Read(buf, binary.BigEndian, &p.Type); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	// read packet length
	if err := binary.Read(buf, binary.BigEndian, &p.Length); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	if int(p.Length)+4 > len(data) {
		return errors.Join(ErrPacketDecode, ErrPacketIncomplete)
	}

	// read payload
	val := reflect.ValueOf(p.Payload)
	if val.Kind() == reflect.Invalid || val.IsNil() {
		payload, err := newPayload(p.Type)
		if err != nil {
			return errors.Join(ErrPacketDecode, err)
		}
		p.Payload = any(payload).(T)
	}
	if err := p.Payload.Decode(data[4 : 4+p.Length]); err != nil {
		return errors.Join(ErrPacketDecode, err)
	}

	return nil
}
