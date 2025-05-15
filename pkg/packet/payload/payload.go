package payload

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"kevin-rd/my-tier/pkg/utils"
	"net"
)

type DataPayload struct {
	Data []byte
}

// todo

type StringPayload string

func (s *StringPayload) Encode() ([]byte, error) {
	return []byte(*s), nil
}

func (s *StringPayload) Decode(data []byte) error {
	*s = StringPayload(data)
	return nil
}

func (s *StringPayload) Length() int {
	return len(*s)
}

type HandshakeInitPayload struct {
	ID [32]byte

	// CIDR e.g. 192.168.1.1/24
	VirtualIP utils.IPMask
	DHCP      bool
}

func (p *HandshakeInitPayload) Encode() ([]byte, error) {
	return p.MarshalBinary()
}

func (p *HandshakeInitPayload) Decode(data []byte) error {
	return p.UnmarshalBinary(data)
}

func (p *HandshakeInitPayload) Length() int {
	return 32 + 1 + 5
}

func (p *HandshakeInitPayload) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// ID
	if err := binary.Write(buf, binary.BigEndian, p.ID); err != nil {
		return nil, err
	}

	// is DHCP
	if err := binary.Write(buf, binary.BigEndian, p.DHCP); err != nil {
		return nil, err
	}

	// IP CIDR
	ip, ipNet, err := net.ParseCIDR(p.VirtualIP.String())
	if err != nil {
		return nil, fmt.Errorf("invalid VirtualIP: %w", err)
	}
	// IP must be IPv4
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("only IPv4 supported")
	}
	if _, err = buf.Write(ip); err != nil {
		return nil, err
	}
	maskSize, _ := ipNet.Mask.Size()
	if maskSize < 0 || maskSize > 32 {
		return nil, fmt.Errorf("invalid mask size: %d", maskSize)
	}
	if err = buf.WriteByte(byte(maskSize)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *HandshakeInitPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 32+1+5 {
		return fmt.Errorf("data too short: %d", len(data))
	}

	buf := bytes.NewReader(data)

	// Read ID
	if _, err := buf.Read(p.ID[:]); err != nil {
		return err
	}

	// Read DHCP flag
	if err := binary.Read(buf, binary.BigEndian, &p.DHCP); err != nil {
		return err
	}

	// Read IP and mask length
	if _, err := buf.Read(p.VirtualIP[:]); err != nil {
		return err
	}
	maskLen := int(p.VirtualIP[4])
	if maskLen < 0 || maskLen > 32 {
		return fmt.Errorf("invalid mask length: %d", maskLen)
	}

	return nil
}

type HandshakeReplyPayload struct {
	Hello     string
	VirtualIP utils.IPMask // client virtual IP
}

func (h *HandshakeReplyPayload) Encode() ([]byte, error) {
	return []byte(h.Hello), nil
}

func (h *HandshakeReplyPayload) Decode(data []byte) error {
	h.Hello = string(data)
	return nil
}

func (h *HandshakeReplyPayload) Length() int {
	return len(h.Hello)
}
