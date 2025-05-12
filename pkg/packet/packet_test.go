package packet

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -----------------------------
//        Unit Tests
// -----------------------------

// MockPackable is a mock implementation of the Packable interface for testing.
type MockPackable struct {
	encodeFunc func() ([]byte, error)
	decodeFunc func(data []byte) error
	lengthFunc func() int
}

func (m *MockPackable) Encode() ([]byte, error) {
	return m.encodeFunc()
}

func (m *MockPackable) Decode(data []byte) error {
	return m.decodeFunc(data)
}

func (m *MockPackable) Length() int {
	return m.lengthFunc()
}

// Constants for test
const (
	TestVersion = byte(1)
	TestType    = TypeHandshake
)

var (
	TestError = errors.New("test error")
)

func TestNewPacket(t *testing.T) {
	mock := &MockPackable{
		lengthFunc: func() int { return 10 },
	}

	pkt := NewPacket[Packable](TestType, mock)

	assert.Equal(t, TestVersion, pkt.Version) // shifted left by 4
	assert.Equal(t, TestType, pkt.Type)
	assert.Equal(t, uint16(10), pkt.Length)
	assert.NotNil(t, pkt.Payload)
}

func TestWithPayload(t *testing.T) {
	mock1 := &MockPackable{lengthFunc: func() int { return 10 }}
	mock2 := &MockPackable{lengthFunc: func() int { return 20 }}

	pkt1 := &Packet[Packable]{
		Version: TestVersion,
		Type:    TestType,
		Length:  10,
		Payload: mock1,
	}

	pkt2 := pkt1.WithPayload(mock2)

	assert.Equal(t, pkt1.Version, pkt2.Version)
	assert.Equal(t, pkt1.Type, pkt2.Type)
	assert.Equal(t, uint16(20), pkt2.Length)
	assert.Same(t, mock2, pkt2.Payload)
}

func TestEncode_Success(t *testing.T) {
	mock := &MockPackable{
		encodeFunc: func() ([]byte, error) { return []byte{0x01, 0x02}, nil },
		lengthFunc: func() int { return 2 },
	}

	pkt := &Packet[Packable]{
		Version: TestVersion,
		Type:    TestType,
		Length:  2,
		Payload: mock,
	}

	expected := []byte{
		TestVersion << 4, // Version(4),Reserved(4)
		TestType,         // Type(8)
		0x00, 0x02,       // Length(16)
		0x01, 0x02, // Payload
	}

	data, err := pkt.Encode()

	assert.NoError(t, err)
	assert.Equal(t, expected, data)
}

func TestEncode_PacketTooLarge(t *testing.T) {
	mock := &MockPackable{
		lengthFunc: func() int { return 0xFFFF - 7 }, // just over limit
	}

	pkt := &Packet[Packable]{
		Version: TestVersion << 4,
		Type:    TestType,
		Length:  uint16(mock.Length()),
		Payload: mock,
	}

	data, err := pkt.Encode()

	assert.ErrorIs(t, err, ErrPacketTooLarge)
	assert.Nil(t, data)
}

func TestDecode_Success(t *testing.T) {
	payloadBytes := []byte{0x01, 0x02}
	data := []byte{
		TestVersion << 4, // Version
		TestType,         // Type
		0x00, 0x02,       // Length
		0x01, 0x02, // Payload
	}

	mock := &MockPackable{
		decodeFunc: func(b []byte) error {
			assert.Equal(t, payloadBytes, b)
			return nil
		},
	}

	pkt := &Packet[Packable]{
		Payload: mock,
	}

	err := pkt.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, TestVersion, pkt.Version)
	assert.Equal(t, TestType, pkt.Type)
	assert.Equal(t, uint16(2), pkt.Length)
}

func TestDecode_TooSmall(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03} // less than 4 bytes

	pkt := &Packet[Packable]{}

	err := pkt.Decode(data)

	assert.ErrorIs(t, err, ErrPacketTooSmall)
}

func TestDecode_IncompleteData(t *testing.T) {
	data := []byte{
		TestVersion << 4, 0x00,
		TestType,
		0x00, 0x03, // Length=3
		0x01,
	} // total length < 4+3

	pkt := &Packet[Packable]{}

	err := pkt.Decode(data)

	assert.ErrorIs(t, err, ErrPacketIncomplete)
}

func TestDecode_UnknownType(t *testing.T) {
	data := []byte{
		TestVersion << 4,
		0xFF, // unknown type
		0x00, 0x00,
	}

	pkt := &Packet[Packable]{}

	err := pkt.Decode(data)

	assert.ErrorContains(t, err, "unknown packet type")
}

func TestDecode_PayloadDecodeError(t *testing.T) {
	data := []byte{
		TestVersion << 4,
		TestType,
		0x00, 0x02,
		0x00, 0x00,
	}

	mock := &MockPackable{
		decodeFunc: func(b []byte) error {
			return TestError
		},
	}

	pkt := &Packet[Packable]{
		Payload: mock,
	}

	err := pkt.Decode(data)

	assert.ErrorIs(t, err, TestError)
}

func TestDecodePayload_Success(t *testing.T) {
	data := []byte{ //0x10, 0x04, 0x00, 0x00,
		'h', 'e', 'l', 'l', 'o', 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 192, 168, 22, 1, 24,
	}

	pl, err := decodePayload[Packable](TypeHandshake, data)

	assert.NoError(t, err)
	assert.NotNil(t, pl)
}

func TestDecodePayload_UnknownType(t *testing.T) {
	_, err := decodePayload[Packable](0xFF, nil)

	assert.ErrorContains(t, err, "unknown packet type")
}
