package peer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"kevin-rd/my-tier/pkg/packet"
	"net"
)

type PayloadHandshake struct {
	ID [32]byte

	// Deprecated: DHCP 暂时不支持
	DHCP bool

	IpCidr string // e.g. 192.168.1.1/24
}

func (p *PayloadHandshake) Encode() ([]byte, error) {
	return p.MarshalBinary()
}

func Decode(data []byte) (*PayloadHandshake, error) {
	var payload PayloadHandshake
	err := payload.UnmarshalBinary(data)
	return &payload, err
}

func (p *PayloadHandshake) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// ID
	if err := binary.Write(buf, binary.BigEndian, p.ID); err != nil {
		return nil, errors.Join(packet.ErrPacketEncode, err)
	}

	// is DHCP
	if err := binary.Write(buf, binary.BigEndian, p.DHCP); err != nil {
		return nil, errors.Join(packet.ErrPacketEncode, err)
	}

	// IP CIDR
	ip, ipNet, err := net.ParseCIDR(p.IpCidr)
	if err != nil {
		return nil, errors.Join(packet.ErrPacketEncode, err)
	}
	// IP must be IPv4
	ip = ip.To4()
	if ip == nil {
		return nil, errors.Join(packet.ErrPacketEncode, fmt.Errorf("only IPv4 supported"))
	}
	if _, err := buf.Write(ip); err != nil {
		return nil, errors.Join(packet.ErrPacketEncode, err)
	}
	maskSize, _ := ipNet.Mask.Size()
	if maskSize < 0 || maskSize > 255 {
		return nil, fmt.Errorf("invalid mask size: %d: %w", maskSize, packet.ErrPacketEncode)
	}
	if err := buf.WriteByte(byte(maskSize)); err != nil {
		return nil, errors.Join(packet.ErrPacketEncode, err)
	}

	return buf.Bytes(), nil
}

func (p *PayloadHandshake) UnmarshalBinary(data []byte) error {
	if len(data) < 32+1+5 {
		return errors.Join(packet.ErrPacketDecode, packet.ErrPacketTooSmall)
	}

	buf := bytes.NewReader(data)

	// Read ID
	if _, err := buf.Read(p.ID[:]); err != nil {
		return errors.Join(packet.ErrPacketDecode, err)
	}

	// Read DHCP flag
	if err := binary.Read(buf, binary.BigEndian, &p.DHCP); err != nil {
		return errors.Join(packet.ErrPacketDecode, err)
	}

	// Read IP and mask length
	ipBytes := make([]byte, 4)
	if _, err := buf.Read(ipBytes); err != nil {
		return errors.Join(packet.ErrPacketDecode, err)
	}
	ip := net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
	maskLenByte, err := buf.ReadByte()
	if err != nil {
		return errors.Join(packet.ErrPacketDecode, err)
	}
	maskLen := int(maskLenByte)
	if maskLen < 0 || maskLen > 32 {
		return errors.Join(packet.ErrPacketEncode, fmt.Errorf("invalid mask length: %d", maskLen))
	}
	ipNet := net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(maskLen, 32),
	}
	p.IpCidr = ipNet.String()

	return nil
}
