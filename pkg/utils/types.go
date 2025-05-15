package utils

import (
	"fmt"
	"net"
)

type IPv4 [4]byte

type IPMask [5]byte

// Must2IPMask parse 192.168.56.1/24 or 192.168.57.1 to IPMask, IP Mask bits default 24
func Must2IPMask(addr string) IPMask {
	ip, ipNet, err := net.ParseCIDR(addr)
	if err != nil {
		panic(err)
	}
	ip = ip.To4()
	if ip == nil {
		panic(fmt.Sprintf("invalid ip: %v", ip.String()))
	}
	_, bits := ipNet.Mask.Size()
	if bits <= 0 || bits > 32 {
		panic(fmt.Sprintf("invalid mask size: %d", bits))
	}
	return IPMask{ip[0], ip[1], ip[2], ip[3], byte(bits)}
}

func (ip IPMask) String() string {
	return fmt.Sprintf("%d.%d.%d.%d/%d", ip[0], ip[1], ip[2], ip[3], ip[4])
}

func (ip IPMask) toIP() net.IP {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3])
}
