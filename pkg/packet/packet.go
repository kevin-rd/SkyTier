package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Packet format
// +------------+------------+----------+--------------------+
// | Version(1) | Type(1)    | Length(2)| Payload (variable) |
// +------------+------------+----------+--------------------+
type Packet struct {
	Version byte
	Type    byte
	Length  uint16
	Payload []byte
}

// Encode converts Packet struct to raw bytes
func (p *Packet) Encode() ([]byte, error) {
	if len(p.Payload) > 0xFFFF {
		return nil, ErrPacketTooLarge
	}

	buf := new(bytes.Buffer)

	// Version 4 bit
	if err := buf.WriteByte(p.Version<<4 | 0x00); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// Type 8 bit
	if err := buf.WriteByte(p.Type); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// Length 16 bit
	if err := binary.Write(buf, binary.BigEndian, uint16(len(p.Payload))); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}
	// Payload
	if _, err := buf.Write(p.Payload); err != nil {
		return nil, errors.Join(ErrPacketEncode, err)
	}

	return buf.Bytes(), nil
}

// Deprecated: Decode parses raw bytes to Packet struct
func Decode(data []byte) (*Packet, error) {
	if len(data) < 4 {
		return nil, ErrPacketTooSmall
	}

	version := data[0] >> 4
	typ := data[1]
	length := binary.BigEndian.Uint16(data[2:4])

	if int(length)+4 > len(data) {
		return nil, ErrPacketIncomplete
	}

	payload := data[4 : 4+length]

	return &Packet{
		Version: version,
		Type:    typ,
		Length:  length,
		Payload: payload,
	}, nil
}

func ReadPacket(r io.Reader) (*Packet, error) {
	var p Packet

	if err := binary.Read(r, binary.BigEndian, &p.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &p.Type); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &p.Length); err != nil {
		return nil, err
	}

	p.Payload = make([]byte, p.Length)
	if _, err := r.Read(p.Payload); err != nil {
		return nil, err
	}

	return &p, nil
}
