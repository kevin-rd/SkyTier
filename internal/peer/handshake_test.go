package peer

import (
	"bytes"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 自定义 Buffer 用于模拟写入失败
type faultBuffer struct {
	bytes.Buffer
	failAt int
	count  int
}

func (b *faultBuffer) Write(p []byte) (n int, err error) {
	if b.count >= b.failAt {
		return 0, errors.New("write error")
	}
	n, err = b.Buffer.Write(p)
	b.count += n
	return n, err
}

func TestPayloadHandshake_MarshalBinary(t *testing.T) {
	tests := []struct {
		name        string
		payload     PayloadHandshake
		expectError bool
		errorMsg    string
		failAt      int // -1 表示不失败
	}{
		{
			name: "Valid IPv4",
			payload: PayloadHandshake{
				ID:     [32]byte{'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'},
				DHCP:   true,
				IpCidr: "192.168.1.0/24",
			},
			expectError: false,
		},
		{
			name: "Invalid CIDR",
			payload: PayloadHandshake{
				IpCidr: "invalid_cidr",
			},
			expectError: true,
			errorMsg:    "packet encode error",
		},
		{
			name: "IPv6 Not Supported",
			payload: PayloadHandshake{
				IpCidr: "2001:db8::/32",
			},
			expectError: true,
			errorMsg:    "only IPv4 supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &tt.payload

			data, err := p.MarshalBinary()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
			}
		})
	}
}

func TestPayloadHandshake_UnmarshalBinary(t *testing.T) {
	validData := func() []byte {
		id := [32]byte{'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'}
		dhcp := byte(1)
		ip := net.ParseIP("192.168.1.1").To4()
		mask := byte(24)

		buf := new(bytes.Buffer)
		buf.Write(id[:])
		buf.WriteByte(dhcp)
		buf.Write(ip)
		buf.WriteByte(mask)
		return buf.Bytes()
	}()

	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
		expected    PayloadHandshake
	}{
		{
			name:        "Valid Data",
			data:        validData,
			expectError: false,
			expected: PayloadHandshake{
				ID:     [32]byte{'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'},
				DHCP:   true,
				IpCidr: "192.168.1.1/24",
			},
		},
		{
			name:        "Data Too Small",
			data:        []byte{1, 2, 3},
			expectError: true,
			errorMsg:    "packet decode error",
		},
		{
			name: "Invalid Mask Length",
			data: func() []byte {
				base := append([]byte(nil), validData...)
				base[len(base)-1] = 33 // mask length > 32
				return base
			}(),
			expectError: true,
			errorMsg:    "invalid mask length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p PayloadHandshake
			err := p.UnmarshalBinary(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, p.ID)
				assert.Equal(t, tt.expected.DHCP, p.DHCP)
				assert.Equal(t, tt.expected.IpCidr, p.IpCidr)
			}
		})
	}
}
