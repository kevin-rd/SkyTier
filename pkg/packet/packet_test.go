package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// -----------------------------
//        Unit Tests
// -----------------------------

func TestEncode_Normal(t *testing.T) {
	p := &Packet{
		Version: 0x0A,
		Type:    0x01,
		Payload: []byte{0x01, 0x02, 0x03},
	}

	data, err := p.Encode()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xA0, 0x01, 0x00, 0x03, 0x01, 0x02, 0x03}, data)
}

func TestEncode_PayloadTooLarge(t *testing.T) {
	p := &Packet{
		Version: 0x0F,
		Type:    0xFF,
		Payload: make([]byte, 0x10000),
	}

	_, err := p.Encode()
	assert.ErrorIs(t, err, ErrPacketTooLarge)
}

func TestEncode_EmptyPayload(t *testing.T) {
	p := &Packet{
		Version: 0x05,
		Type:    0x00,
		Payload: []byte{},
	}

	data, err := p.Encode()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x50, 0x00, 0x00, 0x00}, data)
}

func TestDecode_Normal(t *testing.T) {
	data := []byte{0xA0, 0x01, 0x00, 0x03, 0x01, 0x02, 0x03}
	packet, err := Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, byte(0x0A), packet.Version)
	assert.Equal(t, byte(0x01), packet.Type)
	assert.Equal(t, uint16(3), packet.Length)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, packet.Payload)
}

func TestDecode_TooSmall(t *testing.T) {
	data := []byte{0xA0, 0x01}
	packet, err := Decode(data)

	assert.ErrorIs(t, err, ErrPacketTooSmall)
	assert.Nil(t, packet)
}

func TestDecode_Incomplete(t *testing.T) {
	data := []byte{0xA0, 0x01, 0x00, 0x05, 0x01, 0x02} // 长度要求5字节payload，实际只有2
	packet, err := Decode(data)

	assert.ErrorIs(t, err, ErrPacketIncomplete)
	assert.Nil(t, packet)
}

func TestDecode_EmptyPayload(t *testing.T) {
	data := []byte{0x50, 0x00, 0x00, 0x00}
	packet, err := Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, byte(0x05), packet.Version)
	assert.Equal(t, byte(0x00), packet.Type)
	assert.Equal(t, uint16(0), packet.Length)
	assert.Empty(t, packet.Payload)
}
